package tests

import (
	"encoding/hex"
	"testing"
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
		proof := make([][]byte, len(c.proof))
		for i, p := range c.proof {
			b, err := hex.DecodeString(p)
			if err != nil {
				t.Errorf("Failed to decode proof element: %s\n", p)
			}
			proof[i] = b
		}
		leaf, err := hex.DecodeString(c.leaf)
		if err != nil {
			t.Errorf("Failed to decode leaf: %s\n", c.leaf)
		}

		// Verify proof
		ok, err := VerifyMetadataProof(root, proof, leaf, c.index)
		if err != nil {
			t.Errorf("Failed to verify proof: %v\n", err)
		}
		if ok != c.valid {
			t.Errorf("Incorrect proof verification: expected %v, got %v\n", c.valid, ok)
		}
	}
}
