package main

import (
	"testing"

)

func TestTokens(t *testing.T) {
	testTag := "`protobuf:\"bytes,2004,rep,name=historical_roots,json=historicalRoots,proto3\" json:\"historical_roots,omitempty\" ssz-max:\"16777216\" ssz-size:\"?,32\"`"
	tp := &TagParser{}
	tp.Init(testTag)
	tags := tp.GetSSZTags()
	sszSize, ok := tags["ssz-size"]
	if !ok {
		t.Errorf("ssz-size tag not present")
	}
	expectedSize := "?,32"
	if sszSize != expectedSize {
		t.Errorf("expected ssz-size value '%s', got '%s'", expectedSize, sszSize)
	}
	sszMax, ok := tags["ssz-max"]
	if !ok {
		t.Errorf("ssz-max tag not present")
	}
	expectedMax := "16777216"
	if sszMax != expectedMax {
		t.Errorf("expected ssz-max value '%s', got '%s'", expectedMax, sszMax)
	}
}

func TestFullTag(t *testing.T) {
	tag := "`protobuf:\"bytes,1002,opt,name=genesis_validators_root,json=genesisValidatorsRoot,proto3\" json:\"genesis_validators_root,omitempty\" ssz-size:\"32\"`"
	_, err := extractSSZDimensions(tag)
	if err != nil {
		t.Errorf("Unexpected error calling extractSSZDimensions: %v", err)
	}
}

func TestListOfVector(t *testing.T) {
	tag := "`protobuf:\"bytes,2004,rep,name=historical_roots,json=historicalRoots,proto3\" json:\"historical_roots,omitempty\" ssz-max:\"16777216\" ssz-size:\"?,32\"`"
	_, err := extractSSZDimensions(tag)
	if err != nil {
		t.Errorf("Unexpected error calling extractSSZDimensions: %v", err)
	}
}

func TestWildcardSSZSize(t *testing.T)  {
	tag := "`ssz-max:\"16777216\" ssz-size:\"?,32\"`"
	dims, err := extractSSZDimensions(tag)
	if err != nil {
		t.Errorf("Unexpected error calling extractSSZDimensions: %v", err)
	}
	expectedDims := 2
	if len(dims) != expectedDims {
		t.Errorf("expected %d dimensions from ssz tags, got %d", expectedDims, len(dims))
	}
	if !dims[0].IsList() {
		t.Errorf("Expected the first dimension to be a list")
	}
	if dims[0].IsVector() {
		t.Errorf("Expected the first dimension to not be a vector")
	}
	if dims[0].ListLen() != 16777216 {
		t.Errorf("Expected max size of list to be %d, got %d", 16777216, dims[0].ListLen())
	}
	if !dims[1].IsVector() {
		t.Errorf("Expected the first dimension to be a vector")
	}
	if dims[1].IsList() {
		t.Errorf("Expected the second dimension to not be a list")
	}
	if dims[1].VectorLen() != 32 {
		t.Errorf("Expected size of vector to be %d, got %d", 32, dims[1].VectorLen())
	}
}

func TestListOfList(t *testing.T) {
	tag := "`protobuf:\"bytes,14,rep,name=transactions,proto3\" json:\"transactions,omitempty\" ssz-max:\"1048576,1073741824\" ssz-size:\"?,?\"`"
	dims, err := extractSSZDimensions(tag)
	if err != nil {
		t.Errorf("Unexpected error calling extractSSZDimensions: %v", err)
	}
	expectedDims := 2
	if len(dims) != expectedDims {
		t.Errorf("expected %d dimensions from ssz tags, got %d", expectedDims, len(dims))
	}
	if !dims[0].IsList() {
		t.Errorf("Expected both dimensions to be lists, but the first dimension is not")
	}
	if !dims[1].IsList() {
		t.Errorf("Expected both dimensions to be lists, but the second dimension is not")
	}
	if dims[0].IsVector() {
		t.Errorf("Expected neither dimension to be vector, but the first dimension is")
	}
	if dims[1].IsVector() {
		t.Errorf("Expected neither dimension to be vector, but the second dimension is")
	}
	if dims[0].ListLen() != 1048576 {
		t.Errorf("Expected ssz-max of first dimension to be %d, got %d", 1048576 , dims[0].ListLen())
	}
	if dims[1].ListLen() != 1073741824 {
		t.Errorf("Expected ssz-max of first dimension to be %d, got %d", 1073741824, dims[1].ListLen())
	}
}
