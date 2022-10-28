package testcases

import (
	"github.com/ferranbt/fastssz/sszgen/testcases/resolve-packages/other"
)

//go:generate go run ../../main.go --path single.go

type Case3B struct {
}

type Case3A struct {
	A Case3B
	B *Case3B
	C other.Case3B
	D *other.Case3B
}
