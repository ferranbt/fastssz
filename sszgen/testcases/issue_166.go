package testcases

//go:generate go run ../main.go --path issue_166.go --include ./other

import "github.com/ferranbt/fastssz/sszgen/testcases/other"

type Case4Alias = other.Case4Bytes

type Issue165 struct {
	A other.Case4Bytes
	B Case4Alias
}
