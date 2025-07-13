package generator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

type tokenState int

const (
	tsBegin tokenState = iota
	tsLabel
	tsValue
	tsCloseTick
)

func GetSSZTags(tag string) (map[string]string, error) {
	var lastErr error
	accumulateError := func(_ *scanner.Scanner, msg string) {
		lastErr = errors.New(msg)
	}

	sr := strings.NewReader(tag)
	sc := scanner.Scanner{}
	sc.Init(sr)
	sc.Filename = "tag"
	sc.Mode ^= scanner.ScanRawStrings
	sc.Error = accumulateError

	var labelStr string
	var state tokenState
	tags := make(map[string]string)
	for tok := sc.Scan(); tok != scanner.EOF; tok = sc.Scan() {
		if lastErr != nil {
			return nil, fmt.Errorf("GetSSZTags failed: token scanner error = %s", lastErr)
		}
		if state == tsCloseTick {
			return nil, errors.New("GetSSZTags failed: undefined behavior when scanning beyond the end of the tag")
		}
		txt := sc.TokenText()
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
	return tags, nil
}

// cannot compare untyped nil to typed nil
// this value gives us a nil with type of *int
// to compare to ssz-size = '?' values
var nilInt *SSZLength

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

var errDimNotFound = fmt.Errorf("no ssz-size or ssz-max tags found for element")

func extractSSZDimensions(tag string) ([]*SSZDimension, error) {
	// parse the ssz-max and ssz-size key/value pairs out of the tag
	tags, err := GetSSZTags(tag)
	if err != nil {
		return nil, err
	}
	sszSizes, sizeDefined := tags["ssz-size"]
	sszMax, maxDefined := tags["ssz-max"]
	if !sizeDefined && !maxDefined {
		return nil, errDimNotFound
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
			return nil, fmt.Errorf("at dimension %d both ssz-size and ssz-max had a '?' value. For each dimension, either ssz-size or ssz-max must have a value. Ex: 'ssz-size:\"?,32\" ssz-max:\"100\" defines a List with 100 element limit, containing 32 byte fixed-sized vectors", i)
		}
		switch szi {
		case "?", "":
			if mxi == "?" || mxi == "" {
				return nil, fmt.Errorf("no numeric ssz-size or ssz-max tag for value at dimesion %d", i)
			}
			m, err := newSSZLength(mxi)
			if err != nil {
				return nil, fmt.Errorf("atoi failed on value %s for ssz-max at dimension %d err=%s", mxi, i, err)
			}
			dims[i] = &SSZDimension{
				isBitlist:  isbl,
				ListLength: m,
			}
		default: // szi is not empty or "?"
			s, err := newSSZLength(szi)
			if err != nil {
				return nil, fmt.Errorf("atoi failed on value %s for ssz-size at dimension %d, err=%s", szi, i, err)
			}
			dims[i] = &SSZDimension{
				isBitlist:    isbl,
				VectorLength: s,
			}
			continue
		}
	}
	return dims, nil
}

type SSZLength struct {
	Size *uint64
	Dyn  string
}

func (s *SSZLength) EncodeTemplate() string {
	if s == nil {
		return "0 // nil SSZLength check"
	}
	if s.Size != nil {
		return strconv.FormatUint(*s.Size, 10)
	}
	if s.Dyn != "" {
		return s.Dyn
	}
	panic("SSZLength has no size or dynamic value defined")
}

func newSSZLengthNum(val uint64) *SSZLength {
	return &SSZLength{Size: &val}
}

func newSSZLength(val string) (*SSZLength, error) {
	if val == "" {
		return &SSZLength{}, fmt.Errorf("newSSZLength called with empty value")
	}

	// Check if the first character is a digit
	if unicode.IsDigit(rune(val[0])) {
		// Try to parse as integer
		size, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		sizeU64 := uint64(size)
		return &SSZLength{Size: &sizeU64}, nil
	}

	// Otherwise, treat as dynamic string - must be valid Go identifier
	if !isValidGoIdentifier(val) {
		return nil, fmt.Errorf("invalid Go identifier: %s", val)
	}

	return &SSZLength{Dyn: val}, nil
}

// isValidGoIdentifier checks if a string is a valid Go identifier
func isValidGoIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// First character must be a letter or underscore
	first := rune(s[0])
	if !unicode.IsLetter(first) && first != '_' {
		return false
	}

	// Remaining characters must be letters, digits, or underscores
	for _, r := range s[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}

	return true
}

type SSZDimension struct {
	VectorLength *SSZLength
	ListLength   *SSZLength
	isBitlist    bool
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

func trimTagQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}
