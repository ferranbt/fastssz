package generator

import (
	"fmt"
	"strconv"
	"strings"
)

// unmarshal creates a function that decodes the structs with the input byte in SSZ format.
func (e *env) unmarshal(name string, v *Value) string {
	tmpl := `// UnmarshalSSZ ssz unmarshals the {{.name}} object
	func (:: *{{.name}}) UnmarshalSSZ(buf []byte) error {
		return ssz.UnmarshalSSZ(::, buf)
	}
	
	// UnmarshalSSZTail unmarshals the {{.name}} object and returns the remaining bufferº
	func (:: *{{.name}}) UnmarshalSSZTail(buf []byte) (rest []byte, err error) {
		{{.unmarshal}}
	}`

	str := execTmpl(tmpl, map[string]interface{}{
		"name":      name,
		"unmarshal": v.umarshalContainer(true, "buf"),
	})

	return appendObjSignature(str, v)
}

func (v *Value) unmarshal(dst string) string {
	switch obj := v.typ.(type) {
	case *Container, *Reference:
		return v.umarshalContainer(false, dst)

	case *Bytes:
		if !obj.IsList && !obj.IsGoDyn {
			return fmt.Sprintf("buf = ssz.UnmarshalFixedBytes(::.%s[:], buf)", v.name)
		}

		// both fixed and dynamic are decoded equally
		var tmpl string
		if !v.isFixed() {
			// dynamic bytes, we need to validate the size of the buffer
			tmpl = `if ::.{{.name}}, err = ssz.UnmarshalDynamicBytes(::.{{.name}}, {{.dst}}, {{.size}}); err != nil {
			return
			}`
		} else {
			tmpl = `::.{{.name}}, buf = ssz.UnmarshalBytes(::.{{.name}}, buf, {{.size}})`
		}

		return execTmpl(tmpl, map[string]interface{}{
			"name":  v.name,
			"dst":   dst,
			"size":  obj.Size,
			"isRef": v.ref != "",
			"obj":   v,
		})

	case *BitList:
		// This is always a dynamic element type so we do not need to consume buffer
		tmpl := `if ::.{{.name}}, err = ssz.UnmarshalBitList(::.{{.name}}, {{.dst}}, {{.size}}); err != nil {
			return nil, err
		}`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"dst":  dst,
			"size": obj.Size,
		})

	case *Uint:
		intType := uintVToLowerCaseName2(obj)

		var objRef string
		if v.ref != "" {
			objRef = v.objRef()
		} else if v.obj != "" {
			objRef = v.obj
		}

		var tmpl string
		if objRef != "" {
			tmpl = `{
				var val {{.type}}
				val, buf = ssz.UnmarshallValue[{{.type}}](buf)
				::.{{.name}} = {{.objRef}}(val)
			}`
		} else {
			tmpl = `::.{{.name}}, buf = ssz.UnmarshallValue[{{.type}}](buf)`
		}

		return execTmpl(tmpl, map[string]interface{}{
			"name":   v.name,
			"type":   intType,
			"objRef": objRef,
		})

		/*
			if v.ref != "" {
				// alias, we need to cast the value
				return fmt.Sprintf("::.%s, buf = %s(ssz.UnmarshallValue[%s](buf))", v.name, v.objRef(), intType)
			}
			if v.obj != "" {
				// alias to a type on the same package
				return fmt.Sprintf("::.%s, buf = %s(ssz.UnmarshallValue[%s](buf))", v.name, v.obj, intType)
			}
			return fmt.Sprintf("::.%s, buf = ssz.UnmarshallValue[%s](buf)", v.name, intType)
		*/

	case *Bool:
		return fmt.Sprintf("::.%s, buf = ssz.UnmarshallValue[bool](buf)", v.name)

	case *Time:
		return fmt.Sprintf("::.%s, buf = ssz.UnmarshalTime(buf)", v.name)

	case *Vector:
		if obj.Elem.isFixed() {
			tmpl := `{{.create}}
			for ii := 0; ii < {{.size}}; ii++ {
				{{.unmarshal}}
			}`
			return execTmpl(tmpl, map[string]interface{}{
				"create":    v.createSlice(false),
				"size":      obj.Size,
				"unmarshal": obj.Elem.unmarshal("buf"),
			})
		} else {
			return v.unmarshalList(dst)
		}

	case *List:
		return v.unmarshalList(dst)

	default:
		panic(fmt.Errorf("unmarshal not implemented for type %s", v.Type()))
	}
}

