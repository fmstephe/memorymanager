package store

import "fmt"

type BytePointer struct {
	chunk  int32
	offset int32
	size   int32
}

func (p BytePointer) IsNil() bool {
	return p.chunk == 0 && p.offset == 0
}

type ByteStore struct {
	// Immutable fields
	chunkSize int32

	// Mutable fields
	offset int32
	bytes  [][]byte
}

func NewByteStore(chunkSize int32) *ByteStore {
	// Initialise bytes with a single empty chunk available
	bytes := [][]byte{make([]byte, chunkSize)}

	return &ByteStore{
		chunkSize: chunkSize,
		offset:    0,
		bytes:     bytes,
	}
}

func (s *ByteStore) New(size int32) (BytePointer, error) {
	if size > s.chunkSize {
		return BytePointer{}, fmt.Errorf("data of size %d, is greater than object size limit %d", size, s.chunkSize)
	}

	// If current chunk is too small, create a new chunk and use that
	if s.offset+size > s.chunkSize {
		s.offset = 0
		s.bytes = append(s.bytes, make([]byte, s.chunkSize))
	}

	// Create BytePointer pointing to the new slice
	obj := BytePointer{
		chunk:  int32(len(s.bytes)),
		offset: s.offset + 1,
		size:   size,
	}

	// Update offset
	s.offset += size

	return obj, nil
}

func (s *ByteStore) Get(obj BytePointer) []byte {
	// There are no pre-checks here - if you pass in a malformed BytePointer
	// this method may return nonsense or just panic

	// Note the chunk and offset are always 1 greater than their actual
	// values so we subtract 1 from them before use.  They are 1 greater to
	// allow a pointer with zero values to represent 'nil'
	chunk := obj.chunk - 1
	offset := obj.offset - 1

	// Grab the actual byte values.  It's worth noting here that a pointer
	// can have size 0, in this case a zero length slice is returned
	bytes := s.bytes[chunk][offset : offset+obj.size]
	return bytes
}
