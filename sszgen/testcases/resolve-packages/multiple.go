package testcases

import (
	"github.com/ferranbt/fastssz/sszgen/testcases/resolve-packages/other"
	alias "github.com/ferranbt/fastssz/sszgen/testcases/resolve-packages/other2"
)

//go:generate go run ../../main.go --include ./other,./other2 --path multiple.go

type Case4 struct {
	A other.Case4Interface  `ssz-size:"96"`
	B *other.Case4Interface `ssz-size:"96"`
	C alias.Case4Slot
	D other.Case4Bytes          `ssz-size:"96"`
	E other.Case4FixedSignature `ssz-size:"96"`
}
