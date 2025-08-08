package generator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

// Vector -> fixed size
// List -> variable size

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
			m, err := parseSize(mxi)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ssz-size at dimension %d err=%s", i, err)
			}
			dims[i] = &SSZDimension{
				isBitlist:  isbl,
				ListLength: m,
			}
		default: // szi is not empty or "?"
			s, err := parseSize(szi)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ssz-size at dimension %d, err=%s", i, err)
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

func parseSize(tagValue string) (*Size, error) {
	if strings.HasPrefix(tagValue, "var(") {
		// variable declaration
		tagValue = strings.TrimPrefix(tagValue, "var(")
		tagValue = strings.TrimSuffix(tagValue, ")")

		if tagValue == "" {
			return nil, fmt.Errorf("variable size tag cannot be empty")
		}
		return &Size{
			Size:    0,
			VarSize: tagValue,
		}, nil
	}

	m, err := strconv.Atoi(tagValue)
	if err != nil {
		return nil, fmt.Errorf("atoi failed on value %s err=%s", tagValue, err)
	}
	return &Size{
		Size:    uint64(m),
		VarSize: "",
	}, nil
}

type SSZDimension struct {
	VectorLength *Size
	ListLength   *Size
	isBitlist    bool
}

func (dim *SSZDimension) Type() string {
	if dim.IsVector() {
		return "vector"
	}
	if dim.IsList() {
		return "list"
	}
	if dim.IsBitlist() {
		return "bitlist"
	}
	return "undefined"
}

func (dim *SSZDimension) IsVector() bool {
	return dim.VectorLength != nil
}

func (dim *SSZDimension) IsList() bool {
	return dim.ListLength != nil
}

func (dim *SSZDimension) IsBitlist() bool {
	return dim.isBitlist
}

func (dim *SSZDimension) ListLen() Size {
	return *dim.ListLength
}

func (dim *SSZDimension) VectorLen() Size {
	return *dim.VectorLength
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
