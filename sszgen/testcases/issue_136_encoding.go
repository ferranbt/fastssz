// Code generated by fastssz. DO NOT EDIT.
// Hash: 38f3ed802150f463fcd7ca9cb33a94e40338ff06d004ed18896fa1e6ef81401a
// Version: 0.1.4
package testcases

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the Issue136 object
func (i *Issue136) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(i)
}

// MarshalSSZTo ssz marshals the Issue136 object to a target array
func (i *Issue136) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'C'
	if dst, err = i.C.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the Issue136 object
func (i *Issue136) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 0 {
		return ssz.ErrSize
	}

	// Field (0) 'C'
	if err = i.C.UnmarshalSSZ(buf[0:0]); err != nil {
		return err
	}

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Issue136 object
func (i *Issue136) SizeSSZ() (size int) {
	size = 0
	return
}

// HashTreeRoot ssz hashes the Issue136 object
func (i *Issue136) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(i)
}

// HashTreeRootWith ssz hashes the Issue136 object with a hasher
func (i *Issue136) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'C'
	if err = i.C.HashTreeRootWith(hh); err != nil {
		return
	}

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Issue136 object
func (i *Issue136) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(i)
}
