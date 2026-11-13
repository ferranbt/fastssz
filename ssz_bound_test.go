package ssz

import (
	"encoding/binary"
	"runtime"
	"testing"
)

// DecodeDynamicLength derives the element count from the first 4-byte offset and
// checks it only against the type's ssz-max, never against len(buf). A valid
// first offset always points just past the offset table, so offset <= len(buf);
// without that check a 4-byte input can claim (and Extend can allocate) the full
// ssz-max element count. This mirrors the bound the fixed-size path already has.
func TestUnmarshalDynamicSliceRejectsOffsetBeyondBuf(t *testing.T) {
	const maxElements = 1 << 20 // e.g. ExecutionPayload.Transactions ssz-max

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(4*maxElements)) // first offset = 4*maxElements

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	var slice [][]byte
	err := UnmarshalDynamicSliceWithCallback(&slice, buf, maxElements,
		func(uint64, []byte) error { return nil })

	runtime.ReadMemStats(&after)
	alloc := int64(after.TotalAlloc - before.TotalAlloc)
	t.Logf("err=%v alloc=%d len(slice)=%d", err, alloc, len(slice))

	if err == nil {
		t.Fatal("expected an error for a first offset beyond the buffer, got nil")
	}
	if alloc > 1*1024*1024 {
		t.Fatalf("allocated %d bytes from a 4-byte input before validating the offset (DoS)", alloc)
	}
}

// DecodeDynamicLength itself must reject an offset past the end of the buffer.
func TestDecodeDynamicLengthRejectsOffsetBeyondBuf(t *testing.T) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, 4_000_000) // offset far past len(buf)=4
	if _, err := DecodeDynamicLength(buf, 1<<20); err == nil {
		t.Fatal("expected an error for an offset beyond the buffer, got nil")
	}
}

// A well-formed empty dynamic list still decodes.
func TestDecodeDynamicLengthEmpty(t *testing.T) {
	n, err := DecodeDynamicLength(nil, 1<<20)
	if err != nil || n != 0 {
		t.Fatalf("empty input: got n=%d err=%v, want 0/nil", n, err)
	}
}

// A well-formed single-element offset table still decodes to the right count.
func TestDecodeDynamicLengthValid(t *testing.T) {
	// One element: offset table is a single 4-byte offset = 4, then element data.
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[:4], 4)
	n, err := DecodeDynamicLength(buf, 1<<20)
	if err != nil || n != 1 {
		t.Fatalf("valid single element: got n=%d err=%v, want 1/nil", n, err)
	}
}
