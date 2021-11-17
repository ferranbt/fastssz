package main

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

type tokenState int

const (
	tsBegin tokenState = iota
	tsLabel
	tsValue
	tsCloseTick
)

type TagParser struct {
	sc scanner.Scanner
	buffer string
}

func (tp *TagParser) Init(tag string) {
	sr := strings.NewReader(tag)
	tp.sc = scanner.Scanner{}
	tp.sc.Init(sr)
	tp.sc.Filename = "tag"
	tp.sc.Mode ^= scanner.ScanRawStrings
}

func (tp TagParser) GetSSZTags() map[string]string {
	var labelStr string
	var state tokenState
	tags := make(map[string]string)
	for tok := tp.sc.Scan(); tok != scanner.EOF; tok = tp.sc.Scan() {
		if state == tsCloseTick {
			panic("undefined behavior when scanning beyond the end of the tag")
		}
		txt := tp.sc.TokenText()
		switch txt {
		case "`":
			if state == tsLabel {
				state = tsCloseTick
				continue
			}
			if state == tsBegin {
				state = tsLabel
				continue
			}
		case ":":
			if state == tsLabel {
				state = tsValue
				continue
			}
		case "\"":
			continue
		default:
			if state == tsValue {
				tags[labelStr] = trimTagQuotes(txt)
				state = tsLabel
				labelStr = ""
				continue
			}
			if state == tsLabel {
				labelStr += txt
				continue
			}
		}
	}
	return tags
}

// cannot compare untyped nil to typed nil
// this value gives us a nil with type of *int
// to compare to ssz-size = '?' values
var nilInt *int

// handle tag structured like 'ssz:"bitlist"'
// this is not used in prysm but needs to be supported for fastssz tests
func isBitList(tags map[string]string) bool {
	for k, v := range tags {
		if k == "ssz" {
			parts := strings.Split(v, ",")
			for _, p := range parts {
				if p == "bitlist" {
					return true
				}
			}
		}
	}
	return false
}

func extractSSZDimensions(tag string) ([]*SSZDimension, error) {
	tp := &TagParser{}
	tp.Init(tag)
	// parse the ssz-max and ssz-size key/value pairs out of the tag
	tags := tp.GetSSZTags()
	sszSizes, sizeDefined := tags["ssz-size"]
	sszMax, maxDefined := tags["ssz-max"]
	if !sizeDefined && !maxDefined {
		return nil, fmt.Errorf("No ssz-size or ssz-max tags found for element. tag=%s", tag)
	}

	// split each tag by ",". each position in the csv represents a dimension of an n-dimensional array
	sizeSplit := strings.Split(sszSizes, ",")
	maxSplit := strings.Split(sszMax, ",")
	// find the largest of the two dimensions. for backward compat we'll be permissive and let them be uneven
	ndims := len(sizeSplit)
	if len(maxSplit) > len(sizeSplit) {
		ndims = len(maxSplit)
	}
	dims := make([]*SSZDimension, ndims)
	for i := 0; i < ndims; i++ {
		isbl := false
		// bitlist can only be the inner-most element by definition
		if i == ndims-1 {
			if isBitList(tags) {
				isbl = true
			}
		}
		var szi, mxi string
		if len(sizeSplit) > i {
			szi = sizeSplit[i]
		}
		if len(maxSplit) > i {
			mxi = maxSplit[i]
		}
		if szi == "?" && mxi == "?" {
			return nil, fmt.Errorf("At dimension %d both ssz-size and ssz-max had a '?' value, tag=%s", i, tag)
		}
		switch szi {
		case "?", "":
			if mxi == "?" || mxi == "" {
				return nil, fmt.Errorf("no numeric ssz-size or ssz-max tag for value at dimesion %d, tag=%s", i, tag)
			}
			m, err := strconv.Atoi(mxi)
			if err != nil {
				return nil, fmt.Errorf("atoi failed on value %s for ssz-max at dimension %d, tag=%s. err=%s", mxi, i, tag, err)
			}
			dims[i] = &SSZDimension{
				isBitlist: isbl,
				ListLength:  &m,
			}
		default: // szi is not empty or "?"
			s, err := strconv.Atoi(szi)
			if err != nil {
				return nil, fmt.Errorf("atoi failed on value %s for ssz-size at dimension %d, tag=%s. err=%s", szi, i, tag, err)
			}
			dims[i] = &SSZDimension{
				isBitlist: isbl,
				VectorLength:  &s,
			}
			continue
		}
	}
	return dims, nil
}

type SSZDimension struct {
	VectorLength *int
	ListLength *int
	isBitlist bool
}

func (dim *SSZDimension) IsVector() bool {
	return dim.VectorLength != nilInt
}

func (dim *SSZDimension) IsList() bool {
	return dim.ListLength != nilInt
}

func (dim *SSZDimension) IsBitlist() bool {
	return dim.isBitlist
}

func (dim *SSZDimension) ListLen() int {
	return *dim.ListLength
}

func (dim *SSZDimension) VectorLen() int {
	return *dim.VectorLength
}

// ValueType returns a Type enum to be used in the construction of a fastssz Value type
func (dim *SSZDimension) ValueType() Type {
	if dim.IsVector() {
		return TypeVector
	}
	if dim.IsList() {
		return TypeList
	}
	return TypeUndefined
}

// ValueType returns ssz-max or ssz-size, to be used in the construction of a fastssz Value type
func (dim *SSZDimension) ValueLen() uint64 {
	if dim.IsList() {
		return uint64(dim.ListLen())
	}
	if dim.IsVector() {
		return uint64(dim.VectorLen())
	}
	return 0
}

func trimTagQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}
