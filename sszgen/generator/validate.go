package generator

func validateBytesArray(name string, size uint64, fixed bool) string {
	// for variable size values, we want to ensure it doesn't exceed max size bound
	cmp := ">"
	if fixed {
		cmp = "!="
	}

	tmpl := `if size := len(::.{{.name}}); size {{.cmp}} {{.size}} {
			err = ssz.ErrBytesLengthFn("--.{{.name}}", size, {{.size}})
			return
		}
	`
	return execTmpl(tmpl, map[string]interface{}{
		"cmp":  cmp,
		"name": name,
		"size": size,
	})
}

func (v *Value) validate() string {
	switch obj := v.typ.(type) {
	case *BitList:
		return validateBytesArray(v.name, obj.Size, false)
	case *Bytes:
		if obj.IsList {
			// for lists of bytes, we need to validate the size of the buffer
			return validateBytesArray(v.name, obj.Size, false)
		}
		if obj.IsGoDyn {
			// always validate dynamic bytes
			return validateBytesArray(v.name, obj.Size, true)
		}
		return ""
	case *Vector:
		if !obj.IsDyn {
			return ""
		}

		// We only have vectors for [][]byte roots
		tmpl := `if size := len(::.{{.name}}); size != {{.size}} {
			err = ssz.ErrVectorLengthFn("--.{{.name}}", size, {{.size}})
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"size": obj.Size,
		})

	case *List:
		tmpl := `if size := len(::.{{.name}}); size > {{.size}} {
			err = ssz.ErrListTooBigFn("--.{{.name}}", size, {{.size}})
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"size": obj.MaxSize,
		})

	default:
		return ""
	}
}
