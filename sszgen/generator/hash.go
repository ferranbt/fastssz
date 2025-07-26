package generator

import (
	"fmt"
	"strings"
)

// hashTreeRoot creates a function that SSZ hashes the structs,
func (e *env) hashTreeRoot(name string, v *Value) string {
	tmpl := `// HashTreeRoot ssz hashes the {{.name}} object
	func (:: *{{.name}}) HashTreeRoot() ([32]byte, error) {
		return ssz.HashWithDefaultHasher(::)
	}
	
	// HashTreeRootWith ssz hashes the {{.name}} object with a hasher	
	func (:: *{{.name}}) HashTreeRootWith(hh ssz.HashWalker) (err error) {
		{{.hashTreeRoot}}
		return
	}`

	data := map[string]interface{}{
		"name":         name,
		"hashTreeRoot": v.hashTreeRootContainer(true),
	}
	str := execTmpl(tmpl, data)
	return appendObjSignature(str, v)
}

func (v *Value) hashRoots(isList bool) string {
	innerObj := getElem(v.v2)

	subName := "i"
	if innerObj.c {
		subName += "[:]"
	}
	inner := ""
	if obj, ok := innerObj.v2.(*Bytes); ok && (obj.IsGoDyn || obj.IsList) {
		inner = `if len(i) != %d {
			err = ssz.ErrBytesLength
			return
		}
		`
		inner = fmt.Sprintf(inner, obj.Size)
	}

	var appendFn string
	var elemSize uint64

	if obj, ok := innerObj.v2.(*Bytes); ok {
		// [][]byte
		if obj.Size != 32 {
			// we need to use PutBytes in order to hash the result since
			// is higher than 32 bytes
			appendFn = "PutBytes"
			elemSize = obj.Size
		} else {
			appendFn = "Append"
			elemSize = 32
		}
	} else {
		// []uint64
		appendFn = "Append" + uintVToName2(*innerObj.v2.(*Uint))
		elemSize = uint64(innerObj.fixedSize())
	}

	var merkleize string
	if isList {
		innerListSize := v.v2.(*List).MaxSize

		// the limit for merkleize with mixin depends on the internal type
		// if the type is basic, the size depends on CalculateLimit
		// if the type is complex (TypeVector), the limit is the size.
		// TODO: Generalize a list of complex objects
		isComplex := false
		if innerObj.t == TypeBytes {
			// TypeVector alias
			isComplex = true
		}

		tmpl := `numItems := uint64(len(::.{{.name}}))
		hh.MerkleizeWithMixin(subIndx, numItems, {{if .isComplex}} {{.listSize}} {{ else }} ssz.CalculateLimit({{.listSize}}, numItems, {{.elemSize}}) {{ end }})`

		merkleize = execTmpl(tmpl, map[string]interface{}{
			"name":      v.name,
			"listSize":  innerListSize,
			"elemSize":  elemSize,
			"isComplex": isComplex,
		})

		// when doing []uint64 we need to round up the Hasher bytes to 32
		if innerObj.t == TypeUint {
			merkleize = "hh.FillUpTo32()\n" + merkleize
		}
	} else {
		merkleize = "hh.Merkleize(subIndx)"
	}

	tmpl := `{
		{{.outer}}subIndx := hh.Index()
		for _, i := range ::.{{.name}} {
			{{.inner}}hh.{{.appendFn}}({{.subName}})
		}
		{{.merkleize}}
	}`
	return execTmpl(tmpl, map[string]interface{}{
		"outer":     v.validate(),
		"inner":     inner,
		"name":      v.name,
		"subName":   subName,
		"appendFn":  appendFn,
		"merkleize": merkleize,
	})
}

