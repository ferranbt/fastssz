// Code generated by fastssz. DO NOT EDIT.
// Hash: 1746a8d4be0b81f7db97c9708cc4a5e299e38fac75a97bba4e67f854c33d984b
// Version: 0.1.3
package testcases

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the Issue156Aux object
func (i *Issue156Aux) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(i)
}

// MarshalSSZTo ssz marshals the Issue156Aux object to a target array
func (i *Issue156Aux) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'A'
	dst = ssz.MarshalUint64(dst, i.A)

	return
}

// UnmarshalSSZ ssz unmarshals the Issue156Aux object
func (i *Issue156Aux) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 8 {
		return ssz.ErrSize
	}

	// Field (0) 'A'
	i.A = ssz.UnmarshallUint64(buf[0:8])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Issue156Aux object
func (i *Issue156Aux) SizeSSZ() (size int) {
	size = 8
	return
}

// HashTreeRoot ssz hashes the Issue156Aux object
func (i *Issue156Aux) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(i)
}

// HashTreeRootWith ssz hashes the Issue156Aux object with a hasher
func (i *Issue156Aux) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'A'
	hh.PutUint64(i.A)

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Issue156Aux object
func (i *Issue156Aux) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(i)
}

// MarshalSSZ ssz marshals the Issue156 object
func (i *Issue156) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(i)
}

// MarshalSSZTo ssz marshals the Issue156 object to a target array
func (i *Issue156) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'A'
	dst = append(dst, i.A[:]...)

	// Field (1) 'A2'
	dst = append(dst, i.A2[:]...)

	// Field (2) 'A3'
	dst = append(dst, i.A3[:]...)

	// Field (3) 'A4'
	if size := len(i.A4); size != 32 {
		err = ssz.ErrBytesLengthFn("Issue156.A4", size, 32)
		return
	}
	dst = append(dst, i.A4...)

	return
}

// UnmarshalSSZ ssz unmarshals the Issue156 object
func (i *Issue156) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 128 {
		return ssz.ErrSize
	}

	// Field (0) 'A'
	copy(i.A[:], buf[0:32])

	// Field (1) 'A2'
	copy(i.A2[:], buf[32:64])

	// Field (2) 'A3'
	copy(i.A3[:], buf[64:96])

	// Field (3) 'A4'
	if cap(i.A4) == 0 {
		i.A4 = make([]byte, 0, len(buf[96:128]))
	}
	i.A4 = append(i.A4, buf[96:128]...)

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Issue156 object
func (i *Issue156) SizeSSZ() (size int) {
	size = 128
	return
}

// HashTreeRoot ssz hashes the Issue156 object
func (i *Issue156) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(i)
}

// HashTreeRootWith ssz hashes the Issue156 object with a hasher
func (i *Issue156) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'A'
	hh.PutBytes(i.A[:])

	// Field (1) 'A2'
	hh.PutBytes(i.A2[:])

	// Field (2) 'A3'
	hh.PutBytes(i.A3[:])

	// Field (3) 'A4'
	if size := len(i.A4); size != 32 {
		err = ssz.ErrBytesLengthFn("Issue156.A4", size, 32)
		return
	}
	hh.PutBytes(i.A4)

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Issue156 object
func (i *Issue156) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(i)
}
