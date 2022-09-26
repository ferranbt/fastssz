package testcases

//go:generate go run ../main.go --path case6.go

const Case6Size = 32

type Case6 struct {
	A [Case6Size]byte `ssz-size:"32"`
}