// takes a "name" param so that the variable name can be replaced with a local name
// ie within a for loop for a list, the we want to refer to "elem" w/o a receiver variable
// when not specified, name will be set to "::." + v.name. In the final templating pass,
// the output formatter replaces all instances of "::" with the receiver variable for the container.
// appendBytes is a control variable which changes the fastssz.Hasher method used to handle byte slices
// when it is false, the default behavior is to call PutBytes, which merkleizes the buffer after appending
// the bytes. when true, the generated code calls AppendBytes32, which appends the bytes to the buffer
// with padding and leaves the merkleization for a following step. This is because in the case of ByteLists,
// the length of the list needs to be mixed in as part of the merkleization process, which happens in a separate
// call to MerkleizeWithMixin.
func (v *Value) hashTreeRoot(name string, appendBytes bool) string {
	if name == "" {
		name = "::." + v.name
	}

	switch obj := v.v2.(type) {
	case *Container, *Reference:
		return v.hashTreeRootContainer(false)

	case *Uint:
		if v.ref != "" || v.obj != "" {
			// alias to uint*
			name = fmt.Sprintf("%s(%s)", uintVToLowerCaseName2(obj), name)
		}
		bitLen := v.fixedSize() * 8
		return fmt.Sprintf("hh.PutUint%d(%s)", bitLen, name)

	case *BitList:
		tmpl := `if len({{.name}}) == 0 {
			err = ssz.ErrEmptyBitlist
			return
		}
		hh.PutBitlist({{.name}}, {{.size}})
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": name,
			"size": obj.Size,
		})

	case *Bool:
		return fmt.Sprintf("hh.PutBool(%s)", name)

	case *Vector:
		return v.hashRoots(false)

	case *List:
		if obj.Elem.isFixed() {
			if obj.Elem.t == TypeUint || obj.Elem.t == TypeBytes {
				return v.hashRoots(true)
			}
		}

		tmpl := `{
			subIndx := hh.Index()
			num := uint64(len({{.name}}))
			if num > {{.num}} {
				err = ssz.ErrIncorrectListSize
				return
			}
			for _, elem := range {{.name}} {
{{.htrCall}}
			}
			hh.MerkleizeWithMixin(subIndx, num, {{.num}})
		}`
		var htrCall string
		if obj.Elem.t == TypeBytes {
			eName := "elem"
			// ByteLists should be represented as Value with TypeBytes and .m set instead of .s (isFixed == true)
			htrCall = obj.Elem.hashTreeRoot(eName, true)
		} else {
			htrCall = execTmpl(`if err = elem.HashTreeRootWith(hh); err != nil {
	return
}`,
				map[string]interface{}{"name": name})
		}
		return execTmpl(tmpl, map[string]interface{}{
			"name":    name,
			"num":     obj.MaxSize,
			"htrCall": htrCall,
		})

	case *Time:
		return fmt.Sprintf("hh.PutUint64(uint64(%s.Unix()))", name)

	case *Bytes:
		if !obj.IsGoDyn && !obj.IsList {
			name += "[:]"
		}
		if v.isFixed() {
			tmpl := `{{.validate}}hh.PutBytes({{.name}})`
			return execTmpl(tmpl, map[string]interface{}{
				"validate": v.validate(),
				"name":     name,
				"size":     obj.Size,
			})
		} else {
			// dynamic bytes require special handling, need length mixed in
			hMethod := "Append"
			if appendBytes {
				hMethod = "AppendBytes32"
			}
			tmpl := `{
	elemIndx := hh.Index()
	byteLen := uint64(len({{.name}}))
	if byteLen > {{.maxLen}} {
		err = ssz.ErrIncorrectListSize
		return
    }
	hh.{{.hashMethod}}({{.name}})
	hh.MerkleizeWithMixin(elemIndx, byteLen, ({{.maxLen}}+31)/32)
}`
			return execTmpl(tmpl, map[string]interface{}{
				"hashMethod": hMethod,
				"name":       name,
				"maxLen":     obj.Size,
			})
		}

	default:
		panic(fmt.Errorf("hash not implemented for type %s", v.t.String()))
	}
}

func (v *Value) hashTreeRootContainer(start bool) string {
	if !start {
		tmpl := `{{ if .check }}if ::.{{.name}} == nil {
			::.{{.name}} = new({{ref .obj}})
		}
		{{ end }}if err = ::.{{.name}}.HashTreeRootWith(hh); err != nil {
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
	for indx, i := range v.getObjs() {
		// the call to hashTreeRoot below is ugly because it's currently hacked to support ByteLists
		// the first argument allows the element name to be overriden when calling .HashTreeRoot on it
		// used to specify the name "elem" when called as part of a for loop iteration. when the string
		// is empty, it defaults to the .name parameter of the value
		// the second field tells the code generator to specifically generate a call to AppendBytes32
		// this is used by List[List[byte, N]] so that lists of lists of bytes are not double-merkleized.
		str := fmt.Sprintf("// Field (%d) '%s'\n%s\n", indx, i.name, i.hashTreeRoot("", false))
		out = append(out, str)
	}

	tmpl := `indx := hh.Index()

	{{.fields}}

	hh.Merkleize(indx)`

	return execTmpl(tmpl, map[string]interface{}{
		"fields": strings.Join(out, "\n"),
	})
}
