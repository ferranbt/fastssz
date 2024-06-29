package ssz

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
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

func TestProveRepeated(t *testing.T) {
	expectedProofHex := []string{
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"2dba5dbc339e7316aea2683faf839c1b7b1ee2313db792112588118df066aa35",
		"5310a330e8f970388503c73349d80b45cd764db615f1bced2801dcd4524a2ff4",
		"80d1bf4dd6c1f75bba022337a3f0842078f5c2e7f3f59dfd33ccbb8e963367b2",
		"1492e66e89e186840231850712161255d203b5bbf48d21242f0b51519b5eb3d4",
		"03a82289eea21de37e72ad6c07865dcab3f2cd681ad47c1cd0ea30e1751ad996",
		"35603b6278eb5d320c99eeb68354d448493e1ab9857cb0bddb9f7fa72250a3a8",
		"8ff9103704f4e7dfee6106551eb439d3ac6bc5cc4873ced8ec33eaf2d42f4c31",
		"259ca0ef3ecb66bb9f02e2ca9de6c7ff13951ad824ece4c680555cfef4321d17",
		"4f52a2051143520841633a6e53f1ad5948a584dcdbc8ea206d8008d1cfe104a9",
		"7ef919cf6137226a4c132f3bcab47a11aa1dfe78a357c19c0c804508829f2623",
		"cbafa51c68b69bc206500c4733c2cc4cc6b67f712cc5fbad5b2d365998ba37a0",
		"e11746324aa6ce20024a6e4796ae38d2dce7d5e015071a4a2cc96c9b71fafb32",
		"e3b4036e156dd6ccf9e41e36b011fd00f79645e361d02a9484eaba96e3be7179",
		"a3cbeb34d17bf5aa47054abd93e0ea1c992eef8359ad6a0f596ea48e455d540a",
		"d1f9c8fa1339b232013cc9585b380372614a869fd0fd2e3d07bec2f3c96b4c6d",
		"162947673d0323a56dcccedf09d1c45dfe40c1ddcdb14f69880ae36f60ee434f",
		"62c0a299966c9ec0a031d01d8bd5330b191461a5dd13a4ac5dd662097c6fc099",
		"c324ca782716eac179133ee5f4b315c2a9e6e922aead7963a95302412f5e0001",
		"dcb5df19b4aa726ff826a34a97704b11e111c3cc519ebbe9133f189e05766dfa",
	}

	chunks := make([][]byte, 1048576)
	for i := uint32(0); i < 1048576; i++ {
		x := make([]byte, 4)
		binary.LittleEndian.PutUint32(x, i)
		chunks = append(chunks, x)
	}

	r, err := TreeFromChunks(chunks)
	if err != nil {
		t.Errorf("Failed to construct tree: %v\n", err)
		t.Fail()
	}

	// Repeatedly prove the same entry, to ensure that there are no mutations
	// as a result of proving.
	for i := 0; i < 1024; i++ {
		p, err := r.Prove(1048576)
		if err != nil {
			t.Errorf("Failed to generate proof: %v\n", err)
			t.Fail()
		}

		for i, n := range p.Hashes {
			e, err := hex.DecodeString(expectedProofHex[i])
			if err != nil {
				t.Errorf("Failed to decode hex string: %v\n", err)
				t.Fail()
			}
			if !bytes.Equal(e, n) {
				t.Errorf("Invalid proof item. Expected %s, got %s\n", expectedProofHex[i], hex.EncodeToString(n))
				t.Fail()
			}
		}
	}
}

func BenchmarkProve(b *testing.B) {
	chunks := make([][]byte, 1048576)
	for i := uint32(0); i < 1048576; i++ {
		x := make([]byte, 4)
		binary.LittleEndian.PutUint32(x, i)
		chunks = append(chunks, x)
	}

	r, err := TreeFromChunks(chunks)
	if err != nil {
		b.Errorf("Failed to construct tree: %v\n", err)
		b.Fail()
	}

	for i := 0; i < b.N; i++ {
		//nolint
		_, err := r.Prove(rand.Intn(1048575) + 1)
		if err != nil {
			b.Errorf("Failed to generate proof: %v\n", err)
			b.Fail()
		}
	}
}
