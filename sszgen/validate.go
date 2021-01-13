package main

func (v *Value) validate(returnStatement string) string {
	switch v.t {
	case TypeBitList, TypeBytes:
		cmp := "!="
		if v.t == TypeBitList {
			cmp = ">"
		}
		if v.c {
			return ""
		}
		// fixed []byte
		size := v.s
		if size == 0 {
			// dynamic []byte
			size = v.m
			cmp = ">"
		}

		tmpl := `if len(::.{{.name}}) {{.cmp}} {{.size}} {
			err = ssz.ErrBytesLength
			{{.return}}
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"cmp":    cmp,
			"name":   v.name,
			"size":   size,
			"return": returnStatement,
		})

	case TypeVector:
		if v.c {
			return ""
		}
		// We only have vectors for [][]byte roots
		tmpl := `if len(::.{{.name}}) != {{.size}} {
			err = ssz.ErrVectorLength
			{{.return}}
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name":   v.name,
			"size":   v.s,
			"return": returnStatement,
		})

	case TypeList:
		tmpl := `if len(::.{{.name}}) > {{.size}} {
			err = ssz.ErrListTooBig
			{{.return}}
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name":   v.name,
			"size":   v.s,
			"return": returnStatement,
		})

	default:
		return ""
	}
}
