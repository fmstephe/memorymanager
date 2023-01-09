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
	rootFree Pointer[O]
	meta     [][]meta[O]
	objects  [][]O
}

// If the meta for an object has a non-nil nextFree pointer then the
// object is currently free.  Object's which have never been allocated are
// implicitly free, but have a nil nextFree point in their meta.
type meta[O any] struct {
	nextFree Pointer[O]
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

func (s *Store[O]) Alloc() (Pointer[O], *O) {
	s.allocs++

	if s.rootFree.IsNil() {
		return s.newFromOffset()
	}

	s.reused++
	return s.newFromFree()
}

func (s *Store[O]) Get(p Pointer[O]) *O {
	m := s.getMeta(p)
	if !m.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Get freed object %v", p))
	}
	return s.getObject(p)
}

func (s *Store[O]) Free(p Pointer[O]) {
	meta := s.getMeta(p)

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed object %v", p))
	}

	s.frees++

	if s.rootFree.IsNil() {
		meta.nextFree = p
	} else {
		meta.nextFree = s.rootFree
	}

	s.rootFree = p
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

func (s *Store[O]) newFromFree() (Pointer[O], *O) {
	oldFree := s.rootFree

	freeMeta := s.getMeta(oldFree)
	s.rootFree = freeMeta.nextFree
	freeMeta.nextFree = Pointer[O]{}
	return oldFree, s.getObject(oldFree)
}

func (s *Store[O]) newFromOffset() (Pointer[O], *O) {
	chunk := uint32(len(s.objects))
	s.offset++
	offset := s.offset
	if s.offset == s.chunkSize {
		// Create a new chunk
		s.meta = append(s.meta, make([]meta[O], s.chunkSize))
		s.objects = append(s.objects, make([]O, s.chunkSize))
		s.offset = 0
	}
	return Pointer[O]{
		chunk:  chunk,
		offset: offset,
	}, &s.objects[chunk-1][offset-1]
}

func (s *Store[O]) getObject(p Pointer[O]) *O {
	return &s.objects[p.chunk-1][p.offset-1]
}

func (s *Store[O]) getMeta(p Pointer[O]) *meta[O] {
	return &s.meta[p.chunk-1][p.offset-1]
}
