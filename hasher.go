package ssz

import (
	"fmt"
	"hash"
	"math/bits"
	"sync"

	"encoding/binary"

	"github.com/minio/sha256-simd"
)

var (
	// ErrIncorrectByteSize means that the byte size is incorrect
	ErrIncorrectByteSize = fmt.Errorf("incorrect byte size")

	// ErrIncorrectListSize means that the size of the list is incorrect
	ErrIncorrectListSize = fmt.Errorf("incorrect list size")
)

var zeroHashes [65][32]byte
var zeroHashLevels map[string]int
var trueBytes, falseBytes []byte

func init() {
	falseBytes = make([]byte, 32)
	trueBytes = make([]byte, 32)
	trueBytes[0] = 1
	zeroHashLevels = make(map[string]int)
	zeroHashLevels[string(falseBytes)] = 0

	tmp := [64]byte{}
	for i := 0; i < 64; i++ {
		copy(tmp[:32], zeroHashes[i][:])
		copy(tmp[32:], zeroHashes[i][:])
		zeroHashes[i+1] = sha256.Sum256(tmp[:])
		zeroHashLevels[string(zeroHashes[i+1][:])] = i + 1
	}
}

// HashWithDefaultHasher hashes a HashRoot object with a Hasher from
// the default HasherPool
func HashWithDefaultHasher(v HashRoot) ([32]byte, error) {
	hh := DefaultHasherPool.Get()
	if err := v.HashTreeRootWith(hh); err != nil {
		DefaultHasherPool.Put(hh)
		return [32]byte{}, err
	}
	root, err := hh.HashRoot()
	DefaultHasherPool.Put(hh)
	return root, err
}

var zeroBytes = make([]byte, 32)

// DefaultHasherPool is a default hasher pool
var DefaultHasherPool HasherPool

// Hasher is a utility tool to hash SSZ structs
type Hasher struct {
	// buffer array to store hashing values
	buf []byte

	// tmp array used for uint64 and bitlist processing
	tmp []byte

	// tmp array used during the merkleize process
	merkleizeTmp []byte

	// sha256 hash function
	hash hash.Hash
}

// NewHasher creates a new Hasher object
func NewHasher() *Hasher {
	return &Hasher{
		hash: sha256.New(),
		tmp:  make([]byte, 32),
	}
}

// NewHasher creates a new Hasher object with a custom hash function
func NewHasherWithHash(hh hash.Hash) *Hasher {
	return &Hasher{
		hash: hh,
		tmp:  make([]byte, 32),
	}
}

// Reset resets the Hasher obj
func (h *Hasher) Reset() {
	h.buf = h.buf[:0]
	h.hash.Reset()
}

func (h *Hasher) appendBytes32(b []byte) {
	h.buf = append(h.buf, b...)
	if rest := len(b) % 32; rest != 0 {
		// pad zero bytes to the left
		h.buf = append(h.buf, zeroBytes[:32-rest]...)
	}
}

// PutUint64 appends a uint64 in 32 bytes
func (h *Hasher) PutUint64(i uint64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	h.appendBytes32(buf)
}

// PutUint32 appends a uint32 in 32 bytes
func (h *Hasher) PutUint32(i uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, i)
	h.appendBytes32(buf)
}

// PutUint16 appends a uint16 in 32 bytes
func (h *Hasher) PutUint16(i uint16) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, i)
	h.appendBytes32(buf)
}

// PutUint16 appends a uint16 in 32 bytes
func (h *Hasher) PutUint8(i uint8) {
	h.appendBytes32([]byte{byte(i)})
}

func CalculateLimit(maxCapacity, numItems, size uint64) uint64 {
	limit := (maxCapacity*size + 31) / 32
	if limit != 0 {
		return limit
	}
	if numItems == 0 {
		return 1
	}
	return numItems
}

func (h *Hasher) FillUpTo32() {
	// pad zero bytes to the left
	if rest := len(h.buf) % 32; rest != 0 {
		h.buf = append(h.buf, zeroBytes[:32-rest]...)
	}
}

func (h *Hasher) AppendUint64(i uint64) {
	h.buf = MarshalUint64(h.buf, i)
}

func (h *Hasher) Append(i []byte) {
	h.buf = append(h.buf, i...)
}

// PutRootVector appends an array of roots
func (h *Hasher) PutRootVector(b [][]byte, maxCapacity ...uint64) error {
	indx := h.Index()
	for _, i := range b {
		if len(i) != 32 {
			return fmt.Errorf("bad root")
		}
		h.buf = append(h.buf, i...)
	}

	if len(maxCapacity) == 0 {
		h.Merkleize(indx)
	} else {
		numItems := uint64(len(b))
		limit := CalculateLimit(maxCapacity[0], numItems, 32)

		h.MerkleizeWithMixin(indx, numItems, limit)
	}
	return nil
}

// PutUint64Array appends an array of uint64
func (h *Hasher) PutUint64Array(b []uint64, maxCapacity ...uint64) {
	indx := h.Index()
	for _, i := range b {
		h.AppendUint64(i)
	}

	// pad zero bytes to the left
	h.FillUpTo32()

	if len(maxCapacity) == 0 {
		// Array with fixed size
		h.Merkleize(indx)
	} else {
		numItems := uint64(len(b))
		limit := CalculateLimit(maxCapacity[0], numItems, 8)

		h.MerkleizeWithMixin(indx, numItems, limit)
	}
}

