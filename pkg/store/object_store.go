package store

import "fmt"

const objectChunkSize = 1024

type ObjectStoreStats struct {
	Allocs    int
	Frees     int
	RawAllocs int
	Live      int
	Reused    int
	Chunks    int
}

type ObjectStore[O any] struct {
	// Immutable fields
	chunkSize int32

	// Accounting fields
	allocs int
	frees  int
	reused int

	offset   int32
	nextFree ObjectPointer[O]
	meta     [][]objectMeta[O]
	objects  [][]O
}

// If the objectMeta for an object has a non-nil nextFree pointer then the
// object is currently free.  Object's which have never been allocated are
// implicitly free, but have a nil nextFree point in their objectMeta.
type objectMeta[O any] struct {
	nextFree ObjectPointer[O]
}

type ObjectPointer[O any] struct {
	chunk  int32
	offset int32
}

func (p ObjectPointer[O]) IsNil() bool {
	return p.chunk == 0 && p.offset == 0
}

func NewObjectStore[O any]() *ObjectStore[O] {
	chunkSize := int32(objectChunkSize)
	// Initialise the first chunk
	meta := [][]objectMeta[O]{make([]objectMeta[O], chunkSize)}
	objects := [][]O{make([]O, chunkSize)}
	return &ObjectStore[O]{
		chunkSize: chunkSize,
		offset:    0,
		meta:      meta,
		objects:   objects,
	}
}

func (s *ObjectStore[O]) Alloc() (ObjectPointer[O], *O) {
	s.allocs++

	if s.nextFree.IsNil() {
		return s.newFromOffset()
	}

	s.reused++
	return s.newFromFree()
}

func (s *ObjectStore[O]) Get(p ObjectPointer[O]) *O {
	m := s.getMeta(p)
	if !m.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Get freed object %v", p))
	}
	return s.getObject(p)
}

func (s *ObjectStore[O]) Free(p ObjectPointer[O]) {
	meta := s.getMeta(p)

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed object %v", p))
	}

	s.frees++

	if s.nextFree.IsNil() {
		meta.nextFree = p
	} else {
		meta.nextFree = s.nextFree
	}

	s.nextFree = p
}

func (s *ObjectStore[O]) GetStats() ObjectStoreStats {
	return ObjectStoreStats{
		Allocs:    s.allocs,
		Frees:     s.frees,
		RawAllocs: s.allocs - s.reused,
		Live:      s.allocs - s.frees,
		Reused:    s.reused,
		Chunks:    len(s.objects),
	}
}

func (s *ObjectStore[O]) newFromFree() (ObjectPointer[O], *O) {
	oldFree := s.nextFree

	freeMeta := s.getMeta(oldFree)
	s.nextFree = freeMeta.nextFree
	freeMeta.nextFree = ObjectPointer[O]{}
	return oldFree, s.getObject(oldFree)
}

func (s *ObjectStore[O]) newFromOffset() (ObjectPointer[O], *O) {
	chunk := int32(len(s.objects))
	s.offset++
	offset := s.offset
	if s.offset == s.chunkSize {
		// Create a new chunk
		s.meta = append(s.meta, make([]objectMeta[O], s.chunkSize))
		s.objects = append(s.objects, make([]O, s.chunkSize))
		s.offset = 0
	}
	return ObjectPointer[O]{
		chunk:  chunk,
		offset: offset,
	}, &s.objects[chunk-1][offset-1]
}

func (s *ObjectStore[O]) getObject(p ObjectPointer[O]) *O {
	return &s.objects[p.chunk-1][p.offset-1]
}

func (s *ObjectStore[O]) getMeta(p ObjectPointer[O]) *objectMeta[O] {
	return &s.meta[p.chunk-1][p.offset-1]
}
