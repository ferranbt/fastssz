// Code generated by fastssz. DO NOT EDIT.
// Hash: 8477a1e3c842f2c10c5c5006f021156a5dfad300568442f0881fb828a8952f25
// Version: 0.1.3
package testcases

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the Issue159[B] object
func (i *Issue159[B]) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(i)
}

// MarshalSSZTo ssz marshals the Issue159[B] object to a target array
func (i *Issue159[B]) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Data'
	dst = append(dst, i.Data[:]...)

	return
}

// UnmarshalSSZ ssz unmarshals the Issue159[B] object
func (i *Issue159[B]) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 48 {
		return ssz.ErrSize
	}

	// Field (0) 'Data'
	copy(i.Data[:], buf[0:48])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Issue159[B] object
func (i *Issue159[B]) SizeSSZ() (size int) {
	size = 48
	return
}

// HashTreeRoot ssz hashes the Issue159[B] object
func (i *Issue159[B]) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(i)
}

// HashTreeRootWith ssz hashes the Issue159[B] object with a hasher
func (i *Issue159[B]) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'Data'
	hh.PutBytes(i.Data[:])

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Issue159[B] object
func (i *Issue159[B]) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(i)
}
