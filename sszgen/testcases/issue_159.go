package testcases

//go:generate go run ../main.go --path issue_159.go -objs Issue159

type Issue22 [96]byte

// Issue159 is a struct with a Data field
type Issue159[B [48]byte] struct {
	Data B `ssz-size:"48"`
}
