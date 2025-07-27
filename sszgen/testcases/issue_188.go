package testcases

//go:generate go run ../main.go --path issue_188.go

type Issue188 struct {
	Name, Address, notSet []byte `ssz-size:"32"`
}