func (v *Value) unmarshalList(dst string) string {
	var size Size
	if obj, ok := v.typ.(*List); ok {
		size = obj.MaxSize
	} else if obj, ok := v.typ.(*Vector); ok {
		size = obj.Size
	} else {
		panic(fmt.Errorf("unmarshalList not implemented for type %s", v.Type()))
	}

	inner := getElem(v.typ)
	if inner.isFixed() {
		var tmpl string
		var innerSize string

		if inner.isContainer() && !inner.noPtr {
			tmpl = `if err = ssz.UnmarshalSliceSSZ(&::.{{.name}}, {{.dst}}, {{.max}}); err != nil {
			return nil, err
		}`
		} else {
			// it is a basic type, manually infer the size
			switch obj := inner.typ.(type) {
			case *Uint:
				innerSize = fmt.Sprintf("%d", obj.Size)
			case *Bytes:
				innerSize = fmt.Sprintf("%d", obj.Size)
			case *Container:
				innerSize = inner.fixedSizeForContainer()
			default:
				// TODO: It is my impression that maybe calling inner.fixedSize() would work for all the cases
				panic(fmt.Errorf("unmarshalList not implemented for type %s", inner.Type()))
			}

			tmpl = `if err = ssz.UnmarshalSliceWithIndexCallback(&::.{{.name}}, {{.dst}}, {{.size}}, {{.max}}, func(ii int, buf []byte) (err error) {
			{{.unmarshal}}
			return nil
		}); err != nil {
			return nil, err
		}`
		}
		return execTmpl(tmpl, map[string]interface{}{
			"size":      innerSize,
			"max":       size,
			"name":      v.name,
			"unmarshal": inner.unmarshal("buf"),
			"dst":       dst,
		})
	}

	// Decode list with a dynamic element. 'ssz.DecodeDynamicLength' ensures
	// that the number of elements do not surpass the 'ssz-max' tag.

	var tmpl string

	if inner.isContainer() && !inner.noPtr {
		tmpl = `if err = ssz.UnmarshalDynamicSliceSSZ(&::.{{.name}}, {{.dst}}, {{.max}}); err != nil {
			return nil, err
		}`
	} else {
		tmpl = `if err = ssz.UnmarshalDynamicSliceWithCallback(&::.{{.name}}, {{.dst}}, {{.max}}, func(indx int, buf []byte) (err error) {
		{{.unmarshal}}
		return nil
	}); err != nil {
		return nil, err
	}`
	}

	inner.name = v.name + "[indx]"

	data := map[string]interface{}{
		"max":       size,
		"name":      v.name,
		"create":    v.createSlice(true),
		"unmarshal": inner.unmarshal("buf"),
		"dst":       dst,
	}
	return execTmpl(tmpl, data)
}

func isInOffset(dst string) bool {
	return strings.HasPrefix(dst, "tail[")
}

