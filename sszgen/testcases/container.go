package testcases

//go:generate go run ../main.go --path container.go

type Vec struct {
	Values []uint64 `ssz-size:"6"`
}
