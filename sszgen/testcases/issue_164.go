package testcases

//go:generate go run ../main.go --path issue_164.go

const Issue164ElemSize = 20

type Issue164Elem [Issue164ElemSize]byte

type Issue164 struct {
	A Issue164Elem
}
