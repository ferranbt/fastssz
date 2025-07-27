package testcases

//go:generate go run ../main.go --path integration_uint.go --exclude-objs Data

type IntegrationUint struct {
	A uint8
	B uint16
	C uint32
	D uint64

	A1 []uint8  `ssz-max:"400"`
	A2 []uint16 `ssz-max:"400"`
	A3 []uint32 `ssz-max:"400"`
	A4 []uint64 `ssz-max:"400"`
}
