package tests

import (
	"bytes"
	"errors"
	"math"

	"github.com/minio/sha256-simd"
)

type Hash []byte

type Metadata struct {
	Version    uint8
	CodeHash   Hash `ssz-size:"32"`
	CodeLength uint16
}

type Chunk struct {
	FIO  uint8
	Code []byte `ssz-size:"32"` // Last chunk is right-padded with zeros
}

type CodeTrieSmall struct {
	Metadata *Metadata
	Chunks   []*Chunk `ssz-max:"4"`
}

// Verifies a single merkle branch for the Metadata schema
func VerifyProof(root []byte, proof [][]byte, leaf []byte, index int) (bool, error) {
	if len(proof) != getPathLength(index) {
		return false, errors.New("Invalid proof length")
	}

	node := leaf[:]
	tmp := make([]byte, 64)
	for i, h := range proof {
		if getPosAtLevel(index, i) {
			copy(tmp[:32], h[:])
			copy(tmp[32:], node[:])
			node = hash(tmp)
		} else {
			copy(tmp[:32], node[:])
			copy(tmp[32:], h[:])
			node = hash(tmp)
		}
	}

	return bytes.Equal(root, node), nil
}

// Returns the position (i.e. false for left, true for right)
// of an index at a given level.
// Level 0 is the actual index's level, Level 1 is the position
// of the parent, etc.
func getPosAtLevel(index int, level int) bool {
	return (index & (1 << level)) > 0
}

// Return the length of the path to a node represented by its generalized index
func getPathLength(index int) int {
	return int(math.Log2(float64(index)))
}

func hash(data []byte) []byte {
	res := sha256.Sum256(data)
	return res[:]
}
