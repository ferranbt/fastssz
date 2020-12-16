package ssz

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestTreeFromChunks(t *testing.T) {
	chunks := [][]byte{
		{0x01, 0x01},
		{0x02, 0x02},
		{0x03, 0x03},
		{0x00, 0x00},
	}

	r, err := TreeFromChunks(chunks)
	if err != nil {
		t.Errorf("Failed to construct tree: %v\n", err)
	}
	for i := 4; i < 8; i++ {
		l, err := r.Get(i)
		if err != nil {
			t.Errorf("Failed getting leaf: %v\n", err)
		}
		if !bytes.Equal(l.value, chunks[i-4]) {
			t.Errorf("Incorrect leaf at index %d\n", i)
		}
	}
}

func TestHashTree(t *testing.T) {
	expectedRootHex := "6621edd5d039d27d1ced186d57691a04903ac79b389187c2d453b5d3cd65180e"
	expectedRoot, err := hex.DecodeString(expectedRootHex)
	if err != nil {
		t.Errorf("Failed to decode hex string\n")
	}

	chunks := [][]byte{
		{0x01, 0x01},
		{0x02, 0x02},
		{0x03, 0x03},
		{0x00, 0x00},
	}

	r, err := TreeFromChunks(chunks)
	if err != nil {
		t.Errorf("Failed to construct tree: %v\n", err)
	}

	h := r.Hash()
	if !bytes.Equal(h, expectedRoot) {
		t.Errorf("Computed hash is incorrect. Expected %s, got %s\n", expectedRootHex, hex.EncodeToString(h))
	}
}

func TestProve(t *testing.T) {
	expectedProofHex := []string{
		"0000",
		"5db57a86b859d1c286b5f1f585048bf8f6b5e626573a8dc728ed5080f6f43e2c",
	}
	chunks := [][]byte{
		{0x01, 0x01},
		{0x02, 0x02},
		{0x03, 0x03},
		{0x00, 0x00},
	}

	r, err := TreeFromChunks(chunks)
	if err != nil {
		t.Errorf("Failed to construct tree: %v\n", err)
	}

	p, err := r.Prove(6)
	if err != nil {
		t.Errorf("Failed to generate proof: %v\n", err)
	}

	if len(p) != len(expectedProofHex) {
		t.Errorf("Proof has invalid length. Expected %d, got %d\n", len(expectedProofHex), len(p))
	}
	for i, n := range p {
		e, err := hex.DecodeString(expectedProofHex[i])
		if err != nil {
			t.Errorf("Failed to decode hex string: %v\n", err)
		}
		if !bytes.Equal(e, n) {
			t.Errorf("Invalid proof item. Expected %s, got %s\n", expectedProofHex[i], hex.EncodeToString(n))
		}
	}
}
