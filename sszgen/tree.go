package main

import (
	"fmt"
	"strings"
)

// getTree creates a function that SSZ hashes the structs,
func (e *env) getTree(name string, v *Value) string {
	tmpl := `// GetTree returns tree-backing for the {{.name}} object
	func (:: *{{.name}}) GetTree() (*ssz.Node, error) {
		{{.getTree}}
	}`

	data := map[string]interface{}{
		"name":    name,
		"getTree": v.getTreeContainer(true),
	}
	str := execTmpl(tmpl, data)
	return appendObjSignature(str, v)
}

func (v *Value) getTrees(isList bool, elem Type) string {
	subName := "i"
	if v.e.c {
		subName += "[:]"
	}
	inner := ""
	if !v.e.c && elem == TypeBytes {
		inner = `if len(i) != %d {
			err = ssz.ErrBytesLength
			return nil, err
		}
		`
		inner = fmt.Sprintf(inner, v.e.s)
	}

	var appendFn string
	var elemSize uint64
	if elem == TypeBytes {
		// [][]byte
		appendFn = "Append"
		elemSize = 32
	} else {
		// []uint64
		appendFn = "AppendUint64"
		elemSize = 8
	}

	var merkleize string
	if isList {
		tmpl := `numItems := uint64(len(::.{{.name}}))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit({{.listSize}}, numItems, {{.elemSize}}))`

		merkleize = execTmpl(tmpl, map[string]interface{}{
			"name":     v.name,
			"listSize": v.s,
			"elemSize": elemSize,
		})

		// when doing []uint64 we need to round up the Hasher bytes to 32
		if elem == TypeUint {
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
		"outer":     v.validate("return nil, err"),
		"inner":     inner,
		"name":      v.name,
		"subName":   subName,
		"appendFn":  appendFn,
		"merkleize": merkleize,
	})
}

func (v *Value) getTree() string {
	switch v.t {
	case TypeContainer, TypeReference:
		return v.getTreeContainer(false)

	case TypeBytes:
		// There are only fixed []byte
		name := v.name
		if v.c {
			name += "[:]"
		}

		tmpl := `{{.validate}}tmp = ssz.LeafFromBytes(::.{{.name}})`
		return execTmpl(tmpl, map[string]interface{}{
			"validate": v.validate("return nil, err"),
			"name":     name,
			"size":     v.s,
		})

	case TypeUint:
		var name string
		if v.ref != "" || v.obj != "" {
			// alias to Uint64
			name = fmt.Sprintf("uint64(::.%s)", v.name)
		} else {
			name = "::." + v.name
		}
		bitLen := v.n * 8
		return fmt.Sprintf("tmp = ssz.LeafFromUint%d(%s)", bitLen, name)

	case TypeBitList:
		tmpl := `if len(::.{{.name}}) == 0 {
			err = ssz.ErrEmptyBitlist
			return
		}
		hh.PutBitlist(::.{{.name}}, {{.size}})
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"size": v.m,
		})

	case TypeBool:
		return fmt.Sprintf("tmp = ssz.LeafFromBool(::.%s)", v.name)

	case TypeVector:
		return v.getTrees(false, v.e.t)

	case TypeList:
		if v.e.isFixed() {
			if v.e.t == TypeUint || v.e.t == TypeBytes {
				// return hashBasicSlice(v)
				return v.getTrees(true, v.e.t)
			}
		}
		tmpl := `{
			num := uint64(len(::.{{.name}}))
			subLeaves := make([]*ssz.Node, num)
			if num > {{.num}} {
				err = ssz.ErrIncorrectListSize
				return nil, err
			}
			for i := uint64(0); i < num; i++ {
				n, err := ::.{{.name}}[i].GetTree()
				if err != nil {
					return nil, err
				}
				subLeaves[i] = n
			}
			tmp, err = ssz.TreeFromNodesWithMixin(subLeaves, {{.num}})
			if err != nil {
				return nil, err
			}
		}`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"num":  v.m,
		})

	default:
		panic(fmt.Errorf("hash not implemented for type %s", v.t.String()))
	}
}

func (v *Value) getTreeContainer(start bool) string {
	if !start {
		return fmt.Sprintf("tmp, err = ::.%s.GetTree()\n if err != nil {\n return nil, err\n}", v.name)
	}

	numLeaves := nextPowerOfTwo(uint64(len(v.o)))
	out := []string{}
	for indx, i := range v.o {
		str := fmt.Sprintf("// Field (%d) '%s'\n%s\nleaves[%d] = tmp\n", indx, i.name, i.getTree(), indx)
		out = append(out, str)
	}
	// Empty leaves
	emptyLeaves := ""
	if numLeaves-uint(len(v.o)) > 0 {
		emptyLeaves = fmt.Sprintf("for i := 0; i < %d; i++ {\nleaves[i+%d] = ssz.EmptyLeaf()\n}", numLeaves-uint(len(v.o)), len(v.o))
	}

	tmpl := `leaves := make([]*ssz.Node, {{.numLeaves}})
	var tmp *ssz.Node
	var err error

	{{.fields}}
	{{.emptyLeaves}}

	return ssz.TreeFromNodes(leaves)`

	return execTmpl(tmpl, map[string]interface{}{
		"numLeaves":   numLeaves,
		"fields":      strings.Join(out, "\n"),
		"emptyLeaves": emptyLeaves,
	})
}

func nextPowerOfTwo(v uint64) uint {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return uint(v)
}
