package generator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// size creates a function that returns the SSZ size of the struct. There are two components:
// 1. Fixed: Size that we can determine at compilation time (i.e. uint, fixed bytes, fixed vector...)
// 2. Dynamic: Size that depends on the input (i.e. lists, dynamic containers...)
// Note that if any of the internal fields of the struct is nil, we will not fail, only not add up
// that field to the size. It is up to other methods like marshal to fail on that scenario.
func (e *env) size(name string, v *Value) string {
	tmpl := `// SizeSSZ returns the ssz encoded size in bytes for the {{.name}} object
	func (:: *{{.name}}) SizeSSZ() (size int) {
		size = {{.fixed}}{{if .dynamic}}

		{{.dynamic}}
		{{end}}
		return
	}`

	str := execTmpl(tmpl, map[string]interface{}{
		"name":    name,
		"fixed":   v.fixedSize(),
		"dynamic": v.sizeContainer("size", true),
	})
	return appendObjSignature(str, v)
}

func (v *Value) fixedSize() uint64 {
	switch obj := v.v2.(type) {
	case *Bool:
		return 1
	case *Uint:
		return obj.Size
	case *Int:
		return obj.Size
	case *Bytes:
		return obj.Size
	case *BitList:
		return obj.Size
	case *Container:
		var fixed uint64
		for _, f := range v.getObjs() {
			if f.isFixed() {
				fixed += f.fixedSize()
			} else {
				// we don't want variable size objects to recursively calculate their inner sizes
				fixed += bytesPerLengthOffset
			}
		}
		return fixed
	case *Vector:
		if obj.Elem.isFixed() {
			return obj.Size * obj.Elem.fixedSize()
		} else {
			return obj.Size * bytesPerLengthOffset
		}

	case *Reference:
		if !v.isFixed() {
			return bytesPerLengthOffset
		}
		return obj.Size

	default:
		panic(fmt.Errorf("fixed size not implemented for type %s", reflect.TypeOf(v.v2)))
	}
}

func (v *Value) sizeContainer(name string, start bool) string {
	if !start {
		tmpl := `{{if .check}} if ::.{{.name}} == nil {
			::.{{.name}} = new({{ref .obj}})
		}
		{{end}} {{ .dst }} += ::.{{.name}}.SizeSSZ()`

		check := true
		if v.isListElem() {
			check = false
		}
		if v.noPtr {
			check = false
		}
		return execTmpl(tmpl, map[string]interface{}{
			"name":  v.name,
			"dst":   name,
			"obj":   v,
			"check": check,
		})
	}
	out := []string{}
	for indx, v := range v.getObjs() {
		if !v.isFixed() {
			out = append(out, fmt.Sprintf("// Field (%d) '%s'\n%s", indx, v.name, v.size(name)))
		}
	}
	return strings.Join(out, "\n\n")
}

// 'name' is the name of target variable we assign the size too. We also use this function
// during marshalling to figure out the size of the offset
func (v *Value) size(name string) string {
	if v.isFixed() {
		if v.t == TypeContainer {
			return v.sizeContainer(name, false)
		}
		if v.fixedSize() == 1 {
			return name + "++"
		}
		return name + " += " + strconv.Itoa(int(v.fixedSize()))
	}

	switch v.v2.(type) {
	case *Container, *Reference:
		return v.sizeContainer(name, false)

	case *BitList:
		return fmt.Sprintf(name+" += len(::.%s)", v.name)

	case *Bytes:
		return fmt.Sprintf(name+" += len(::.%s)", v.name)

	case *List, *Vector:
		inner := getElem(v.v2)

		if inner.isFixed() {
			return fmt.Sprintf("%s += len(::.%s) * %d", name, v.name, inner.fixedSize())
		}
		inner.name = v.name + "[ii]"
		tmpl := `for ii := 0; ii < len(::.{{.name}}); ii++ {
			{{.size}} += 4
			{{.dynamic}}
		}`
		return execTmpl(tmpl, map[string]interface{}{
			"name":    v.name,
			"size":    name,
			"dynamic": inner.size(name),
		})

	default:
		panic(fmt.Errorf("size not implemented for type %s", v.t.String()))
	}
}
