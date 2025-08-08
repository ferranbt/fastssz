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
	tmpl := `// fixedSize returns the fixed size of the {{.name}} object
	func (:: *{{.name}}) fixedSize() int {
		return int({{.fixed}})
	}
	
	// SizeSSZ returns the ssz encoded size in bytes for the {{.name}} object
	func (:: *{{.name}}) SizeSSZ() (size int) {
		size = ::.fixedSize(){{if .dynamic}}

		{{.dynamic}}
		{{end}}
		return
	}`

	str := execTmpl(tmpl, map[string]interface{}{
		"name":    name,
		"fixed":   v.fixedSizeForContainer(),
		"dynamic": v.sizeContainer("size", true),
	})
	return appendObjSignature(str, v)
}

func (v *Value) fixedSizeForContainer() string {
	acc := &SizeAccumulator{
		Size: 0,
		Vars: []string{},
	}

	v.fixedSizeForContainerAcc(acc)
	return acc.String()
}

type SizeAccumulator struct {
	Size uint64
	Vars []string
}

func NewSizeAccumulator() *SizeAccumulator {
	return &SizeAccumulator{
		Size: 0,
		Vars: []string{},
	}
}

func (s *SizeAccumulator) IsVariable() bool {
	return len(s.Vars) > 0
}

func (s *SizeAccumulator) AddVar(name string) {
	if name == "" {
		return
	}
	s.Vars = append(s.Vars, name)
}

func (s *SizeAccumulator) AddInt(size uint64) {
	s.Size += size
}

func (s *SizeAccumulator) Merge(ss *SizeAccumulator) {
	if ss.IsVariable() {
		// If there are variables, that means the value cannot be accumulated as a fixed
		// int, it includes arightmetic operations and we have to use the parenthesis to
		// avoid precedence issues
		s.Vars = append(s.Vars, "("+ss.String()+")")
	} else {
		// If there are no variables, just merge the size as an int
		s.Size += ss.Size
	}
}

func (s *SizeAccumulator) String() string {
	vars := s.Vars
	if s.Size != 0 {
		vars = append([]string{fmt.Sprintf("%d", s.Size)}, vars...)
	}
	if len(vars) == 0 {
		return "0"
	}
	return strings.Join(vars, " + ")
}

func (v *Value) fixedSizeForContainerAcc(acc *SizeAccumulator) {
	if !v.isContainer() {
		panic(fmt.Sprintf("fixedSizeForContainer called on non-container type %s", reflect.TypeOf(v.typ)))
	}

	subAcc := NewSizeAccumulator()
	for _, f := range v.getObjs() {
		switch obj := f.typ.(type) {
		case *Vector:
			if obj.Elem.isFixed() {
				vectorAcc := NewSizeAccumulator()
				obj.Elem.fixedSizeAcc(vectorAcc)

				if obj.Size.Size != 0 {
					// two cases: fixed size or variable size for the inner element
					if vectorAcc.IsVariable() {
						// variable size, accumulate on top of subAcc
						subAcc.AddVar(fmt.Sprintf("(%d * %s)", obj.Size.Size, vectorAcc.String()))
					} else {
						// fixed size, we can precompute all the size
						subAcc.AddInt(obj.Size.Size * vectorAcc.Size)
					}
				} else {
					// variable size, it is going to be an arithmetic expression
					subAcc.AddVar(fmt.Sprintf("(%s * %s)", obj.Size.VarSize, vectorAcc.String()))
				}
			} else {
				if obj.Size.Size != 0 {
					// known size at compilation time. precompute it.
					subAcc.AddInt(obj.Size.Size * bytesPerLengthOffset)
				} else {
					// variable
					subAcc.AddVar(fmt.Sprintf("(%s * %d)", obj.Size.VarSize, bytesPerLengthOffset))
				}
			}

		case *List:
			// lists are variable size, so we don't add them to the fixed size
			subAcc.AddInt(bytesPerLengthOffset)
		default:
			f.fixedSizeAcc(subAcc)
		}
	}

	acc.Merge(subAcc)
}

func (v *Value) fixedSize() string {
	acc := &SizeAccumulator{
		Size: 0,
		Vars: []string{},
	}

	v.fixedSizeAcc(acc)
	return acc.String()
}

func (v *Value) fixedSizeAcc(acc *SizeAccumulator) {
	switch obj := v.typ.(type) {
	case *Bool:
		acc.AddInt(1)
	case *Uint:
		acc.AddInt(obj.Size)
	case *Int:
		acc.AddInt(obj.Size)
	case *Bytes:
		acc.AddInt(obj.Size)
	case *BitList:
		acc.AddInt(bytesPerLengthOffset)
	case *Time:
		acc.AddInt(8)
	case *Container:
		if v.isFixed() {
			v.fixedSizeForContainerAcc(acc)
		} else {
			acc.AddInt(bytesPerLengthOffset)
		}
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
