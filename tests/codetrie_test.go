package tests

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/rand"
	"testing"
	"time"

	ssz "github.com/ferranbt/fastssz"
	"github.com/ferranbt/fastssz/spectests"
	"github.com/minio/sha256-simd"
)

func TestVerifyMetadataProof(t *testing.T) {
	testCases := []struct {
		root  string
		proof []string
		leaf  string
		index int
		valid bool
	}{
		{
			root: "2a23ef2b7a7221eaac2ffb3842a506a981c009ca6c2fcbf20adbc595e56f1a93",
			proof: []string{
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
			},
			leaf:  "0100000000000000000000000000000000000000000000000000000000000000",
			index: 4,
			valid: true,
		},
		{
			root: "2a23ef2b7a7221eaac2ffb3842a506a981c009ca6c2fcbf20adbc595e56f1a93",
			proof: []string{
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
			},
			leaf:  "0100000000000000000000000000000000000000000000000000000000000000",
			index: 6,
			valid: false,
		},
		{
			root: "2a23ef2b7a7221eaac2ffb3842a506a981c009ca6c2fcbf20adbc595e56f1a93",
			proof: []string{
				"0100000000000000000000000000000000000000000000000000000000000000",
				"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
			},
			leaf:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			index: 5,
			valid: true,
		},
	}

	for _, c := range testCases {
		// Decode values from string to []byte
		root, err := hex.DecodeString(c.root)
		if err != nil {
			t.Errorf("Failed to decode root: %s\n", c.root)
		}
		hashes := make([][]byte, len(c.proof))
		for i, p := range c.proof {
			b, err := hex.DecodeString(p)
			if err != nil {
				t.Errorf("Failed to decode proof element: %s\n", p)
			}
			hashes[i] = b
		}
		leaf, err := hex.DecodeString(c.leaf)
		if err != nil {
			t.Errorf("Failed to decode leaf: %s\n", c.leaf)
		}

		// Verify proof
		proof := &ssz.Proof{Hashes: hashes, Leaf: leaf, Index: c.index}
		ok, err := ssz.VerifyProof(root, proof)
		if err != nil {
			t.Errorf("Failed to verify proof: %v\n", err)
		}
		if ok != c.valid {
			t.Errorf("Incorrect proof verification: expected %v, got %v\n", c.valid, ok)
		}
	}
}

func TestVerifyCodeTrieProof(t *testing.T) {
	testCases := []struct {
		root  string
		proof []string
		leaf  string
		index int
		valid bool
	}{
		{
			root: "f1824b0084956084591ff4c91c11bcc94a40be82da280e5171932b967dd146e9",
			proof: []string{
				"35210d64853aee79d03f30cf0f29c1398706cbbcacaf05ab9524f00070aec91e",
				"f38a181470ef1eee90a29f0af0a9dba6b7e5d48af3c93c29b4f91fa11b777582",
			},
			leaf:  "0100000000000000000000000000000000000000000000000000000000000000",
			index: 7,
			valid: true,
		},
		{
			root: "f1824b0084956084591ff4c91c11bcc94a40be82da280e5171932b967dd146e9",
			proof: []string{
				"0000000000000000000000000000000000000000000000000000000000000000",
				"0000000000000000000000000000000000000000000000000000000000000000",
				"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
				"0100000000000000000000000000000000000000000000000000000000000000",
				"f38a181470ef1eee90a29f0af0a9dba6b7e5d48af3c93c29b4f91fa11b777582",
			},
			leaf:  "6001000000000000000000000000000000000000000000000000000000000000",
			index: 49,
			valid: true,
		},
	}

	for _, c := range testCases {
		// Decode values from string to []byte
		root, err := hex.DecodeString(c.root)
		if err != nil {
			t.Errorf("Failed to decode root: %s\n", c.root)
		}
		hashes := make([][]byte, len(c.proof))
		for i, p := range c.proof {
			b, err := hex.DecodeString(p)
			if err != nil {
				t.Errorf("Failed to decode proof element: %s\n", p)
			}
			hashes[i] = b
		}
		leaf, err := hex.DecodeString(c.leaf)
		if err != nil {
			t.Errorf("Failed to decode leaf: %s\n", c.leaf)
		}

		// Verify proof
		proof := &ssz.Proof{Hashes: hashes, Leaf: leaf, Index: c.index}
		ok, err := ssz.VerifyProof(root, proof)
		if err != nil {
			t.Errorf("Failed to verify proof: %v\n", err)
		}
		if ok != c.valid {
			t.Errorf("Incorrect proof verification: expected %v, got %v\n", c.valid, ok)
		}
	}
}

