package testcases

// Data152 is a byte array with 48 elements
//
//go:generate go run ../main.go --path pr_152.go
type Data152 [48]byte

// PR1512 is a struct with a Data152 field
type PR1512 struct {
	D []Data152 `ssz-max:"32"`
}
