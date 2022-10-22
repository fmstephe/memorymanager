package store

import "fmt"

const objectChunkSize = 1024

type ObjectStore[O any] struct {
	// Immutable fields
	chunkSize int32

	// Accounting fields
	allocs int
	frees  int

	offset   int32
	nextFree ObjectPointer[O]
	objects  [][]objectWrapper[O]
}

type objectWrapper[O any] struct {
	nextFree ObjectPointer[O]
	object   O
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
	objects := [][]objectWrapper[O]{make([]objectWrapper[O], chunkSize)}
	return &ObjectStore[O]{
		chunkSize: chunkSize,
		offset:    0,
		objects:   objects,
	}
}

func (s *ObjectStore[O]) New() (ObjectPointer[O], *O) {
	s.allocs++

	if s.nextFree.IsNil() {
		return s.newFromOffset()
	}

	return s.newFromFree()
}

func (s *ObjectStore[O]) Get(p ObjectPointer[O]) *O {
	wrapper := s.getWrapper(p)
	if !wrapper.nextFree.IsNil() {
		panic(fmt.Errorf("Attempted to Get freed object %v", p))
	}
	return &wrapper.object
}

func (s *ObjectStore[O]) Free(p ObjectPointer[O]) {
	wrapper := s.getWrapper(p)

	if !wrapper.nextFree.IsNil() {
		panic(fmt.Errorf("Attempted to Free freed object %v", p))
	}

	s.frees++

	if s.nextFree.IsNil() {
		wrapper.nextFree = p
	} else {
		wrapper.nextFree = s.nextFree
	}

	s.nextFree = p
}

func (s *ObjectStore[O]) AllocCount() int {
	return s.allocs
}

func (s *ObjectStore[O]) FreeCount() int {
	return s.frees
}

func (s *ObjectStore[O]) LiveCount() int {
	return s.allocs - s.frees
}

func (s *ObjectStore[O]) Chunks() int {
	return len(s.objects)
}

func (s *ObjectStore[O]) newFromFree() (ObjectPointer[O], *O) {
	oldFree := s.nextFree

	freeWrapper := s.getWrapper(oldFree)
	s.nextFree = freeWrapper.nextFree
	freeWrapper.nextFree = ObjectPointer[O]{}
	return oldFree, &freeWrapper.object
}

func (s *ObjectStore[O]) newFromOffset() (ObjectPointer[O], *O) {
	chunk := int32(len(s.objects))
	s.offset++
	offset := s.offset
	if s.offset == s.chunkSize {
		// Create a new chunk
		s.objects = append(s.objects, make([]objectWrapper[O], s.chunkSize))
		s.offset = 0
	}
	return ObjectPointer[O]{
		chunk:  chunk,
		offset: offset,
	}, &s.objects[chunk-1][offset-1].object
}

func (s *ObjectStore[O]) getWrapper(p ObjectPointer[O]) *objectWrapper[O] {
	return &s.objects[p.chunk-1][p.offset-1]
}
