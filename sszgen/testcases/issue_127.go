package testcases

//go:generate go run ../main.go --path issue_127.go --exclude-objs Data

type Data []byte

type Obj2 struct {
	T1 []Data `ssz-max:"1024,256" ssz-size:"?,?"`
}