func TestVerifyCodeTrieMultiProof2(t *testing.T) {
	// https://etherscan.io/tx/0x138a5f8ba7950521d9dec66ee760b101e0c875039e695c9fcfb34f5ef02a881b
	// 0x02f873011a8405f5e10085037fcc60e182520894f7eaaf75cb6ec4d0e2b53964ce6733f54f7d3ffc880b6139a7cbd2000080c080a095a7a3cbb7383fc3e7d217054f861b890a935adc1adf4f05e3a2f23688cf2416a00875cdc45f4395257e44d709d04990349b105c22c11034a60d7af749ffea2765
	// https://etherscan.io/tx/0xfb0ee9de8941c8ad50e6a3d2999cd6ef7a541ec9cb1ba5711b76fcfd1662dfa9
	// 0xf8708305dc6885029332e35883019a2894500b0107e172e420561565c8177c28ac0f62017f8810ffb80e6cc327008025a0e9c0b380c68f040ae7affefd11979f5ed18ae82c00e46aa3238857c372a358eca06b26e179dd2f7a7f1601755249f4cff56690c4033553658f0d73e26c36fe7815
	// https://etherscan.io/tx/0x45e7ee9ba1a1d0145de29a764a33bb7fc5620486b686d68ec8cb3182d137bc90
	// 0xf86c0785028fa6ae0082520894098d880c4753d0332ca737aa592332ed2522cd22880d2f09f6558750008026a0963e58027576b3a8930d7d9b4a49253b6e1a2060e259b2102e34a451d375ce87a063f802538d3efed17962c96fcea431388483bbe3860ea9bb3ef01d4781450fbf
	// https://etherscan.io/tx/0x9d48b4a021898a605b7ae49bf93ad88fa6bd7050e9448f12dde064c10f22fe9c
	// 0x02f87601836384348477359400850517683ba883019a28943678fce4028b6745eb04fa010d9c8e4b36d6288c872b0f1366ad800080c080a0b6b7aba1954160d081b2c8612e039518b9c46cd7df838b405a03f927ad196158a071d2fb6813e5b5184def6bd90fb5f29e0c52671dea433a7decb289560a58416e

	raw := []string{"0x02f873011a8405f5e10085037fcc60e182520894f7eaaf75cb6ec4d0e2b53964ce6733f54f7d3ffc880b6139a7cbd2000080c080a095a7a3cbb7383fc3e7d217054f861b890a935adc1adf4f05e3a2f23688cf2416a00875cdc45f4395257e44d709d04990349b105c22c11034a60d7af749ffea2765", "0xf8708305dc6885029332e35883019a2894500b0107e172e420561565c8177c28ac0f62017f8810ffb80e6cc327008025a0e9c0b380c68f040ae7affefd11979f5ed18ae82c00e46aa3238857c372a358eca06b26e179dd2f7a7f1601755249f4cff56690c4033553658f0d73e26c36fe7815", "0xf86c0785028fa6ae0082520894098d880c4753d0332ca737aa592332ed2522cd22880d2f09f6558750008026a0963e58027576b3a8930d7d9b4a49253b6e1a2060e259b2102e34a451d375ce87a063f802538d3efed17962c96fcea431388483bbe3860ea9bb3ef01d4781450fbf", "0x02f87601836384348477359400850517683ba883019a28943678fce4028b6745eb04fa010d9c8e4b36d6288c872b0f1366ad800080c080a0b6b7aba1954160d081b2c8612e039518b9c46cd7df838b405a03f927ad196158a071d2fb6813e5b5184def6bd90fb5f29e0c52671dea433a7decb289560a58416e"}

	byteTxs := make([][]byte, len(raw))
	for i := range byteTxs {
		byteTxs[i], _ = hex.DecodeString(raw[i][2:])
	}

	bellatrixPayloadTxs := spectests.ExecutionPayloadTransactions{Transactions: byteTxs}

	rootNode, err := bellatrixPayloadTxs.GetTree()
	if err != nil {
		t.Errorf("Failed to construct tree for transactions: %v\n", err)
	}

	rootNode.Hash()

	// using gen index formula: 2 ** 21
	baseGeneralizedIndex := int(math.Pow(float64(2), float64(21)))
	generalizedIndexes := make([]int, 2)
	// prove inclusion of some transactions
	generalizedIndexes[0] = baseGeneralizedIndex
	generalizedIndexes[1] = baseGeneralizedIndex + 3

	multiProof, err := rootNode.ProveMulti(generalizedIndexes)
	if err != nil {
		t.Errorf("Failed to generate multiproof: %v\n", err)
	}
	if ok, err := ssz.VerifyMultiproof(rootNode.Hash(), multiProof.Hashes, multiProof.Leaves, multiProof.Indices); !ok || err != nil {
		t.Errorf("NOT OK while verifying multiproof: %v\n", err)
	} else {
		t.Logf("OK while verifying multiproof\n")
	}
}

