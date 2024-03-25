package testcases

//go:generate go run ../main.go --path list.go

type BytesWrapper struct {
	Bytes []byte `ssz-size:"48"`
}

type ListC struct {
	Elems []BytesWrapper `ssz-max:"32"`
}

type ListP struct {
	Elems []*BytesWrapper `ssz-max:"32"`
}
