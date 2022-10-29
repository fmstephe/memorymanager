package store

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

type BytePointer struct {
	chunk  uint32
	offset uint32
	size   uint32
}

func (p BytePointer) IsNil() bool {
	return p.chunk == 0 && p.offset == 0
}

type ByteStore struct {
	slabs []byteSlab
}

func NewByteStore(_ int32) *ByteStore {
	slabs := make([]byteSlab, 16)
	totalSize := uint32(8)
	for range slabs {
		allocSize := totalSize - 4
		idx := indexForSize(allocSize)
		slabs[idx] = newByteSlab(allocSize, 2<<10)
		totalSize *= 2
	}

	return &ByteStore{
		slabs: slabs,
	}
}

func (s *ByteStore) New(size uint32) (BytePointer, error) {
	idx := indexForSize(size)
	return s.slabs[idx].alloc(size)
}

func (s *ByteStore) Get(obj BytePointer) []byte {
	idx := indexForSize(obj.size)
	return s.slabs[idx].get(obj)
}

func indexForSize(size uint32) uint32 {
	size += zeroToOne(size) // Use this to make 0 -> 1
	slotSize := size + 4
	count := uint32(bits.Len32(nextPowerOf2(slotSize)))
	return count - 4
}

// This oddity converts zero to 1, other number are unaffected
func zeroToOne(val uint32) uint32 {
	val |= val >> 1
	val = val &^ (1 << 31)
	val = val - 1
	val = val >> 31
	return val
}

func nextPowerOf2(val uint32) uint32 {
	// Check if val _is_ a power of 2
	if val > 0 && val&(val-1) == 0 {
		return val
	}
	return 1 << bits.Len32(val)
}

type byteSlab struct {
	// Immutable fields
	slotBits  uint32 // Debugging only
	allocSize uint32
	slotSize  uint32
	slotBatch uint32
	chunkSize uint32

	// Mutable fields
	offset uint32
	bytes  [][]byte
}

// Convenience constants to make it easier to declare large max object size values
const (
	_         = iota // ignore first value by assigning to blank identifier
	KB uint32 = 1 << (10 * iota)
	MB
	GB
)

func newByteSlab(allocSize, slotBatch uint32) byteSlab {
	slotSize := allocSize + 4
	chunkSize := slotSize * slotBatch

	// Initialise bytes with a single empty chunk available
	bytes := [][]byte{make([]byte, chunkSize)}

	return byteSlab{
		slotBits:  uint32(bits.Len32(slotSize)),
		slotSize:  slotSize,
		allocSize: allocSize,
		slotBatch: slotBatch,
		chunkSize: chunkSize,
		offset:    0,
		bytes:     bytes,
	}
}

func (s *byteSlab) alloc(size uint32) (BytePointer, error) {
	if size > s.allocSize {
		panic(fmt.Errorf("Bad alloc size, max allowed %d, %d was requested", s.slotSize-4, size))
	}
	// If we have used up the last chunk create a new one
	if s.offset == s.chunkSize {
		s.offset = 0
		s.bytes = append(s.bytes, make([]byte, s.chunkSize))
	}

	s.writeSize(uint32(len(s.bytes))-1, s.offset, size)

	// Create BytePointer pointing to the new slice
	obj := BytePointer{
		chunk:  uint32(len(s.bytes)),
		offset: s.offset + 1,
		size:   size,
	}

	// Update offset
	s.offset += s.chunkSize

	return obj, nil
}

func (s *byteSlab) writeSize(chunk, offset, size uint32) {
	bytes := s.bytes[chunk][offset : offset+s.slotSize]
	binary.LittleEndian.PutUint32(bytes, uint32(size))
}

func (s *byteSlab) read(chunk, offset uint32) []byte {
	bytes := s.bytes[chunk][offset : offset+s.slotSize]
	size := binary.LittleEndian.Uint32(bytes)
	return bytes[4 : 4+size]
}

func (s *byteSlab) get(obj BytePointer) []byte {
	// There are no pre-checks here - if you pass in a malformed BytePointer
	// this method may return nonsense or just panic

	// Note the chunk and offset are always 1 greater than their actual
	// values so we subtract 1 from them before use.  They are 1 greater to
	// allow a pointer with zero values to represent 'nil'
	chunk := obj.chunk - 1
	offset := obj.offset - 1

	bytes := s.read(chunk, offset)
	return bytes
}
