package testcases

import "github.com/ferranbt/fastssz/sszgen/testcases/other"

//go:generate go run ../main.go --path issue_158.go --include ./other

type MyType struct {
	MyField0 []byte
	MyField1 []other.Case4Bytes `ssz-max:"2,2"`
}
