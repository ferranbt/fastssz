package testcases

import "github.com/ferranbt/fastssz/sszgen/testcases/other2"

//go:generate go run ../main.go --path issue_164.go --include ./other2

type Issue64 struct {
	// Encoding generated will be for slice and not array
	FeeRecipientAddress [other2.Size]byte
}
