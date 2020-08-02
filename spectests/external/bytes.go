package external

import ssz "github.com/ferranbt/fastssz"

// Signature is a 96 bytes array external reference
type Signature [96]byte

// Bytes is a dynamic array of bytes
type Bytes []byte

// DynamicBytes is a dynamic reference
type DynamicBytes []byte

// SizeSSZ implements the fastssz Marshaler interface
func (d *DynamicBytes) SizeSSZ() (size int) {
	return len(*d)
}

// MarshalSSZTo implements the fastssz Marshaler interface
func (d *DynamicBytes) MarshalSSZTo(buf []byte) ([]byte, error) {
	if len(*d) > 256 {
		return nil, ssz.ErrBytesLength
	}
	return append(buf, *d...), nil
}

// HashTreeRootWith implements the fastssz HashRoot interface
func (d *DynamicBytes) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	if len(*d) > 256 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(*d)
	return
}

// UnmarshalSSZ implements the fastssz Unmarshaler interface
func (d *DynamicBytes) UnmarshalSSZ(buf []byte) error {
	if len(buf) > 256 {
		return ssz.ErrBytesLength
	}
	(*d) = append(*d, buf...)
	return nil
}
