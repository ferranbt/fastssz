package ssz

import (
	"encoding/binary"
	"fmt"
)

// ---- Unmarshal functions ----

// UnmarshallUint64 unmarshals a little endian uint64 from the src input
func UnmarshallUint64(src []byte) uint64 {
	return binary.LittleEndian.Uint64(src)
}

// UnmarshallUint32 unmarshals a little endian uint32 from the src input
func UnmarshallUint32(src []byte) uint32 {
	return binary.LittleEndian.Uint32(src[:4])
}

// UnmarshallUint16 unmarshals a little endian uint16 from the src input
func UnmarshallUint16(src []byte) uint16 {
	return binary.LittleEndian.Uint16(src[:2])
}

// UnmarshallUint8 unmarshals a little endian uint8 from the src input
func UnmarshallUint8(src []byte) uint8 {
	return uint8(src[0])
}

// UnmarshalBool unmarshals a boolean from the src input
func UnmarshalBool(src []byte) bool {
	if src[0] == 1 {
		return true
	}
	return false
}

// ---- Marshal functions ----

// MarshalFixedBytes marshals buf of fixed size to dst
func MarshalFixedBytes(dst []byte, buf []byte, size int) ([]byte, error) {
	if len(buf) == 0 {
		buf = make([]byte, size)
	}
	if len(buf) != size {
		return nil, fmt.Errorf("expected size %d but found %d", len(buf), size)
	}
	dst = append(dst, buf...)
	return dst, nil
}

// MarshalUint64 marshals a little endian uint64 to dst
func MarshalUint64(dst []byte, i uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	dst = append(dst, buf...)
	return dst
}

// MarshalUint32 marshals a little endian uint32 to dst
func MarshalUint32(dst []byte, i uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, i)
	dst = append(dst, buf...)
	return dst
}

// MarshalUint16 marshals a little endian uint16 to dst
func MarshalUint16(dst []byte, i uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, i)
	dst = append(dst, buf...)
	return dst
}

// MarshalUint8 marshals a little endian uint8 to dst
func MarshalUint8(dst []byte, i uint8) []byte {
	dst = append(dst, byte(i))
	return dst
}

// MarshalBool marshals a boolean to dst
func MarshalBool(dst []byte, b bool) []byte {
	if b {
		dst = append(dst, 1)
	} else {
		dst = append(dst, 0)
	}
	return dst
}

// ---- offset functions ----

// WriteOffset writes an offset to dst
func WriteOffset(dst []byte, i int) []byte {
	return MarshalUint32(dst, uint32(i))
}

// ReadOffset reads an offset from buf
func ReadOffset(buf []byte) uint64 {
	return uint64(binary.LittleEndian.Uint32(buf))
}

func safeReadOffset(buf []byte) (uint64, []byte, error) {
	if len(buf) < 4 {
		return 0, nil, fmt.Errorf("")
	}
	offset := ReadOffset(buf)
	return offset, buf[4:], nil
}

// ---- extend functions ----

// ExtendUint64 extends a uint64 buffer to a given size
func ExtendUint64(b []uint64, needLen int) []uint64 {
	b = b[:cap(b)]
	if n := needLen - cap(b); n > 0 {
		b = append(b, make([]uint64, n)...)
	}
	return b[:needLen]
}

// ExtendUint16 extends a uint16 buffer to a given size
func ExtendUint16(b []uint16, needLen int) []uint16 {
	b = b[:cap(b)]
	if n := needLen - cap(b); n > 0 {
		b = append(b, make([]uint16, n)...)
	}
	return b[:needLen]
}

// ---- unmarshal dynami content ----

const bytesPerLengthOffset = 4

// DecodeDynamicLength decodes the length from the dynamic input
func DecodeDynamicLength(buf []byte, maxSize int) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	if len(buf) < 4 {
		return 0, fmt.Errorf("not enough data")
	}
	offset := binary.LittleEndian.Uint32(buf[:4])
	length, ok := DivideInt(int(offset), bytesPerLengthOffset)
	if !ok {
		return 0, fmt.Errorf("bad")
	}
	if length > maxSize {
		return 0, fmt.Errorf("too big for the list")
	}
	return length, nil
}

// UnmarshalDynamic unmarshals the dynamic items from the input
func UnmarshalDynamic(src []byte, length int, f func(indx int, b []byte) error) error {
	var err error
	if length == 0 {
		return nil
	}

	size := uint64(len(src))

	indx := 0
	dst := src

	var offset, endOffset uint64
	offset, dst = ReadOffset(src), dst[4:]

	for {
		if length != 1 {
			endOffset, dst, err = safeReadOffset(dst)
			if err != nil {
				return err
			}
		} else {
			endOffset = uint64(len(src))
		}
		if offset > endOffset {
			return fmt.Errorf("four")
		}
		if endOffset > size {
			return fmt.Errorf("five")
		}

		err := f(indx, src[offset:endOffset])
		if err != nil {
			return err
		}

		indx++

		offset = endOffset
		if length == 1 {
			break
		}
		length--
	}
	return nil
}

// DivideInt divides the int fully
func DivideInt(a, b int) (int, bool) {
	return a / b, a%b == 0
}
