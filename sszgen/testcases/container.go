package testcases

//go:generate go run ../main.go --path container.go

type Vec struct {
	Values []uint64 `ssz-size:"6"`
}

type Vec2 struct {
	Values2 []uint32 `ssz-max:"100"`
}
