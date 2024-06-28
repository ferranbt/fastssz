package testcases

//go:generate go run ../main.go --path issue_165.go --include ./other

import "github.com/ferranbt/fastssz/sszgen/testcases/other"

type Issue165 struct {
	A other.Case4Bytes
}
