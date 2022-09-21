package store

const objectChunkSize = 1024

type ObjectStore[O any] struct {
	// Immutable fields
	chunkSize int32

	offset  int32
	objects [][]O
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
	objects := [][]O{make([]O, chunkSize)}
	return &ObjectStore[O]{
		chunkSize: chunkSize,
		offset:    0,
		objects:   objects,
	}
}

func (s *ObjectStore[O]) New() (ObjectPointer[O], *O) {
	chunk := int32(len(s.objects))
	s.offset++
	offset := s.offset
	if s.offset == s.chunkSize {
		// Create a new chunk
		s.objects = append(s.objects, make([]O, s.chunkSize))
		s.offset = 0
	}
	return ObjectPointer[O]{
		chunk:  chunk,
		offset: offset,
	}, &s.objects[chunk-1][offset-1]
}

func (s *ObjectStore[O]) Get(p ObjectPointer[O]) *O {
	return &s.objects[p.chunk-1][p.offset-1]
}