func parseBitlist(dst, buf []byte) ([]byte, uint64) {
	msb := uint8(bits.Len8(buf[len(buf)-1])) - 1
	size := uint64(8*(len(buf)-1) + int(msb))

	dst = append(dst, buf...)
	dst[len(dst)-1] &^= uint8(1 << msb)

	newLen := len(dst)
	for i := len(dst) - 1; i >= 0; i-- {
		if dst[i] != 0x00 {
			break
		}
		newLen = i
	}
	res := dst[:newLen]
	return res, size
}

// PutBitlist appends a ssz bitlist
func (h *Hasher) PutBitlist(bb []byte, maxSize uint64) {
	var size uint64
	h.tmp, size = parseBitlist(h.tmp[:0], bb)

	// merkleize the content with mix in length
	indx := h.Index()
	h.appendBytes32(h.tmp)
	h.MerkleizeWithMixin(indx, size, (maxSize+255)/256)
}

// PutBool appends a boolean
func (h *Hasher) PutBool(b bool) {
	if b {
		h.buf = append(h.buf, trueBytes...)
	} else {
		h.buf = append(h.buf, falseBytes...)
	}
}

// PutBytes appends bytes
func (h *Hasher) PutBytes(b []byte) {
	if len(b) <= 32 {
		h.appendBytes32(b)
		return
	}

	// if the bytes are longer than 32 we have to
	// merkleize the content
	indx := h.Index()
	h.appendBytes32(b)
	h.Merkleize(indx)
}

// Index marks the current buffer index
func (h *Hasher) Index() int {
	return len(h.buf)
}

// Merkleize is used to merkleize the last group of the hasher
func (h *Hasher) Merkleize(indx int) {
	input := h.buf[indx:]

	// merkleize the input
	input = h.merkleizeImpl(input[:0], input, 0)
	h.buf = append(h.buf[:indx], input...)
}

// MerkleizeWithMixin is used to merkleize the last group of the hasher
func (h *Hasher) MerkleizeWithMixin(indx int, num, limit uint64) {
	input := h.buf[indx:]

	// merkleize the input
	input = h.merkleizeImpl(input[:0], input, limit)

	// mixin with the size
	output := h.tmp[:32]
	for indx := range output {
		output[indx] = 0
	}
	MarshalUint64(output[:0], num)

	input = h.doHash(input, input, output)
	h.buf = append(h.buf[:indx], input...)
}

// HashRoot creates the hash final hash root
func (h *Hasher) HashRoot() (res [32]byte, err error) {
	if len(h.buf) != 32 {
		err = fmt.Errorf("expected 32 byte size")
		return
	}
	copy(res[:], h.buf)
	return
}

// HasherPool may be used for pooling Hashers for similarly typed SSZs.
type HasherPool struct {
	pool sync.Pool
}

// Get acquires a Hasher from the pool.
func (hh *HasherPool) Get() *Hasher {
	h := hh.pool.Get()
	if h == nil {
		return NewHasher()
	}
	return h.(*Hasher)
}

// Put releases the Hasher to the pool.
func (hh *HasherPool) Put(h *Hasher) {
	h.Reset()
	hh.pool.Put(h)
}

func nextPowerOfTwo(v uint64) uint {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return uint(v)
}

func getDepth(d uint64) uint8 {
	if d == 0 {
		return 0
	}
	if d == 1 {
		return 1
	}
	i := nextPowerOfTwo(d)
	return 64 - uint8(bits.LeadingZeros(i)) - 1
}

func (h *Hasher) doHash(dst []byte, a []byte, b []byte) []byte {
	h.hash.Write(a)
	h.hash.Write(b)
	h.hash.Sum(dst[:0])
	h.hash.Reset()
	return dst
}

func (h *Hasher) merkleizeImpl(dst []byte, input []byte, limit uint64) []byte {
	count := uint64(len(input) / 32)
	if limit == 0 {
		limit = count
	} else if count > limit {
		panic(fmt.Sprintf("BUG: count '%d' higher than limit '%d'", count, limit))
	}

	if limit == 0 {
		return append(dst, zeroBytes...)
	}
	if limit == 1 {
		if count == 1 {
			return append(dst, input[:32]...)
		}
		return append(dst, zeroBytes...)
	}

	depth := getDepth(count)
	h.merkleizeTmp = extendByteSlice(h.merkleizeTmp[:0], int(depth+2)*32)

	// reset tmp
	j := uint8(0)
	hh := h.merkleizeTmp[0:32]

	getTmp := func(i uint8) []byte {
		indx := (uint64(i) + 1) * 32
		return h.merkleizeTmp[indx : indx+32]
	}

	merge := func(i uint64, val []byte) {
		hh = append(hh[:0], val...)

		// merge back up from bottom to top, as far as we can
		for j = 0; ; j++ {
			// stop merging when we are in the left side of the next combi
			if i&(uint64(1)<<j) == 0 {
				// if we are at the count, we want to merge in zero-hashes for padding
				if i == count && j < depth {
					h.doHash(hh, hh, zeroHashes[j][:])
				} else {
					// store the merge result (may be no merge, i.e. bottom leaf node)
					copy(getTmp(j), hh)
					break
				}
			} else {
				// keep merging up if we are the right side
				h.doHash(hh, getTmp(j), hh)
			}
		}
	}

	// merge in leaf by leaf.
	for i := uint64(0); i < count; i++ {
		indx := i * 32
		merge(i, input[indx:indx+32])
	}

	// complement with 0 if empty, or if not the right power of 2
	if (uint64(1) << depth) != count {
		merge(count, zeroHashes[0][:])
	}

	// the next power of two may be smaller than the ultimate virtual size,
	// complement with zero-hashes at each depth.
	res := getTmp(depth)
	for j := depth; j < getDepth(limit); j++ {
		res = h.doHash(res, res, zeroHashes[j][:])[:32]
	}
	return append(dst, res...)
}
