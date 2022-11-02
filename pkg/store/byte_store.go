package store

import (
	"math"
	"math/bits"
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
	return &ByteStore{
		slabs: initialiseSlabs(),
	}
}

func initialiseSlabs() []byteSlab {
	slabs := make([]byteSlab, 34)

	allocSize := uint32(1)
	for i := range slabs {
		// Special case for 0 size slab allocations
		if i == 0 {
			slabs[0] = newByteSlab(0)
			continue
		}
		// Special case for allocations greater than 2^31 In principal
		// we would want this slab to be sized 2^32, but with 32 bits
		// that's 0, so we get as close as we can.
		if i == 33 {
			slabs[33] = newByteSlab(math.MaxUint32)
			continue
		}
		slabs[i] = newByteSlab(allocSize)
		allocSize = allocSize << 1
	}

	return slabs
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

func (s *ByteStore) Chunks() int {
	chunks := 0
	for i := range s.slabs {
		chunks += s.slabs[i].chunks()
	}
	return chunks
}

func indexForSize(size uint32) int {
	if size == 0 {
		return 0
	}
	return bits.Len32(size-1) + 1
}
