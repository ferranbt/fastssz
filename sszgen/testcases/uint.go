package testcases

//go:generate go run ../main.go --path uint.go

type Uint8 uint8
type Uint16 uint16
type Uint32 uint32
type Uint64 uint64

type Uints struct {
	Uint8  Uint8
	Uint16 Uint16
	Uint32 Uint32
	Uint64 Uint64
}
