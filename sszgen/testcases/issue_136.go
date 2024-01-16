package testcases

import "github.com/ferranbt/fastssz/sszgen/testcases/other"

//go:generate go run ../main.go --path issue_136.go --include ./other

type Issue136 struct {
	C other.Case3B
}
