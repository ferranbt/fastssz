package main

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
		buf := make([]byte, ::.SizeSSZ())
		return ::.MarshalSSZTo(buf[:0])
	}

	// MarshalSSZTo ssz marshals the {{.name}} object to a target array	
	func (:: *{{.name}}) MarshalSSZTo(dst []byte) ([]byte, error) {
		var err error
		{{.offset}}
		{{.marshal}}
		return dst, err
	}`

	data := map[string]interface{}{
		"name":    name,
		"marshal": v.marshalContainer(true),
		"offset":  "",
	}
	if !v.isFixed() {
		// offset is the position where the offset starts
		data["offset"] = fmt.Sprintf("offset := int(%d)\n", v.n)
	}
	str := execTmpl(tmpl, data)
	return appendObjSignature(str, v)
}

func (v *Value) marshal() string {
	switch v.t {
	case TypeContainer:
		return v.marshalContainer(false)

	case TypeBytes:
		if v.isFixed() {
			// fixed. It ensures that the size is correct
			return fmt.Sprintf("if dst, err = ssz.MarshalFixedBytes(dst, ::.%s, %d); err != nil {\n return nil, errMarshalFixedBytes\n}", v.name, v.s)
		}
		// dynamic
		return fmt.Sprintf("if len(::.%s) > %d {\n return nil, errMarshalDynamicBytes\n}\ndst = append(dst, ::.%s...)", v.name, v.m, v.name)

	case TypeUint:
		return fmt.Sprintf("dst = ssz.Marshal%s(dst, ::.%s)", uintVToName(v), v.name)

	case TypeBitList:
		return fmt.Sprintf("dst = append(dst, ::.%s...)", v.name)

	case TypeBool:
		return fmt.Sprintf("dst = ssz.MarshalBool(dst, ::.%s)", v.name)

	case TypeVector:
		if v.e.isFixed() {
			return v.marshalVector()
		}
		fallthrough

	case TypeList:
		return v.marshalList()

	default:
		panic(fmt.Errorf("marshal not implemented for type %s", v.t.String()))
	}
}

func (v *Value) marshalList() string {
	v.e.name = v.name + "[ii]"

	// bound check
	str := fmt.Sprintf("if len(::.%s) > %d {\n return nil, errMarshalList\n}\n", v.name, v.s)

	if v.e.isFixed() {
		tmpl := `for ii := 0; ii < len(::.{{.name}}); ii++ {
			{{.dynamic}}
		}`
		str += execTmpl(tmpl, map[string]interface{}{
			"name":    v.name,
			"dynamic": v.e.marshal(),
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
		"size":    v.e.size("offset"),
		"marshal": v.e.marshal(),
	})
	return str
}

func (v *Value) marshalVector() (str string) {
	v.e.name = fmt.Sprintf("%s[ii]", v.name)

	tmpl := `if len(::.{{.name}}) != {{.size}} {
		return nil, errMarshalVector
	}
	for ii := 0; ii < {{.size}}; ii++ {
		{{.marshal}}
	}`
	return execTmpl(tmpl, map[string]interface{}{
		"name":    v.name,
		"size":    v.s,
		"marshal": v.e.marshal(),
	})
}

func (v *Value) marshalContainer(start bool) string {
	if !start {
		tmpl := `{{ if .check }}if ::.{{.name}} == nil {
			return nil, errNilStruct
		}
		{{ end }}if dst, err = ::.{{.name}}.MarshalSSZTo(dst); err != nil {
			return nil, err
		}`
		// validate only for fixed structs
		check := v.isFixed()
		if v.isListElem() {
			check = false
		}
		return execTmpl(tmpl, map[string]interface{}{
			"name":  v.name,
			"check": check,
		})
	}

	offset := v.n
	out := []string{}

	for indx, i := range v.o {
		var str string
		if i.isFixed() {
			// write the content
			str = fmt.Sprintf("// Field (%d) '%s'\n%s\n", indx, i.name, i.marshal())
		} else {
			// write the offset
			str = fmt.Sprintf("// Offset (%d) '%s'\ndst = ssz.WriteOffset(dst, offset)\n%s\n", indx, i.name, i.size("offset"))
			offset += i.n
		}
		out = append(out, str)
	}

	// write the dynamic parts
	for indx, i := range v.o {
		if !i.isFixed() {
			out = append(out, fmt.Sprintf("// Field (%d) '%s'\n%s\n", indx, i.name, i.marshal()))
		}
	}
	return strings.Join(out, "\n")
}