func TestVerifyCodeTrieMultiProof(t *testing.T) {
	testCases := []struct {
		root    string
		proof   []string
		leaves  []string
		indices []int
		valid   bool
	}{
		{
			root: "f1824b0084956084591ff4c91c11bcc94a40be82da280e5171932b967dd146e9",
			proof: []string{
				"0000000000000000000000000000000000000000000000000000000000000000",
				"0000000000000000000000000000000000000000000000000000000000000000",
				"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
				"0000000000000000000000000000000000000000000000000000000000000000",
				"0100000000000000000000000000000000000000000000000000000000000000",
				"f58f76419d9235451a8290a88ba380d852350a1843f8f26b8257a421633042b4",
			},
			leaves: []string{
				"0200000000000000000000000000000000000000000000000000000000000000",
				"6001000000000000000000000000000000000000000000000000000000000000",
			},
			indices: []int{10, 49},
			valid:   true,
		},
	}

	for _, c := range testCases {
		// Decode values from string to []byte
		root, err := hex.DecodeString(c.root)
		if err != nil {
			t.Errorf("Failed to decode root: %s\n", c.root)
		}
		proof := make([][]byte, len(c.proof))
		for i, p := range c.proof {
			b, err := hex.DecodeString(p)
			if err != nil {
				t.Errorf("Failed to decode proof element: %s\n", p)
			}
			proof[i] = b
		}
		leaves := make([][]byte, len(c.leaves))
		for i, l := range c.leaves {
			b, err := hex.DecodeString(l)
			if err != nil {
				t.Errorf("Failed to decode leaf: %s\n", l)
			}
			leaves[i] = b
		}

		// Verify proof
		ok, err := ssz.VerifyMultiproof(root, proof, leaves, c.indices)
		if err != nil {
			t.Errorf("Failed to verify proof: %v\n", err)
		}
		if ok != c.valid {
			t.Errorf("Incorrect proof verification: expected %v, got %v\n", c.valid, ok)
		}
	}
}

func TestMetadataTree(t *testing.T) {
	code := []byte{0x60, 0x01}
	codeHash := sha256.Sum256(code)

	codePadded := make([]byte, 32)
	copy(codePadded[:2], code[:])

	md := &Metadata{Version: 1, CodeLength: uint16(len(code)), CodeHash: codeHash[:]}
	mdRoot, err := md.HashTreeRoot()
	if err != nil {
		t.Errorf("failed to hash metadata tree root: %v\n", err)
	}

	mdTree, err := md.GetTree()
	if err != nil {
		t.Errorf("Failed to construct tree for metadata: %v\n", err)
	}

	r := mdTree.Hash()
	if !bytes.Equal(r, mdRoot[:]) {
		t.Errorf("Computed incorrect root. Expected %s, got %s\n", hex.EncodeToString(mdRoot[:]), hex.EncodeToString(r))
	}
}

