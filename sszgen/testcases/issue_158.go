package testcases

import (
	"github.com/ferranbt/fastssz/sszgen/testcases/other"
)

//go:generate go run ../main.go --path issue_158.go --include ./other

type Int = other.Case3B

type Issue158 struct {
	Field *Int
}
