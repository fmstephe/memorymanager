package store

import (
	"fmt"
)

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

const maxChunkSize = (1 << 23) // Smaller slot sizes should all have 8 MB chunks

func newByteSlab(slotSize uint32) byteSlab {
	// NB: This default value only applies to slotSize == 0
	slotCount := uint32(1024)

	if slotSize != 0 {
		slotCount = maxChunkSize / slotSize
	}
	// Special case for slot sizes greater than maxChunkSize - each chunk is a single slot
	if slotSize >= maxChunkSize {
		slotCount = 1
	}

	return byteSlab{
		// Immutable fields
		slotSize:  slotSize,
		slotCount: slotCount,
		chunkSize: slotSize * slotCount,

		// Mutable fields
		slotOffset: slotCount, // By initialising this, we force the creation of a new chunk on first alloc
		byteOffset: slotSize * slotCount,
		meta:       [][]bytesMeta{},
		bytes:      [][]byte{},
	}
}

func (s *byteSlab) alloc(size uint32) (BytePointer, error) {
	if s.nextFree.IsNil() {
		return s.allocFromOffset(size)
	}
	return s.allocFromFree(size)
}

func (s *byteSlab) get(p BytePointer) []byte {
	m := s.getMeta(p)
	if !m.nextFree.IsNil() {
		panic(fmt.Errorf("Attempted to Get freed bytes %v", p))
	}

	return s.getBytes(p)
}

func (s *byteSlab) free(p BytePointer) {
	meta := s.getMeta(p)

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("Attempted to Free freed object %v", p))
	}

	if s.nextFree.IsNil() {
		meta.nextFree = p
	} else {
		meta.nextFree = s.nextFree
	}

	s.nextFree = p
}

func (s *byteSlab) chunks() int {
	return len(s.bytes)
}

func (s *byteSlab) allocFromFree(size uint32) (BytePointer, error) {
	// Get pointer to the next available freed slot
	alloc := s.nextFree

	// Grab the meta-data for the slot and nil out the, now
	// allocated, slot's nextFree pointer
	freeMeta := s.getMeta(alloc)
	nextFree := freeMeta.nextFree
	freeMeta.nextFree = BytePointer{}

	// If the nextFree pointer points to the allocated slot, then
	// there are no more freed slots available
	s.nextFree = nextFree
	if nextFree == alloc {
		s.nextFree = BytePointer{}
	}

	// Set the size to properly reflect the new allocation
	alloc.size = size
	return alloc, nil
}

func (s *byteSlab) allocFromOffset(size uint32) (BytePointer, error) {
	if size > s.slotSize {
		panic(fmt.Errorf("bad alloc size, max allowed %d, %d was requested", s.slotSize-4, size))
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