func (v *Value) umarshalContainer(start bool, dst string) (str string) {
	if !start {
		var tmpl string
		if isInOffset(dst) {
			tmpl = `{{if .ptr}}if err = ssz.UnmarshalField(&::.{{.name}}, {{.dst}}); err != nil {
			return
		}{{else}}if err = ::.{{.name}}.UnmarshalSSZ(buf); err != nil {
			return
		}{{end}}`
		} else {
			tmpl = `{{if .ptr}}if buf, err = ssz.UnmarshalFieldTail(&::.{{.name}}, buf); err != nil {
			return
		}{{else}}if buf, err = ::.{{.name}}.UnmarshalSSZTail(buf); err != nil {
			return
		}{{end}}`
		}
		check := true
		if v.noPtr {
			check = false
		}
		return execTmpl(tmpl, map[string]interface{}{
			"name":  v.name,
			"obj":   v,
			"dst":   dst,
			"check": check,
			"ptr":   !v.noPtr,
		})
	}

	var offsets []string
	offsetsMatch := map[string]string{}

	for indx, i := range v.getObjs() {
		if !i.isFixed() {
			name := "o" + strconv.Itoa(indx)
			if len(offsets) != 0 {
				offsetsMatch[name] = offsets[len(offsets)-1]
			}
			offsets = append(offsets, name)
		}
	}

	// safe check for the size. Two cases:
	// 1. Struct is fixed: The size of the input buffer must be the same as the struct.
	// 2. Struct is dynamic. The size of the input buffer must be higher than the fixed part of the struct.

	var cmp string
	if v.isFixed() {
		cmp = "!="
	} else {
		cmp = "<"
	}

	// If the struct is dynamic we create a set of offset variables that will be readed later.

	tmpl := `size := len(buf)
	fixedSize := ::.fixedSize()
	if size < fixedSize {
		return nil, ssz.ErrSize
	}
	{{if .offsets}}
		tail := buf
		var {{.offsets}} uint64
		marker := ssz.NewOffsetMarker(uint64(size), uint64(fixedSize))
	{{end}}
	`

	str += execTmpl(tmpl, map[string]interface{}{
		"cmp":     cmp,
		"offsets": strings.Join(offsets, ", "),
	})

	//var o0 uint64

	// Marshal the fixed part and offsets

	// used for bounds checking of variable length offsets.
	// for the first offset, use the size of the fixed-length data
	// as the minimum boundary. subsequent offsets will replace this
	// value with the name of the previous offset variable.
	outs := []string{}
	for indx, i := range v.getObjs() {
		var res string
		if i.isFixed() {
			res = fmt.Sprintf("// Field (%d) '%s'\n%s\n\n", indx, i.name, i.unmarshal("buf"))

		} else {
			// read the offset
			offset := "o" + strconv.Itoa(indx)

			data := map[string]interface{}{
				"indx":   indx,
				"name":   i.name,
				"offset": offset,
				"dst":    dst,
			}

			// We need to do two validations for the offset:
			// 1. The offset is lower than the total size of the input buffer
			// 2. The offset i needs to be higher than the offset i-1 (Only if the offset is not the first).

			if prev, ok := offsetsMatch[offset]; ok {
				data["more"] = fmt.Sprintf(" || %s > %s", prev, offset)
			} else {
				data["more"] = ""
			}

			tmpl := `// Offset ({{.indx}}) '{{.name}}'
			if {{.offset}}, buf, err = marker.ReadOffset(buf); err != nil {
				return nil, err
			}`
			res = execTmpl(tmpl, data)
		}
		outs = append(outs, res)
	}

	// Marshal the dynamic parts

	c := 0

	for indx, i := range v.getObjs() {
		if !i.isFixed() {
			from := offsets[c]
			var to string
			if c == len(offsets)-1 {
				to = ""
			} else {
				to = offsets[c+1]
			}
			dst := fmt.Sprintf("tail[%s:%s]", from, to)
			tmpl := `// Field ({{.indx}}) '{{.name}}'
			{{.unmarshal}}
			`

			res := execTmpl(tmpl, map[string]interface{}{
				"indx":      indx,
				"name":      i.name,
				"from":      from,
				"to":        to,
				"unmarshal": i.unmarshal(dst),
			})
			outs = append(outs, res)
			c++
		}
	}

	str += strings.Join(outs, "\n\n")

	if len(offsets) != 0 {
		// it is a dynamic element, we received the
		// str += fmt.Sprintf("\nreturn tail[%s:], nil", offsets[len(offsets)-1])
		str += "\nreturn"
	} else {
		// it is a static element, it should have consumed the whole buffer
		str += "\nreturn buf, nil"
	}

	return
}

// createItem is used to initialize slices of objects
func (v *Value) createSlice(useNumVariable bool) string {
	var sizeU64 Size
	var isVectorCreate bool
	if obj, ok := v.typ.(*List); ok {
		sizeU64 = obj.MaxSize
		isVectorCreate = true
	} else if obj, ok := v.typ.(*Vector); ok {
		sizeU64 = obj.Size
		isVectorCreate = obj.IsDyn
	} else {
		panic("BUG: create item is only intended to be used with vectors and lists")
	}

	size := sizeU64.MarshalTemplate()
	// when useNumVariable is specified, we assume there is a 'num' variable generated beforehand with the expected size.
	if useNumVariable {
		size = "num"
	}

	inner := getElem(v.typ)
	switch obj := inner.typ.(type) {
	case *Uint:
		// []int uses the Extend functions in the fastssz package
		return fmt.Sprintf("::.%s = ssz.Extend(::.%s, %s)", v.name, v.name, size)

	case *Container:
		// []*(ref.)Struct{}
		ptr := "*"
		if inner.noPtr {
			ptr = ""
		}
		return fmt.Sprintf("::.%s = make([]%s%s, %s)", v.name, ptr, inner.objRef(), size)

	case *Bytes:
		if !isVectorCreate {
			return ""
		}

		// Check for a type alias.
		ref := inner.objRef()
		if ref != "" {
			return fmt.Sprintf("::.%s = make([]%s, %s)", v.name, ref, size)
		}

		if obj.IsFixed() {
			return fmt.Sprintf("::.%s = make([][%d]byte, %s)", v.name, obj.Size, size)
		}

		return fmt.Sprintf("::.%s = make([][]byte, %s)", v.name, size)

	default:
		panic(fmt.Sprintf("create not implemented for %s type %s", v.name, inner.Type()))
	}
}
