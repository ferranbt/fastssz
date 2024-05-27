package testcases

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUint(t *testing.T) {
	s := Uints{
		Uint8:  Uint8(123),
		Uint16: Uint16(12345),
		Uint32: Uint32(1234567890),
		Uint64: Uint64(123456789000),
	}
	expectedHash := [32]byte{
		0xea, 0xfc, 0xf7, 0xa2, 0x41, 0x8, 0x51, 0xa2,
		0xa0, 0xb0, 0x23, 0x68, 0xff, 0x4, 0x44, 0xbd,
		0x24, 0xc9, 0x9b, 0xff, 0xe7, 0x81, 0xca, 0x49,
		0xb6, 0xf7, 0xd4, 0x99, 0x28, 0xf3, 0xee, 0xeb,
	}

	bytes, err := s.MarshalSSZ()
	assert.NoError(t, err)

	var s2 Uints
	err = s2.UnmarshalSSZ(bytes)
	assert.NoError(t, err)

	assert.Equal(t, s, s2)

	h, err := s.HashTreeRoot()
	assert.NoError(t, err)

	assert.Equal(t, h, expectedHash)
}
