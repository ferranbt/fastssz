package testcases

type Issue156Aux struct {
	A uint64
}

//go:generate go run ../main.go --path issue_156.go
type Issue156 struct {
	A  [32]byte `ssz-size:"32"`
	A2 [32]byte
	A3 [32]byte `json:"a3"`
	A4 []byte   `json:"a4" ssz-size:"32"`
}
