package testcases

// Issue153 is a struct with a Data152 field
//
//go:generate go run ../main.go --path issue_153.go -include pr_152.go
type Issue153 struct {
	Value1 [32]byte `ssz-size:"32"`
	Value2 [48]byte // will be treated as a non basic type
	Value  Data152  `ssz-size:"48"`
}
