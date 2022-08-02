package testcases

import (
	"bytes"
	"testing"
	"time"
)

func TestTimeRoot(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		root      []byte
	}{
		{
			name:      "Zero",
			timestamp: 0,
			root:      []byte{0xf5, 0xa5, 0xfd, 0x42, 0xd1, 0x6a, 0x20, 0x30, 0x27, 0x98, 0xef, 0x6e, 0xd3, 0x09, 0x97, 0x9b, 0x43, 0x00, 0x3d, 0x23, 0x20, 0xd9, 0xf0, 0xe8, 0xea, 0x98, 0x31, 0xa9, 0x27, 0x59, 0xfb, 0x4b},
		},
		{
			name:      "Max",
			timestamp: 0x7fffffffffffffff,
			root:      []byte{0xd6, 0xa4, 0x84, 0x7f, 0x18, 0x2d, 0xec, 0x2b, 0xd3, 0x99, 0xd9, 0x7d, 0xb5, 0x96, 0xb4, 0x83, 0x15, 0x7c, 0x5c, 0x62, 0x4c, 0x12, 0x0b, 0x25, 0x0a, 0xde, 0xa9, 0xf5, 0xe9, 0x93, 0x6c, 0x4f},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timeStruct := TimeType{Timestamp: time.Unix(test.timestamp, 0)}
			timeRoot, err := timeStruct.HashTreeRoot()
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			// Expect the root to match the expectation.
			if !bytes.Equal(test.root, timeRoot[:]) {
				t.Fatalf("root mismatch with time type")
			}

			rawStruct := TimeRawType{Timestamp: uint64(test.timestamp)}
			rawRoot, err := rawStruct.HashTreeRoot()
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			// Expect the root to match the expectation.
			if !bytes.Equal(test.root, rawRoot[:]) {
				t.Fatalf("root mismatch with raw type")
			}
		})
	}
}

func TestTimeEncode(t *testing.T) {
	tests := []struct {
		name       string
		timestamp  int64
		marshalled []byte
	}{
		{
			name:       "Zero",
			timestamp:  0,
			marshalled: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:       "Real",
			timestamp:  0x62e8e1b1,
			marshalled: []byte{0xb1, 0xe1, 0xe8, 0x62, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:       "Max",
			timestamp:  0x7fffffffffffffff,
			marshalled: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timeStruct := TimeType{Timestamp: time.Unix(test.timestamp, 0)}
			timeMarshalled, err := timeStruct.MarshalSSZ()
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			if !bytes.Equal(test.marshalled, timeMarshalled) {
				t.Fatalf("marshal mismatch with time type (%v != %v)", test.marshalled, timeMarshalled)
			}

			rawStruct := TimeRawType{Timestamp: uint64(test.timestamp)}
			rawMarshalled, err := rawStruct.MarshalSSZ()
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			if !bytes.Equal(test.marshalled, rawMarshalled) {
				t.Fatalf("marshal mismatch with raw type")
			}
		})
	}
}

func TestTimeDecode(t *testing.T) {
	tests := []struct {
		name       string
		marshalled []byte
		timeStruct *TimeType
		rawStruct  *TimeRawType
	}{
		{
			name:       "Zero",
			marshalled: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			timeStruct: &TimeType{
				Timestamp: time.Unix(0, 0),
			},
			rawStruct: &TimeRawType{},
		},
		{
			name:       "Real",
			marshalled: []byte{0xb1, 0xe1, 0xe8, 0x62, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			timeStruct: &TimeType{
				Timestamp: time.Unix(0x62e8e1b1, 0),
			},
			rawStruct: &TimeRawType{
				Timestamp: 0x62e8e1b1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timeStruct := TimeType{}
			if err := timeStruct.UnmarshalSSZ(test.marshalled); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			if !test.timeStruct.Timestamp.Equal(timeStruct.Timestamp) {
				t.Fatalf("unmarshal mismatch with time type (%v != %v)", test.timeStruct.Timestamp, timeStruct.Timestamp)
			}

			rawStruct := TimeRawType{}
			if err := rawStruct.UnmarshalSSZ(test.marshalled); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			if test.rawStruct.Timestamp != rawStruct.Timestamp {
				t.Fatalf("unmarshal mismatch with time type")
			}
		})
	}
}
