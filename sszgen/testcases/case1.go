package testcases

//go:generate go run ../main.go --path case1.go --exclude-objs Bytes

type Bytes []byte

type Case1A struct {
	Foo Bytes `ssz-max:"2048"`
}

type Case2A struct {
	Bar Bytes `ssz-max:"32"`
}
