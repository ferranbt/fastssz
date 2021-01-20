package main

func (v *Value) validate() string {
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
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"cmp":  cmp,
			"name": v.name,
			"size": size,
		})

	case TypeVector:
		if v.c {
			return ""
		}
		// We only have vectors for [][]byte roots
		tmpl := `if len(::.{{.name}}) != {{.size}} {
			err = ssz.ErrVectorLength
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"size": v.s,
		})

	case TypeList:
		tmpl := `if len(::.{{.name}}) > {{.size}} {
			err = ssz.ErrListTooBig
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"size": v.s,
		})

	default:
		return ""
	}
}
