package testcases

//go:generate go run ../main.go --path case5.go --exclude-objs Case5Bytes,Case5Roots

type Case5Bytes []byte

type Case5Roots [][]byte

type Case5A struct {
	A [][]byte     `ssz-size:"2,2"`
	B []Case5Bytes `ssz-size:"2,2"`
	C Case5Roots   `ssz-size:"2,2"`
}
