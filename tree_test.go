package ssz

import (
	"bytes"
	"testing"
)

func TestTreeFromChunks(t *testing.T) {
	chunks := [][]byte{
		{0x01, 0x01},
		{0x02, 0x02},
		{0x03, 0x03},
	}

	r := TreeFromChunks(chunks)
	for i := 4; i < 8; i++ {
		l, err := r.Get(i)
		if err != nil {
			t.Errorf("Failed getting leaf: %v\n", err)
		}
		if i > 6 {
			if l.value != nil {
				t.Errorf("Incorrect leaf at index %d. Leaf should be empty\n", i)
			}
		} else {
			if !bytes.Equal(l.value, chunks[i-4]) {
				t.Errorf("Incorrect leaf at index %d\n", i)
			}
		}
	}
}