func TestChunkTree(t *testing.T) {
	code := []byte{0x60, 0x01}
	codePadded := make([]byte, 32)
	copy(codePadded[:2], code[:])
	chunk := &Chunk{FIO: 0, Code: codePadded[:]}
	chunkRoot, err := chunk.HashTreeRoot()
	if err != nil {
		t.Errorf("Failed to hash chunk to root: %v\n", err)
	}

	tree, err := chunk.GetTree()
	if err != nil {
		t.Errorf("Failed to construct tree for chunk: %v\n", err)
	}

	r := tree.Hash()
	if !bytes.Equal(r, chunkRoot[:]) {
		t.Errorf("Computed incorrect root. Expected %s, got %s\n", hex.EncodeToString(chunkRoot[:]), hex.EncodeToString(r))
	}
}

func TestSmallCodeTrieTree(t *testing.T) {
	code := []byte{0x60, 0x01}
	codeHash := sha256.Sum256(code)

	codePadded := make([]byte, 32)
	copy(codePadded[:2], code[:])

	md := &Metadata{Version: 1, CodeLength: uint16(len(code)), CodeHash: codeHash[:]}
	chunks := []*Chunk{
		{FIO: 0, Code: codePadded[:]},
	}
	codeTrie := &CodeTrieSmall{Metadata: md, Chunks: chunks}
	codeRoot, err := codeTrie.HashTreeRoot()
	if err != nil {
		t.Errorf("failed to hash tree root: %v\n", err)
	}

	tree, err := codeTrie.GetTree()
	if err != nil {
		t.Errorf("Failed to construct tree for codeTrie: %v\n", err)
	}

	r := tree.Hash()
	if !bytes.Equal(r, codeRoot[:]) {
		t.Errorf("Computed incorrect root. Expected %s, got %s\n", hex.EncodeToString(codeRoot[:]), hex.EncodeToString(r))
	}
}

func TestProveSmallCodeTrie(t *testing.T) {
	expectedProofHex := []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0000000000000000000000000000000000000000000000000000000000000000",
		"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
		"0100000000000000000000000000000000000000000000000000000000000000",
		"f38a181470ef1eee90a29f0af0a9dba6b7e5d48af3c93c29b4f91fa11b777582",
	}
	expectedProof, err := parseStringSlice(expectedProofHex)
	if err != nil {
		t.Errorf("Failed to decode expected proof: %v\n", err)
	}

	code := []byte{0x60, 0x01}
	codeHash := sha256.Sum256(code)

	codePadded := make([]byte, 32)
	copy(codePadded[:2], code[:])

	md := &Metadata{Version: 1, CodeLength: uint16(len(code)), CodeHash: codeHash[:]}
	chunks := []*Chunk{
		{FIO: 0, Code: codePadded[:]},
	}
	codeTrie := &CodeTrieSmall{Metadata: md, Chunks: chunks}

	tree, err := codeTrie.GetTree()
	if err != nil {
		t.Errorf("Failed to construct tree for codeTrie: %v\n", err)
	}

	proof, err := tree.Prove(49)
	if err != nil {
		t.Errorf("Failed to generate proof for codeTrie: %v\n", err)
	}

	if proof.Index != 49 {
		t.Errorf("Proof has invalid index\n")
	}
	if !bytes.Equal(proof.Leaf, codePadded) {
		t.Errorf("Proof has invalid leaf\n")
	}
	if len(proof.Hashes) != len(expectedProof) {
		t.Errorf("Generated proof has invalid length\n")
	}

	for i, p := range proof.Hashes {
		if !bytes.Equal(p, expectedProof[i]) {
			t.Errorf("Proof element mismatch. Expected %s, got %s\n", hex.EncodeToString(expectedProof[i]), hex.EncodeToString(p))
		}
	}

	root := tree.Hash()
	ok, err := ssz.VerifyProof(root, proof)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Errorf("Could not verify generated proof")
	}
}

