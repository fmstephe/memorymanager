package store

import (
	"fmt"
	"math/bits"
)

// Convenience constants to make it easier to declare large max object size values
const (
	_         = iota // ignore first value by assigning to blank identifier
	KB uint32 = 1 << (10 * iota)
	MB
	GB
)

type BytePointer struct {
	chunk      uint32
	slotOffset uint32
	byteOffset uint32
	size       uint32
}

func (p BytePointer) IsNil() bool {
	return p.chunk == 0 && p.slotOffset == 0
}

type ByteStore struct {
	allocs int
	frees  int
	slabs  []byteSlab
}

func NewByteStore() *ByteStore {
	slabs := make([]byteSlab, 16)
	allocSize := uint32(1)
	for range slabs {
		idx := indexForSize(allocSize)
		slabs[idx] = newByteSlab(allocSize)
		allocSize = allocSize << 1
	}

	return &ByteStore{
		slabs: slabs,
	}
}

// TODO rename this to Alloc
func (s *ByteStore) New(size uint32) (BytePointer, error) {
	s.allocs++

	idx := indexForSize(size)
	// TODO we should explicitly panic if the idx is out of range here
	// Need a clear and explicit panic message
	// Eventually we will probably manage large allocations separately
	return s.slabs[idx].alloc(size)
}

func (s *ByteStore) Get(p BytePointer) []byte {
	idx := indexForSize(p.size)
	return s.slabs[idx].get(p)
}

func (s *ByteStore) Free(p BytePointer) {
	s.frees++

	idx := indexForSize(p.size)
	s.slabs[idx].free(p)
}

func (s *ByteStore) AllocCount() int {
	return s.allocs
}

func (s *ByteStore) FreeCount() int {
	return s.frees
}

func (s *ByteStore) LiveCount() int {
	return s.allocs - s.frees
}

func indexForSize(size uint32) uint32 {
	if size == 0 {
		return 0
	}
	return uint32(bits.Len32(size - 1))
}

const slotCountSize = 1024

// If the bytesMeta for a byte slot has a non-nil nextFree pointer then the
// byte slot is currently free.  Byte slots which have never been allocated are
// implicitly free, but have a nil nextFree point in their bytesMeta.
type bytesMeta struct {
	nextFree BytePointer
}

type byteSlab struct {
	// Immutable fields
	slotSize  uint32 // Max size per allocation
	slotCount uint32 // Number of slots per chunk
	chunkSize uint32 // Each chunk is sized slotSize*slotCount

	// Mutable fields
	byteOffset uint32        // The offset of unallocated bytes in current chunk
	slotOffset uint32        // The offset of unallocated slots in the current chunk
	nextFree   BytePointer   // The first freed slot, may be nil
	meta       [][]bytesMeta // All meta-data
	bytes      [][]byte      // All actual byte data
}

func newByteSlab(slotSize uint32) byteSlab {
	// We hard code the number of slots per chunk here, this could be configurable
	slotCount := uint32(slotCountSize)

	return byteSlab{
		// Immutable fields
		slotSize:  slotSize,
		slotCount: slotCount,
		chunkSize: slotSize * slotCount,

		// Mutable fields
		slotOffset: slotCount, // By initialising this, we force the creation of a new chunk on first alloc
		byteOffset: slotSize * slotCount,
		bytes:      [][]byte{},
	}
}

func (s *byteSlab) alloc(size uint32) (BytePointer, error) {
	if s.nextFree.IsNil() {
		return s.allocFromOffset(size)
	}
	return s.allocFromFree(size)
}

func (s *byteSlab) allocFromFree(size uint32) (BytePointer, error) {
	oldFree := s.nextFree

	freeMeta := s.getMeta(oldFree)
	s.nextFree = freeMeta.nextFree
	freeMeta.nextFree = BytePointer{}

	oldFree.size = size
	return oldFree, nil
}

func (s *byteSlab) allocFromOffset(size uint32) (BytePointer, error) {
	if size > s.slotSize {
		panic(fmt.Errorf("Bad alloc size, max allowed %d, %d was requested", s.slotSize-4, size))
	}

	// If we have used up the last chunk create a new one
	if s.slotOffset == s.slotCount {
		s.slotOffset = 0
		s.byteOffset = 0
		s.meta = append(s.meta, make([]bytesMeta, s.slotCount))
		s.bytes = append(s.bytes, make([]byte, s.chunkSize))
	}

	// Create BytePointer pointing to the new slice
	p := BytePointer{
		chunk:      uint32(len(s.bytes)),
		slotOffset: s.slotOffset + 1,
		byteOffset: s.byteOffset + 1,
		size:       size,
	}

	// Update offset
	s.slotOffset++
	s.byteOffset += s.slotSize

	return p, nil
}

func (s *byteSlab) get(p BytePointer) []byte {
	m := s.getMeta(p)
	if !m.nextFree.IsNil() {
		panic(fmt.Errorf("Attempted to Get freed bytes %v", p))
	}

	return s.getBytes(p)
}

func (s *byteSlab) getBytes(p BytePointer) []byte {
	chunk := p.chunk - 1
	offset := p.byteOffset - 1
	size := p.size
	return s.bytes[chunk][offset : offset+size]
}

func (s *byteSlab) getMeta(p BytePointer) *bytesMeta {
	chunk := p.chunk - 1
	offset := p.slotOffset - 1
	return &s.meta[chunk][offset]
}

func (s *byteSlab) free(p BytePointer) {
	meta := s.getMeta(p)

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("Attempted to Free freed object %v", p))
	}

	//s.frees++

	if s.nextFree.IsNil() {
		meta.nextFree = p
	} else {
		meta.nextFree = s.nextFree
	}

	s.nextFree = p
}
