package ssz

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"time"
)

// MarshalSSZ marshals an object
func MarshalSSZ(m Marshaler) ([]byte, error) {
	buf := make([]byte, m.SizeSSZ())
	return m.MarshalSSZTo(buf[:0])
}

// Errors

var (
	ErrOffset                = fmt.Errorf("incorrect offset")
	ErrSize                  = fmt.Errorf("incorrect size")
	ErrBytesLength           = fmt.Errorf("bytes array does not have the correct length")
	ErrVectorLength          = fmt.Errorf("vector does not have the correct length")
	ErrListTooBig            = fmt.Errorf("list length is higher than max value")
	ErrEmptyBitlist          = fmt.Errorf("bitlist is empty")
	ErrInvalidVariableOffset = fmt.Errorf("invalid ssz encoding. first variable element offset indexes into fixed value data")
	ErrOffsetNotIncreasing   = fmt.Errorf("offsets are not increasing")
	ErrTailNotEmpty          = fmt.Errorf("buffer was not totally consumed")
)

func ErrBytesLengthFn(name string, found, expected uint64) error {
	return fmt.Errorf("%s (%v): expected %d and %d found", name, ErrBytesLength, expected, found)
}

func ErrVectorLengthFn(name string, found, expected uint64) error {
	return fmt.Errorf("%s (%v): expected %d and %d found", name, ErrBytesLength, expected, found)
}

func ErrListTooBigFn(name string, found, max uint64) error {
	return fmt.Errorf("%s (%v): max expected %d and %d found", name, ErrListTooBig, max, found)
}

// ---- Unmarshal functions ----

func UnmarshalBitList(dst []byte, src []byte, bitLimit uint64) ([]byte, error) {
	if err := ValidateBitlist(src, bitLimit); err != nil {
		return nil, err
	}
	if cap(dst) == 0 {
		dst = make([]byte, 0, len(src))
	}
	dst = append(dst, src...)
	return dst, nil
}

func UnmarshalFixedBytes(buf []byte, src []byte) []byte {
	targetSize := len(buf)
	copy(buf, src[:targetSize])
	return src[targetSize:]
}

func UnmarshalDynamicBytes(src []byte, buf []byte, maxSize ...int) ([]byte, error) {
	if len(maxSize) > 0 && len(buf) > maxSize[0] {
		return nil, ErrBytesLength
	}
	if cap(src) == 0 {
		src = make([]byte, 0, len(buf))
	}
	src = append(src, buf...)
	return src, nil
}

// UnmarshalBytes unmarshals a byte slice from the src input
// If the src is nil, it will create a new byte slice with the content of buf.
func UnmarshalBytes(src []byte, buf []byte, size uint64) ([]byte, []byte) {
	if cap(src) == 0 {
		src = make([]byte, 0, size)
	}
	src = append(src, buf[:size]...)
	return src, buf[size:]
}

type UnmarshallableType interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~bool
}

func UnmarshallValue[T UnmarshallableType](src []byte) (T, []byte) {
	var result any
	tail := src

	switch any(*new(T)).(type) {
	case uint8:
		result = src[0]
		tail = src[1:]
	case uint16:
		result = binary.LittleEndian.Uint16(src[:2])
		tail = src[2:]
	case uint32:
		result = binary.LittleEndian.Uint32(src[:4])
		tail = src[4:]
	case uint64:
		result = binary.LittleEndian.Uint64(src[:8])
		tail = src[8:]
	case bool:
		result = src[0] != 0
		tail = src[1:]
	default:
		panic("unsupported type")
	}

	return result.(T), tail
}

// UnmarshalTime unmarshals a time.Time from the src input
func UnmarshalTime(src []byte) (time.Time, []byte) {
	val, buf := UnmarshallValue[uint64](src)
	return time.Unix(int64(val), 0).UTC(), buf
}

func IsValidBool(src []byte) error {
	if len(src) == 0 {
		// Note, this should not happen since the code generator
		// makes sure that there is at least one byte on src
		return fmt.Errorf("empty slice")
	}

	val := src[0]
	if val != 0 && val != 1 {
		return fmt.Errorf("invalid SSZ boolean byte: 0x%02x", val)
	}

	return nil
}

// ---- Marshal functions ----

type MarshallableType interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~bool
}

func MarshalValue[T MarshallableType](dst []byte, value T) []byte {
	switch any(value).(type) {
	case uint8:
		return append(dst, any(value).(uint8))
	case uint16:
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, any(value).(uint16))
		return append(dst, buf...)
	case uint32:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, any(value).(uint32))
		return append(dst, buf...)
	case uint64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, any(value).(uint64))
		return append(dst, buf...)
	case bool:
		if any(value).(bool) {
			return append(dst, 1)
		} else {
			return append(dst, 0)
		}
	default:
		panic("unsupported type")
	}
}

// MarshalTime marshals a time to dst
func MarshalTime(dst []byte, t time.Time) []byte {
	return MarshalValue[uint64](dst, uint64(t.Unix()))
}

// ---- offset functions ----

// WriteOffset writes an offset to dst
func WriteOffset(dst []byte, i int) []byte {
	return MarshalValue[uint32](dst, uint32(i))
}