func TestProveMultiSmallCodeTrie(t *testing.T) {
	expectedProofHex := []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0000000000000000000000000000000000000000000000000000000000000000",
		"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0100000000000000000000000000000000000000000000000000000000000000",
		"f58f76419d9235451a8290a88ba380d852350a1843f8f26b8257a421633042b4",
	}
	expectedCProofHex := []string{
		"",
		"",
		"",
		"",
		"0100000000000000000000000000000000000000000000000000000000000000",
		"f58f76419d9235451a8290a88ba380d852350a1843f8f26b8257a421633042b4",
	}
	expectedProof, err := parseStringSlice(expectedProofHex)
	if err != nil {
		t.Errorf("Failed to decode expected proof: %v\n", err)
	}
	expectedCProof, err := parseStringSlice(expectedCProofHex)
	if err != nil {
		t.Errorf("Failed to decode expected compressed proof: %v\n", err)
	}

	code := []byte{0x60, 0x01}
	codeHash := sha256.Sum256(code)

	codePadded := make([]byte, 32)
	copy(codePadded[:2], code[:])

	md := &Metadata{Version: 1, CodeLength: uint16(len(code)), CodeHash: codeHash[:]}
	chunks := []*Chunk{
		{FIO: 0, Code: codePadded[:]},
	}
	codeTrie := &CodeTrieSmall{Metadata: md, Chunks: chunks}

	tree, err := codeTrie.GetTree()
	if err != nil {
		t.Errorf("Failed to construct tree for codeTrie: %v\n", err)
	}

	proof, err := tree.ProveMulti([]int{10, 49})
	if err != nil {
		t.Errorf("Failed to generate proof for codeTrie: %v\n", err)
	}

	if len(proof.Hashes) != len(expectedProof) {
		t.Errorf("Generated proof has invalid length\n")
	}

	for i, p := range proof.Hashes {
		if !bytes.Equal(p, expectedProof[i]) {
			t.Errorf("Proof element mismatch. Expected %s, got %s\n", hex.EncodeToString(expectedProof[i]), hex.EncodeToString(p))
		}
	}

	cproof := proof.Compress()
	if len(cproof.Hashes) != len(expectedCProof) {
		t.Errorf("Generated compressed proof has invalid length\n")
	}

	for i, p := range cproof.Hashes {
		e := expectedCProof[i]
		if (p == nil && e != nil) || (p != nil && e == nil) {
			t.Errorf("Proof element at pos %d was unexpectedly empty\n", i)
		}
		if !bytes.Equal(p, e) {
			t.Errorf("Proof element mismatch. Expected %s, got %s\n", hex.EncodeToString(e), hex.EncodeToString(p))
		}
	}

	// Test uncompression
	uncompressed := cproof.Decompress()
	if len(uncompressed.Hashes) != len(expectedProof) {
		t.Errorf("Uncompressed proof has invalid length. Expected %d, got %d\n", len(expectedProof), len(uncompressed.Hashes))
	}

	for i, p := range uncompressed.Hashes {
		e := expectedProof[i]
		if !bytes.Equal(p, e) {
			t.Errorf("Uncompressed proof element mismatch. Expected %s, got %s\n", hex.EncodeToString(e), hex.EncodeToString(p))
		}
	}
}

func BenchmarkHashTreeRootVsNode(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	codeSize := 24 * 1024
	code := make([]byte, codeSize) // 24Kb
	rand.Read(code)
	codeHash := sha256.Sum256(code)

	md := &Metadata{Version: 1, CodeLength: uint16(codeSize), CodeHash: codeHash[:]}
	chunks := make([]*Chunk, codeSize/32)
	for i := 0; i < len(chunks); i++ {
		chunks[i] = &Chunk{FIO: uint8(i % 256), Code: code[i*32 : (i+1)*32]}
	}

	codeTrie := &CodeTrieBig{Metadata: md, Chunks: chunks}

	b.Run("HashTreeRoot", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codeTrie.HashTreeRoot()
		}
	})
	b.Run("NodeHash", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tree, err := codeTrie.GetTree()
			if err != nil {
				b.Errorf("Failed to construct tree for codeTrie: %v\n", err)
			}

			tree.Hash()
		}
	})
}

func parseStringSlice(slice []string) ([][]byte, error) {
	res := make([][]byte, len(slice))
	for i, s := range slice {
		if len(s) == 0 {
			res[i] = nil
			continue
		}

		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}
		res[i] = b
	}
	return res, nil
}
