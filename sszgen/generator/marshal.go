package generator

import (
	"fmt"
	"strings"
)

// marshal creates a function that encodes the structs in SSZ format. It creates two functions:
// 1. MarshalTo(dst []byte) marshals the content to the target array.
// 2. Marshal() marshals the content to a newly created array.
func (e *env) marshal(name string, v *Value) string {
	tmpl := `// MarshalSSZ ssz marshals the {{.name}} object
	func (:: *{{.name}}) MarshalSSZ() ([]byte, error) {
		return ssz.MarshalSSZ(::)
	}

	// MarshalSSZTo ssz marshals the {{.name}} object to a target array	
	func (:: *{{.name}}) MarshalSSZTo(buf []byte) (dst []byte, err error) {
		dst = buf
		{{.offset}}
		{{.marshal}}
		return
	}`

	data := map[string]interface{}{
		"name":    name,
		"marshal": v.marshalContainer(true),
		"offset":  "",
	}
	if !v.isFixed() {
		// offset is the position where the offset starts
		data["offset"] = "offset := ::.fixedSize()\n"
	}
	str := execTmpl(tmpl, data)
	return appendObjSignature(str, v)
}

func (v *Value) marshal() string {
	switch obj := v.typ.(type) {
	case *Bool:
		return fmt.Sprintf("dst = ssz.MarshalValue(dst, ::.%s)", v.name)

	case *Bytes:
		name := v.name
		if obj.IsFixed() {
			name += "[:]"
		}
		tmpl := `{{.validate}}dst = append(dst, ::.{{.name}}...)`

		return execTmpl(tmpl, map[string]interface{}{
			"validate": v.validate(),
			"name":     name,
		})

	case *Uint:
		var name string
		if v.ref != "" || v.obj != "" {
			// alias to uint*
			name = fmt.Sprintf("%s(::.%s)", uintVToLowerCaseName2(obj), v.name)
		} else {
			name = "::." + v.name
		}
		return fmt.Sprintf("dst = ssz.MarshalValue(dst, %s)", name)

	case *BitList:
		return fmt.Sprintf("%sdst = append(dst, ::.%s...)", v.validate(), v.name)

	case *Time:
		return fmt.Sprintf("dst = ssz.MarshalTime(dst, ::.%s)", v.name)

	case *List:
		return v.marshalList()

	case *Vector:
		if obj.Elem.isFixed() {
			return v.marshalVector()
		}
		return v.marshalList()

	case *Container, *Reference:
		return v.marshalContainer(false)

	default:
		panic(fmt.Errorf("marshal not implemented for type %s", v.Type()))
	}
}

func (v *Value) marshalList() string {
	inner := getElem(v.typ)
	inner.name = v.name + "[ii]"

	// bound check
	str := v.validate()

	if inner.isFixed() {
		tmpl := `for ii := 0; ii < len(::.{{.name}}); ii++ {
			{{.dynamic}}
		}`
		str += execTmpl(tmpl, map[string]interface{}{
			"name":    v.name,
			"dynamic": inner.marshal(),
		})
		return str
	}

	// encode a list of dynamic objects:
	// 1. write offsets for each
	// 2. marshal each element

	tmpl := `{
		offset = 4 * len(::.{{.name}})
		for ii := 0; ii < len(::.{{.name}}); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			{{.size}}
		}
	}
	for ii := 0; ii < len(::.{{.name}}); ii++ {
		{{.marshal}}
	}`

	str += execTmpl(tmpl, map[string]interface{}{
		"name":    v.name,
		"size":    inner.size("offset"),
		"marshal": inner.marshal(),
	})
	return str
}

func (v *Value) marshalVector() (str string) {
	inner := getElem(v.typ)

	obj := v.typ.(*Vector)
	inner.name = fmt.Sprintf("%s[ii]", v.name)

	tmpl := `{{.validate}}for ii := 0; ii < {{.size}}; ii++ {
		{{.marshal}}
	}`
	return execTmpl(tmpl, map[string]interface{}{
		"validate": v.validate(),
		"name":     v.name,
		"size":     obj.Size,
		"marshal":  inner.marshal(),
	})
}

func (v *Value) marshalContainer(start bool) string {
	if !start {
		tmpl := `{{ if .check }}if ::.{{.name}} == nil {
			::.{{.name}} = new({{ref .obj}})
		}
		{{ end }}if dst, err = ::.{{.name}}.MarshalSSZTo(dst); err != nil {
			return
		}`
		// validate only for fixed structs
		check := v.isFixed()
		if v.isListElem() {
			check = false
		}
		if v.noPtr {
			check = false
		}
		return execTmpl(tmpl, map[string]interface{}{
			"name":  v.name,
			"obj":   v,
			"check": check,
		})
	}

	out := []string{}

	lastVariableIndx := -1
	for indx, i := range v.getObjs() {
		if !i.isFixed() {
			lastVariableIndx = indx
		}
	}
	for indx, i := range v.getObjs() {
		var str string
		if i.isFixed() {
			// write the content
			str = fmt.Sprintf("// Field (%d) '%s'\n%s\n", indx, i.name, i.marshal())
		} else {
			// write the offset
			str = fmt.Sprintf("// Offset (%d) '%s'\ndst = ssz.WriteOffset(dst, offset)\n", indx, i.name)
			// Update the offset for the next variable field.
			// We don't need to update the offset if the current
			// field is the last variable field in the container.
			if indx != lastVariableIndx {
				str += fmt.Sprintf("%s\n", i.size("offset"))
			}
		}
		out = append(out, str)
	}

	// write the dynamic parts
	for indx, i := range v.getObjs() {
		if !i.isFixed() {
			out = append(out, fmt.Sprintf("// Field (%d) '%s'\n%s\n", indx, i.name, i.marshal()))
		}
	}
	return strings.Join(out, "\n")
}
