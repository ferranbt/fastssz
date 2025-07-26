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
		var err error
		{{.unmarshal}}
		return err
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
			return fmt.Sprintf("copy(::.%s[:], %s)", v.name, dst)
		}
		validate := ""
		if !v.isFixed() {
			// dynamic bytes, we need to validate the size of the buffer
			validate = fmt.Sprintf("if len(%s) > %d { return ssz.ErrBytesLength }\n", dst, obj.Size)
		}

		// both fixed and dynamic are decoded equally
		tmpl := `{{.validate}}if cap(::.{{.name}}) == 0 {
			{{if .isRef}} ::.{{.name}} = {{ ref .obj }}(make([]byte, 0, len({{.dst}}))) {{ else }} ::.{{.name}} = make([]byte, 0, len({{.dst}})) {{ end }}
		}
		::.{{.name}} = append(::.{{.name}}, {{.dst}}...)`
		return execTmpl(tmpl, map[string]interface{}{
			"validate": validate,
			"name":     v.name,
			"dst":      dst,
			"size":     obj.Size,
			"isRef":    v.ref != "",
			"obj":      v,
		})

	case *BitList:
		tmpl := `if err = ssz.ValidateBitlist({{.dst}}, {{.size}}); err != nil {
			return err
		}
		if cap(::.{{.name}}) == 0 {
			::.{{.name}} = make([]byte, 0, len({{.dst}}))
		}
		::.{{.name}} = append(::.{{.name}}, {{.dst}}...)`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"dst":  dst,
			"size": obj.Size,
		})

	case *Uint:
		if v.ref != "" {
			// alias, we need to cast the value
			return fmt.Sprintf("::.%s = %s(ssz.Unmarshall%s(%s))", v.name, v.objRef(), uintVToName2(*obj), dst)
		}
		if v.obj != "" {
			// alias to a type on the same package
			return fmt.Sprintf("::.%s = %s(ssz.Unmarshall%s(%s))", v.name, v.obj, uintVToName2(*obj), dst)
		}
		return fmt.Sprintf("::.%s = ssz.Unmarshall%s(%s)", v.name, uintVToName2(*obj), dst)

	case *Bool:
		return fmt.Sprintf("::.%s = ssz.UnmarshalBool(%s)", v.name, dst)

	case *Time:
		return fmt.Sprintf("::.%s = ssz.UnmarshalTime(%s)", v.name, dst)

	case *Vector:
		if obj.Elem.isFixed() {
			dst = fmt.Sprintf("%s[ii*%d: (ii+1)*%d]", dst, obj.Elem.fixedSize(), obj.Elem.fixedSize())

			tmpl := `{{.create}}
			for ii := 0; ii < {{.size}}; ii++ {
				{{.unmarshal}}
			}`
			return execTmpl(tmpl, map[string]interface{}{
				"create":    v.createSlice(false),
				"size":      obj.Size,
				"unmarshal": obj.Elem.unmarshal(dst),
			})
		} else {
			return v.unmarshalList()
		}

	case *List:
		return v.unmarshalList()

	default:
		panic(fmt.Errorf("unmarshal not implemented for type %s", v.Type()))
	}
}

func (v *Value) unmarshalList() string {
	var size uint64
	if obj, ok := v.typ.(*List); ok {
		size = obj.MaxSize
	} else if obj, ok := v.typ.(*Vector); ok {
		size = obj.Size
	} else {
		panic(fmt.Errorf("unmarshalList not implemented for type %s", v.Type()))
	}

	inner := getElem(v.typ)
	if inner.isFixed() {
		dst := fmt.Sprintf("buf[ii*%d: (ii+1)*%d]", inner.fixedSize(), inner.fixedSize())

		tmpl := `num, err := ssz.DivideInt2(len(buf), {{.size}}, {{.max}})
		if err != nil {
			return err
		}
		{{.create}}
		for ii := 0; ii < num; ii++ {
			{{.unmarshal}}
		}`
		return execTmpl(tmpl, map[string]interface{}{
			"size":      inner.fixedSize(),
			"max":       size,
			"create":    v.createSlice(true),
			"unmarshal": inner.unmarshal(dst),
		})
	}

	// Decode list with a dynamic element. 'ssz.DecodeDynamicLength' ensures
	// that the number of elements do not surpass the 'ssz-max' tag.

	tmpl := `num, err := ssz.DecodeDynamicLength(buf, {{.max}})
	if err != nil {
		return err
	}
	{{.create}}
	err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
		{{.unmarshal}}
		return nil
	})
	if err != nil {
		return err
	}`

	inner.name = v.name + "[indx]"

	data := map[string]interface{}{
		"max":       size,
		"create":    v.createSlice(true),
		"unmarshal": inner.unmarshal("buf"),
	}
	return execTmpl(tmpl, data)
}

