package main

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
	// we use dst as the input buffer where the SSZ data to decode the value is.
	switch v.t {
	case TypeContainer:
		return v.umarshalContainer(false, dst)

	case TypeBytes:
		// both fixed and dynamic are decoded equally
		return fmt.Sprintf("::.%s = append(::.%s, %s...)", v.name, v.name, dst)

	case TypeUint:
		return fmt.Sprintf("::.%s = ssz.Unmarshall%s(%s)", v.name, uintVToName(v), dst)

	case TypeBitList:
		return fmt.Sprintf("::.%s = append(::.%s, %s...)", v.name, v.name, dst)

	case TypeVector:
		if v.e.isFixed() {
			dst = fmt.Sprintf("%s[ii*%d: (ii+1)*%d]", dst, v.e.n, v.e.n)

			tmpl := `{{.create}}
			for ii := 0; ii < {{.size}}; ii++ {
				{{.unmarshal}}
			}`
			return execTmpl(tmpl, map[string]interface{}{
				"create":    v.createSlice(),
				"size":      v.s,
				"unmarshal": v.e.unmarshal(dst),
			})
		}
		fallthrough

	case TypeList:
		return v.unmarshalList()

	case TypeBool:
		return fmt.Sprintf("::.%s = ssz.UnmarshalBool(%s)", v.name, dst)

	default:
		panic(fmt.Errorf("unmarshal not implemented for type %d", v.t))
	}
}

func (v *Value) unmarshalList() string {

	// The Go field must have a 'ssz-max' tag to set the maximum number of items
	maxSize := v.s

	// In order to use createSlice with a dynamic list we need to set v.s to 0
	v.s = 0

	if v.e.isFixed() {
		dst := fmt.Sprintf("buf[ii*%d: (ii+1)*%d]", v.e.n, v.e.n)

		tmpl := `num, ok := ssz.DivideInt(len(buf), {{.size}})
		if !ok {
			return errDivideInt
		}
		if num > {{.max}} {
			return errListTooBig
		}
		{{.create}}
		for ii := 0; ii < num; ii++ {
			{{.unmarshal}}
		}`
		return execTmpl(tmpl, map[string]interface{}{
			"size":      v.e.n,
			"max":       maxSize,
			"create":    v.createSlice(),
			"unmarshal": v.e.unmarshal(dst),
		})
	}

	if v.t == TypeVector {
		panic("it cannot happen")
	}

	// Decode list with a dynamic element. 'ssz.DecodeDynamicLength' ensures
	// that the number of elements do not surpass the 'ssz-max' tag.

	tmpl := `num, err := ssz.DecodeDynamicLength(buf, {{.size}})
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

	v.e.name = v.name + "[indx]"

	data := map[string]interface{}{
		"size":      maxSize,
		"create":    v.createSlice(),
		"unmarshal": v.e.unmarshal("buf"),
	}
	return execTmpl(tmpl, data)
}

func (v *Value) umarshalContainer(start bool, dst string) (str string) {
	if !start {
		tmpl := `if ::.{{.name}} == nil {
			::.{{.name}} = new({{.obj}})
		}
		if err = ::.{{.name}}.UnmarshalSSZ({{.dst}}); err != nil {
			return err
		}`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"obj":  v.objRef(),
			"dst":  dst,
		})
	}

	var offsets []string
	offsetsMatch := map[string]string{}

	for indx, i := range v.o {
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
		return errSize
	}
	{{if .offsets}}
		tail := buf
		var {{.offsets}} uint64
	{{end}}
	`

	str += execTmpl(tmpl, map[string]interface{}{
		"cmp":     cmp,
		"size":    v.n,
		"offsets": strings.Join(offsets, ", "),
	})

	var o0 uint64

	// Marshal the fixed part and offsets

	outs := []string{}
	for indx, i := range v.o {

		// How much it increases on every item
		var incr uint64
		if i.isFixed() {
			incr = i.n
		} else {
			incr = 4
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
			if {{.offset}} = ssz.ReadOffset({{.dst}}); {{.offset}} > size {{.more}} {
				return errOffset
			}
			`
			res = execTmpl(tmpl, data)
		}
		outs = append(outs, res)
	}

	// Marshal the dynamic parts

	c := 0
	for indx, i := range v.o {
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
func (v *Value) createSlice() string {
	if v.t != TypeVector && v.t != TypeList {
		panic("BUG: create item is only intended to be used with vectors and lists")
	}

	// If v.s is set (fixed slice) we use that value, otherwise (variable size)
	// we assume there is a 'num' variable generated beforehand with the expected size.
	var size string
	if v.s == 0 {
		size = "num"
	} else {
		size = strconv.Itoa(int(v.s))
	}

	switch v.e.t {
	case TypeUint:
		// []int uses the Extend functions in the fastssz package
		return fmt.Sprintf("::.%s = ssz.Extend%s(::.%s, %s)", v.name, uintVToName(v.e), v.name, size)

	case TypeContainer:
		// []*(ref.)Struct{}
		return fmt.Sprintf("::.%s = make([]*%s, %s)", v.name, v.e.objRef(), size)

	case TypeBytes:
		// [][]byte
		return fmt.Sprintf("::.%s = make([][]byte, %s)", v.name, size)

	default:
		panic(fmt.Sprintf("create not implemented for type %s", v.e.t.String()))
	}
}
