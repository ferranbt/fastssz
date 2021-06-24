package main

import (
	"strings"
	"testing"
)

func normalizeGenerated(s string) string {
	lines := strings.Split(s, "\n")
	result := make([]string, 0)
	for _, l := range lines {
		result = append(result, strings.TrimSpace(l))
	}
	return strings.Join(result, "\n")
}

var expectedMisalignedInner = `if len(i) != 48 {
	err = ssz.ErrBytesLength
	return
}
padded := make([]byte, 64)
copy(padded[0:48], i[0:48])
hh.Append(padded)`

var expectedAlignedInner = `if len(i) != 32 {
	err = ssz.ErrBytesLength
	return
}
hh.Append(i)`

func TestHashRootInnerMisaligned(t *testing.T) {
	e := TypeBytes
	v := &Value{
		e: &Value{
			c: false,
			s: 48,
		},
	}
	actual := v.hashRootsInner(e)
	if normalizeGenerated(actual) != normalizeGenerated(expectedMisalignedInner) {
		t.Errorf("want=%s, got=%s", expectedMisalignedInner, actual)
	}
}

func TestHashRootAligned(t *testing.T) {
	e := TypeBytes
	v := &Value{
		e: &Value{
			c: false,
			s: 32,
		},
	}
	actual := v.hashRootsInner(e)
	if normalizeGenerated(actual) != normalizeGenerated(expectedAlignedInner) {
		t.Errorf("want=%s, got=%s", expectedAlignedInner, actual)
	}
}

func TestNextChunkAlignemnt(t *testing.T) {
	cases := []struct{
		size int
		aligned int
	}{
		{
			size: 48,
			aligned: 64,
		},
		{
			size: 1,
			aligned: 32,
		},
		{
			size: 32,
			aligned: 32,
		},
		{
			size: 64,
			aligned: 64,
		},
	}
	for _, c := range cases {
		if c.aligned != nextChunkAlignment(c.size) {
			t.Errorf("Expected nextChunkAlignment(%d) == %d, got %d", c.size, c.size, nextChunkAlignment(c.size))
		}
	}
}