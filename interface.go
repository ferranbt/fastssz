package ssz

// Marshaler is the interface implemented by types that can marshal themselves into valid SZZ.
type Marshaler interface {
	MarshalSSZTo(dst []byte) ([]byte, error)
	MarshalSSZ() ([]byte, error)
	SizeSSZ() int
}

// Unmarshaler is the interface implemented by types that can unmarshal a SSZ description of themselves
type Unmarshaler interface {
	UnmarshalSSZ(buf []byte) error
	UnmarshalSSZTail(buf []byte) ([]byte, error)
}

type HashRoot interface {
	GetTree() (*Node, error)
	HashTreeRoot() ([32]byte, error)
	HashTreeRootWith(hh HashWalker) error
}

type HashRootProof interface {
	HashTreeRootWith(hh HashWalker) error
}

type HashWalker interface {
	// Intended for testing purposes to know the latest hash generated during merkleize
	Hash() []byte
	AppendUint8(i uint8)
	AppendUint16(i uint16)
	AppendUint32(i uint32)
	AppendUint64(i uint64)
	AppendBytes32(b []byte)
	PutUint64Array(b []uint64, maxCapacity ...uint64)
	PutUint64(i uint64)
	PutUint32(i uint32)
	PutUint16(i uint16)
	PutUint8(i uint8)
	FillUpTo32()
	Append(i []byte)
	PutBitlist(bb []byte, maxSize uint64)
	PutBool(b bool)
	PutBytes(b []byte)
	Index() int
	Merkleize(indx int)
	MerkleizeWithMixin(indx int, num, limit uint64)
}

type PtrConstraint[T any] interface {
	*T
	Unmarshaler
}

func UnmarshalFieldTail[T any, PT PtrConstraint[T]](field *PT, buf []byte) ([]byte, error) {
	if *field == nil {
		*field = PT(new(T))
	}
	return (*field).UnmarshalSSZTail(buf)
}

func UnmarshalField[T any, PT PtrConstraint[T]](field *PT, buf []byte) error {
	if *field == nil {
		*field = PT(new(T))
	}
	return (*field).UnmarshalSSZ(buf)
}

// UnmarshalSliceWithIndexCallback handles slices with index-aware unmarshal logic
func UnmarshalSliceWithIndexCallback[T any](
	slice *[]T,
	buf []byte,
	itemSize int,
	maxItems int,
	unmarshalCallback func(int, []byte) error,
) error {
	num, err := DivideInt2(len(buf), itemSize, maxItems)
	if err != nil {
		return err
	}

	*slice = make([]T, num)
	for ii := 0; ii < num; ii++ {
		start := ii * itemSize
		end := (ii + 1) * itemSize
		if err := unmarshalCallback(ii, buf[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalDynamicSliceWithCallback handles dynamic slices with custom unmarshal logic
func UnmarshalDynamicSliceWithCallback[T any](
	slice *[]T,
	buf []byte,
	maxElements int,
	unmarshalCallback func(int, []byte) error,
) error {
	num, err := DecodeDynamicLength(buf, maxElements)
	if err != nil {
		return err
	}

	*slice = make([]T, num)
	return UnmarshalDynamic(buf, num, unmarshalCallback)
}

func UnmarshalSliceSSZ[T any, PT PtrConstraint[T]](
	slice *[]PT,
	buf []byte,
	itemSize int,
	maxItems int,
) error {
	return UnmarshalSliceWithIndexCallback(slice, buf, itemSize, maxItems,
		func(ii int, itemBuf []byte) error {
			return UnmarshalField[T, PT](&(*slice)[ii], itemBuf)
		})
}

// UnmarshalDynamicSliceSSZ handles dynamic slices of SSZ types
func UnmarshalDynamicSliceSSZ[T any, PT PtrConstraint[T]](
	slice *[]PT,
	buf []byte,
	maxElements int,
) error {
	return UnmarshalDynamicSliceWithCallback(slice, buf, maxElements,
		func(indx int, itemBuf []byte) error {
			return UnmarshalField[T, PT](&(*slice)[indx], itemBuf)
		})
}

func UnmarshalSSZ(v Unmarshaler, buf []byte) error {
	tail, err := v.UnmarshalSSZTail(buf)
	if err != nil {
		return err
	}
	if len(tail) != 0 {
		return ErrTailNotEmpty
	}
	return err
}
