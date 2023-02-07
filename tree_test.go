package ssz

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestParseTree(t *testing.T) {
	chunk1, err := hex.DecodeString("9a4aaa9f8c50cdb565a05ed94a0019cbea56349bdb4c5b639a26bcfed855c790")
	require.NoError(t, err)
	chunk2, err := hex.DecodeString("632a7e04caca67eed732cd670409acf2daaf88aed3977689446ba6f7d3e43aa4")
	require.NoError(t, err)
	chunk3, err := hex.DecodeString("6314fea8253a30f23d5af34b0b2e675d4c0d475e4f66e4392c535f2ca5c3ae32")
	require.NoError(t, err)

	chunks := [][]byte{chunk1, chunk2, chunk3}

	nodes := []*Node{}
	for _, chunk := range chunks {
		nodes = append(nodes, LeafFromBytes(chunk[:]))
	}

	r, err := TreeFromNodesWithMixin(nodes, len(nodes), 8)
	require.NoError(t, err, "failed to construct tree")
	require.Equal(t, "850f07566ebef9934782eec2db35b997a44a37aa4eab01a4d25f02e807602136", hex.EncodeToString(r.Hash()))
}

func TestSparseTreeWithLeavesWithOtherNodes(t *testing.T) {
	valueIndex2, err := hex.DecodeString("452a7e04caca67eed732cd670409acf2daaf88aed3977689446ba6f7d3e43aa4")
	require.NoError(t, err)
	valueIndex3, err := hex.DecodeString("ef2a7e04caca67eed732cd670409acf2daaf88aed3977689446ba6f7d3e43aa4")
	require.NoError(t, err)
	valueIndex4, err := hex.DecodeString("842a7e04caca67eed732cd670409acf2daaf88aed3977689446ba6f7d3e43aa4")
	require.NoError(t, err)
	valueIndex5, err := hex.DecodeString("722a7e04caca67eed732cd670409acf2daaf88aed3977689446ba6f7d3e43aa4")
	require.NoError(t, err)
	valueIndex6, err := hex.DecodeString("982a7e04caca67eed732cd670409acf2daaf88aed3977689446ba6f7d3e43aa4")
	require.NoError(t, err)
	valueIndex7, err := hex.DecodeString("632a7e04caca67eed732cd670409acf2daaf88aed3977689446ba6f7d3e43aa4")
	require.NoError(t, err)

	nodes := []*Node{
		{
			left: &Node{
				value: valueIndex2,
			},
			right: &Node{
				value: valueIndex3,
			},
		},
		{
			left: &Node{
				value: valueIndex4,
			},
			right: &Node{
				value: valueIndex5,
			},
		},
		{
			left: &Node{
				value: valueIndex6,
			},
			right: &Node{
				value: valueIndex7,
			},
		},
	}

	limit := 8

	r, err := TreeFromNodesWithMixin(nodes, len(nodes), limit)
	require.NoError(t, err, "failed to construct tree")
	require.Equal(t, "8162efcb0b2e5da308a6a6fb2d4c5c8b65a77a475247d7f751ef998f9a70f294", hex.EncodeToString(r.Hash()))
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

	if p.Index != 6 {
		t.Errorf("Proof has invalid index. Expected %d, got %d\n", 6, p.Index)
	}
	if !bytes.Equal(p.Leaf, chunks[2]) {
		t.Errorf("Proof has invalid leaf. Expected %v, got %v\n", chunks[2], p.Leaf)
	}
	if len(p.Hashes) != len(expectedProofHex) {
		t.Errorf("Proof has invalid length. Expected %d, got %d\n", len(expectedProofHex), len(p.Hashes))
	}

	for i, n := range p.Hashes {
		e, err := hex.DecodeString(expectedProofHex[i])
		if err != nil {
			t.Errorf("Failed to decode hex string: %v\n", err)
		}
		if !bytes.Equal(e, n) {
			t.Errorf("Invalid proof item. Expected %s, got %s\n", expectedProofHex[i], hex.EncodeToString(n))
		}
	}
}

func TestProveMulti(t *testing.T) {
	chunks := [][]byte{
		{0x01, 0x01},
		{0x02, 0x02},
		{0x03, 0x03},
		{0x04, 0x04},
	}

	r, err := TreeFromChunks(chunks)
	if err != nil {
		t.Errorf("Failed to construct tree: %v\n", err)
	}

	p, err := r.ProveMulti([]int{6, 7})
	if err != nil {
		t.Errorf("Failed to generate proof: %v\n", err)
	}

	if len(p.Hashes) != 1 {
		t.Errorf("Incorrect number of hashes in proof. Expected 1, got %d\n", len(p.Hashes))
	}
}

func TestGetRequiredIndices(t *testing.T) {
	indices := []int{10, 48, 49}
	expected := []int{25, 13, 11, 7, 4}
	req := getRequiredIndices(indices)
	if len(expected) != len(req) {
		t.Fatalf("Required indices has wrong length. Expected %d, got %d\n", len(expected), len(req))
	}
	for i, r := range req {
		if r != expected[i] {
			t.Errorf("Invalid required index. Expected %d, got %d\n", expected[i], r)
		}
	}
}
