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

func (s *ByteStore) New(data []byte) (BytePointer, error) {
	// Here we check the size of data at word size (likely int64)
	// Because the fields are always int32 if we pass this step, int32 is always large enough
	if len(data) > int(s.chunkSize) {
		return BytePointer{}, fmt.Errorf("data of size %d, is greater than object size limit %d", len(data), s.chunkSize)
	}

	size := int32(len(data))

	// If current chunk is too small, move to the next one
	if s.offset+size > s.chunkSize {
		s.offset = 0
		s.bytes = append(s.bytes, make([]byte, s.chunkSize))
	}

	copy(s.bytes[len(s.bytes)-1][s.offset:], data)

	// Create BytePointer pointing recently copied data
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
	return s.bytes[obj.chunk-1][obj.offset-1 : obj.size]
}