func (v *Value) umarshalContainer(start bool, dst string) (str string) {
	if !start {
		tmpl := `{{ if .check }}if ::.{{.name}} == nil {
			::.{{.name}} = new({{ref .obj}})
		}
		{{ end }}if err = ::.{{.name}}.UnmarshalSSZ({{.dst}}); err != nil {
			return err
		}`
		check := true
		if v.noPtr {
			check = false
		}
		return execTmpl(tmpl, map[string]interface{}{
			"name":  v.name,
			"obj":   v,
			"dst":   dst,
			"check": check,
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

	tmpl := `size := uint64(len(buf))
	if size {{.cmp}} {{.size}} {
		return ssz.ErrSize
	}
	{{if .offsets}}
		tail := buf
		var {{.offsets}} uint64
	{{end}}
	`

	str += execTmpl(tmpl, map[string]interface{}{
		"cmp":     cmp,
		"size":    v.fixedSize(),
		"offsets": strings.Join(offsets, ", "),
	})

	var o0 uint64

	// Marshal the fixed part and offsets

	// used for bounds checking of variable length offsets.
	// for the first offset, use the size of the fixed-length data
	// as the minimum boundary. subsequent offsets will replace this
	// value with the name of the previous offset variable.
	firstOffsetCheck := fmt.Sprintf("%d", v.fixedSize())
	outs := []string{}
	for indx, i := range v.getObjs() {

		// How much it increases on every item
		var incr uint64
		if i.isFixed() {
			incr = i.fixedSize()
		} else {
			incr = bytesPerLengthOffset
		}

		dst = fmt.Sprintf("%s[%d:%d]", "buf", o0, o0+incr)
		o0 += incr

		var res string
		if i.isFixed() {
			res = fmt.Sprintf("// Field (%d) '%s'\n%s\n\n", indx, i.name, i.unmarshal(dst))

		} else {
			// read the offset
			offset := "o" + strconv.Itoa(indx)

			data := map[string]interface{}{
				"indx":             indx,
				"name":             i.name,
				"offset":           offset,
				"dst":              dst,
				"firstOffsetCheck": firstOffsetCheck,
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
			if {{.offset}} = ssz.ReadOffset({{.dst}}); {{.offset}} > size {{.more}} {
				return ssz.ErrOffset
			}
			{{ if .firstOffsetCheck }}
			if {{.offset}} != {{.firstOffsetCheck}} {
				return ssz.ErrInvalidVariableOffset
			}
			{{ end }}
			`
			res = execTmpl(tmpl, data)
			firstOffsetCheck = ""
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
			tmpl := `// Field ({{.indx}}) '{{.name}}'
			{
				buf = tail[{{.from}}:{{.to}}]
				{{.unmarshal}}
			}`
			res := execTmpl(tmpl, map[string]interface{}{
				"indx":      indx,
				"name":      i.name,
				"from":      from,
				"to":        to,
				"unmarshal": i.unmarshal("buf"),
			})
			outs = append(outs, res)
			c++
		}
	}

	str += strings.Join(outs, "\n\n")
	return
}

// createItem is used to initialize slices of objects
func (v *Value) createSlice(useNumVariable bool) string {
	var sizeU64 uint64
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

	size := strconv.Itoa(int(sizeU64))
	// when useNumVariable is specified, we assume there is a 'num' variable generated beforehand with the expected size.
	if useNumVariable {
		size = "num"
	}

	inner := getElem(v.typ)
	switch obj := inner.typ.(type) {
	case *Uint:
		// []int uses the Extend functions in the fastssz package
		return fmt.Sprintf("::.%s = ssz.Extend%s(::.%s, %s)", v.name, uintVToName2(*obj), v.name, size)

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
