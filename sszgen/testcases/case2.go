package testcases

//go:generate go run ../main.go --path case2.go

type Case2A struct {
	A uint64
}

type Case2B struct {
	Case2A

	B uint64
}
