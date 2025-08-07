package generator

import (
	"fmt"
	"reflect"
	"strings"
)

// size creates a function that returns the SSZ size of the struct. There are two components:
// 1. Fixed: Size that we can determine at compilation time (i.e. uint, fixed bytes, fixed vector...)
// 2. Dynamic: Size that depends on the input (i.e. lists, dynamic containers...)
// Note that if any of the internal fields of the struct is nil, we will not fail, only not add up
// that field to the size. It is up to other methods like marshal to fail on that scenario.
func (e *env) size(name string, v *Value) string {
	tmpl := `const {{.fixedName}} = {{.fixed}} 

	// SizeSSZ returns the ssz encoded size in bytes for the {{.name}} object
	func (:: *{{.name}}) SizeSSZ() (size int) {
		size = {{.fixedName}}{{if .dynamic}}

		{{.dynamic}}
		{{end}}
		return
	}`

	fmt.Println("XX")
	fmt.Println(name, v.name)

	str := execTmpl(tmpl, map[string]interface{}{
		"name":      name,
		"fixedName": v.fixedSizeName(),
		"fixed":     v.fixedSizeForContainer(),
		"dynamic":   v.sizeContainer("size", true),
	})
	return appendObjSignature(str, v)
}

func (v *Value) fixedSizeForContainer() string {
	if !v.isContainer() {
		panic(fmt.Sprintf("fixedSizeForContainer called on non-container type %s", reflect.TypeOf(v.typ)))
	}

	sizes := []string{"0"}
	for _, f := range v.getObjs() {
		switch obj := f.typ.(type) {
		case *Vector:
			if obj.Elem.isFixed() {
				sizes = append(sizes, fmt.Sprintf("%s * %s", obj.Size.MarshalTemplate(), obj.Elem.fixedSize()))
			} else {
				sizes = append(sizes, fmt.Sprintf("%s * %d", obj.Size.MarshalTemplate(), bytesPerLengthOffset))
			}
		case *List:
			// lists are variable size, so we don't add them to the fixed size
			sizes = append(sizes, "0 /*list*/")
		default:
			sizes = append(sizes, f.fixedSize())
		}
	}

	return strings.Join(sizes, " + ")
}

func (v *Value) fixedSize() string {
	switch obj := v.typ.(type) {
	case *Bool:
		return "1"
	case *Uint:
		return fmt.Sprintf("%d", obj.Size)
	case *Int:
		return fmt.Sprintf("%d", obj.Size)
	case *Bytes:
		return fmt.Sprintf("%d", obj.Size)
	case *BitList:
		return fmt.Sprintf("%d", obj.Size)
	case *Container:
		return v.fixedSizeName()
	default:
		panic(fmt.Errorf("fixed size not implemented for type %s", reflect.TypeOf(v.typ)))
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
		if _, ok := v.typ.(*Container); ok {
			return v.sizeContainer(name, false)
		}
		return name + " += " + v.fixedSize()
	}

	switch v.typ.(type) {
	case *Container, *Reference:
		return v.sizeContainer(name, false)

	case *BitList:
		return fmt.Sprintf(name+" += len(::.%s)", v.name)

	case *Bytes:
		return fmt.Sprintf(name+" += len(::.%s)", v.name)

	case *List, *Vector:
		inner := getElem(v.typ)

		if inner.isFixed() {
			return fmt.Sprintf("%s += len(::.%s) * %s", name, v.name, inner.fixedSize())
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
		panic(fmt.Errorf("size not implemented for type %s", v.Type()))
	}
}
