package objectstore

import (
	"fmt"
)

const objectChunkSize = 1024

type Stats struct {
	Allocs    int
	Frees     int
	RawAllocs int
	Live      int
	Reused    int
	Chunks    int
}

type Store[O any] struct {
	// Immutable fields
	chunkSize uint32

	// Accounting fields
	allocs int
	frees  int
	reused int

	// Data fields
	offset   uint32
	rootFree Reference[O]
	meta     [][]meta[O]
	objects  [][]O
}

// If the meta for an object has a non-nil nextFree pointer then the
// object is currently free.  Object's which have never been allocated are
// implicitly free, but have a nil nextFree point in their meta.
type meta[O any] struct {
	nextFree Reference[O]
}

func New[O any]() *Store[O] {
	chunkSize := uint32(objectChunkSize)
	// Initialise the first chunk
	meta := [][]meta[O]{make([]meta[O], chunkSize)}
	objects := [][]O{make([]O, chunkSize)}
	return &Store[O]{
		chunkSize: chunkSize,
		offset:    0,
		meta:      meta,
		objects:   objects,
	}
}

func (s *Store[O]) Alloc() (Reference[O], *O) {
	s.allocs++

	if s.rootFree.IsNil() {
		return s.allocFromOffset()
	}

	s.reused++
	return s.allocFromFree()
}

func (s *Store[O]) Get(r Reference[O]) *O {
	m := s.getMeta(r)
	if !m.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Get freed object %v", r))
	}
	return s.getObject(r)
}

func (s *Store[O]) Free(r Reference[O]) {
	meta := s.getMeta(r)

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed object %v", r))
	}

	s.frees++

	if s.rootFree.IsNil() {
		meta.nextFree = r
	} else {
		meta.nextFree = s.rootFree
	}

	s.rootFree = r
}

func (s *Store[O]) GetStats() Stats {
	return Stats{
		Allocs:    s.allocs,
		Frees:     s.frees,
		RawAllocs: s.allocs - s.reused,
		Live:      s.allocs - s.frees,
		Reused:    s.reused,
		Chunks:    len(s.objects),
	}
}

func (s *Store[O]) allocFromFree() (Reference[O], *O) {
	// Get pointer to the next available freed slot
	alloc := s.rootFree

	// Grab the meta-data for the slot and nil out the, now
	// allocated, slot's nextFree pointer
	freeMeta := s.getMeta(alloc)
	nextFree := freeMeta.nextFree
	freeMeta.nextFree = Reference[O]{}

	// If the nextFree pointer points to the just allocated slot, then
	// there are no more freed slots available
	s.rootFree = nextFree
	if nextFree == alloc {
		s.rootFree = Reference[O]{}
	}

	return alloc, s.getObject(alloc)
}

func (s *Store[O]) allocFromOffset() (Reference[O], *O) {
	chunk := uint32(len(s.objects))
	s.offset++
	offset := s.offset
	if s.offset == s.chunkSize {
		// Create a new chunk
		s.meta = append(s.meta, make([]meta[O], s.chunkSize))
		s.objects = append(s.objects, make([]O, s.chunkSize))
		s.offset = 0
	}
	return Reference[O]{
		chunk:  chunk,
		offset: offset,
	}, &s.objects[chunk-1][offset-1]
}

func (s *Store[O]) getObject(r Reference[O]) *O {
	return &s.objects[r.chunk-1][r.offset-1]
}

func (s *Store[O]) getMeta(r Reference[O]) *meta[O] {
	return &s.meta[r.chunk-1][r.offset-1]
}