// ReadOffset reads an offset from buf
func ReadOffset(buf []byte) (uint64, []byte) {
	offset, buf := UnmarshallValue[uint32](buf)
	return uint64(offset), buf
}

func safeReadOffset(buf []byte) (uint64, []byte, error) {
	if len(buf) < 4 {
		return 0, nil, fmt.Errorf("buffer too short for offset reading")
	}
	offset, buf := ReadOffset(buf)
	return offset, buf, nil
}

// ---- extend functions ----

// Extend extends a slice buffer to a given size
func Extend[T any](b []T, needLen uint64) []T {
	if b == nil {
		b = []T{}
	}
	b = b[:cap(b)]
	if n := needLen - uint64(cap(b)); n > 0 {
		b = append(b, make([]T, n)...)
	}
	return b[:needLen]
}

// ---- unmarshal dynamic content ----

const bytesPerLengthOffset = 4

// ValidateBitlist validates that the bitlist is correct
func ValidateBitlist(buf []byte, bitLimit uint64) error {
	byteLen := len(buf)
	if byteLen == 0 {
		return fmt.Errorf("bitlist empty, it does not have length bit")
	}
	// Maximum possible bytes in a bitlist with provided bitlimit.
	maxBytes := (bitLimit >> 3) + 1
	if byteLen > int(maxBytes) {
		return fmt.Errorf("unexpected number of bytes, got %d but found %d", byteLen, maxBytes)
	}

	// The most significant bit is present in the last byte in the array.
	last := buf[byteLen-1]
	if last == 0 {
		return fmt.Errorf("trailing byte is zero")
	}

	// Determine the position of the most significant bit.
	msb := bits.Len8(last)

	// The absolute position of the most significant bit will be the number of
	// bits in the preceding bytes plus the position of the most significant
	// bit. Subtract this value by 1 to determine the length of the bitlist.
	numOfBits := uint64(8*(byteLen-1) + msb - 1)

	if numOfBits > bitLimit {
		return fmt.Errorf("too many bits")
	}
	return nil
}

// BitlistLen provides the bitlist length of a byte array
func BitlistLen(b []byte) uint64 {
	if len(b) == 0 {
		return 0
	}
	// The most significant bit is present in the last byte in the array.
	last := b[len(b)-1]

	// Determine the position of the most significant bit.
	msb := bits.Len8(last)
	if msb == 0 {
		return 0
	}

	// The absolute position of the most significant bit will be the number of
	// bits in the preceding bytes plus the position of the most significant
	// bit. Subtract this value by 1 to determine the length of the bitlist.
	return uint64(8*(len(b)-1) + msb - 1)
}

// DecodeDynamicLength decodes the length from the dynamic input
func DecodeDynamicLength(buf []byte, maxSize uint64) (uint64, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	if len(buf) < 4 {
		return 0, fmt.Errorf("not enough data")
	}
	offset := binary.LittleEndian.Uint32(buf[:4])
	length, ok := DivideInt(uint64(offset), bytesPerLengthOffset)
	if !ok {
		return 0, fmt.Errorf("incorrect length division")
	}
	if length > maxSize {
		return 0, fmt.Errorf("too big for the list")
	}
	return length, nil
}

// UnmarshalDynamic unmarshals the dynamic items from the input
func UnmarshalDynamic(src []byte, length uint64, f func(indx uint64, b []byte) error) error {
	var err error
	size := uint64(len(src))

	if length == 0 {
		if size != 0 && size != 4 {
			return ErrSize
		}
		return nil
	}

	indx := uint64(0)
	dst := src

	var offset, endOffset uint64
	offset, dst = ReadOffset(src)

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
			return fmt.Errorf("incorrect end of offset: %d %d", offset, endOffset)
		}
		if endOffset > size {
			return fmt.Errorf("incorrect end of offset size: %d %d", endOffset, size)
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

func DivideInt2(a, b, max uint64) (uint64, error) {
	num, ok := DivideInt(a, b)
	if !ok {
		return 0, fmt.Errorf("failed to divide int %d by %d", a, b)
	}
	if num > max {
		return 0, fmt.Errorf("num %d is greater than max %d", num, max)
	}
	return num, nil
}

// DivideInt divides the int fully
func DivideInt(a, b uint64) (uint64, bool) {
	return a / b, a%b == 0
}

type OffsetMarker struct {
	TotalSize  uint64
	FixedSize  uint64
	LastOffset uint64
	HasOffset  bool
}

func NewOffsetMarker(totalSize, fixedSize uint64) *OffsetMarker {
	return &OffsetMarker{
		TotalSize:  totalSize,
		FixedSize:  fixedSize,
		LastOffset: 0,
		HasOffset:  false,
	}
}

func (o *OffsetMarker) ReadOffset(buf []byte) (uint64, []byte, error) {
	offset, buf := ReadOffset(buf)

	if offset > o.TotalSize {
		return 0, nil, ErrOffset
	}
	if !o.HasOffset {
		if offset != o.FixedSize {
			return 0, nil, ErrInvalidVariableOffset
		}
		o.HasOffset = true
	} else {
		if offset < o.LastOffset {
			return 0, nil, ErrOffsetNotIncreasing
		}
	}

	o.LastOffset = offset
	return offset, buf, nil
}
