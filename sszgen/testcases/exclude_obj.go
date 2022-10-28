package testcases

//go:generate go run ../main.go --path exclude_obj.go --exclude-objs Bytes

type Bytes []byte

type Case1A struct {
	Foo Bytes `ssz-max:"2048"`
}

type Case1B struct {
	Bar Bytes `ssz-max:"32"`
}
