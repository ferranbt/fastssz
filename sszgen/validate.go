package main

func (v *Value) validate() string {
	switch v.t {
	case TypeBytes:
		if v.c {
			return ""
		}
		// []byte are always fixed
		size := v.s
		tmpl := `if len(::.{{.name}}) != {{.size}} {
			err = ssz.ErrBytesLength
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
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
