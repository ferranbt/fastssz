package testcases

//go:generate go run ../main.go --path case7.go

type Case7 struct {
	BlobKzgs [][]byte `ssz-size:"?,48" ssz-max:"16"`
}
