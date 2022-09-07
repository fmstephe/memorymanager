package bytestore

import "fmt"

type Object struct {
	chunk  int32
	offset int32
	size   int32
}

type ByteStore struct {
	// Immutable fields
	chunkSize int32
	maxChunks int32

	// Mutable fields
	offset int32
	chunks [][]byte
}

func NewByteStore(chunkSize, maxChunks int32) *ByteStore {
	// Initialise chunks with a single empty chunk available
	chunks := make([][]byte, 0, maxChunks)
	chunks = append(chunks, make([]byte, chunkSize))

	return &ByteStore{
		chunkSize: chunkSize,
		maxChunks: maxChunks,
		offset:    0,
		chunks:    chunks,
	}
}

func (s *ByteStore) Set(data []byte) (Object, error) {
	// Ensure that we haven't already reached the limit of
	if int32(len(s.chunks)) > s.maxChunks {
		return Object{}, fmt.Errorf("size limit reached for bytes store")
	}

	// Here we check the size of data at word size (likely int64)
	// Because the fields are always int32 if we pass this step, int32 is always large enough
	if len(data) > int(s.chunkSize) {
		return Object{}, fmt.Errorf("data of size %d, is greater than object size limit %d", len(data), s.chunkSize)
	}

	size := int32(len(data))

	// If current chunk is too small, move to the next one
	if s.offset+size > s.chunkSize {
		s.offset = 0
		s.chunks = append(s.chunks, make([]byte, s.chunkSize))

		// Check that we haven't pushed past the chunk limit
		if int32(len(s.chunks)) > s.maxChunks {
			return Object{}, fmt.Errorf("size limit reached for bytes store")
		}
	}

	copy(s.chunks[len(s.chunks)-1][s.offset:], data)

	// Create Object pointing recently copied data
	obj := Object{
		chunk:  int32(len(s.chunks) - 1),
		offset: s.offset,
		size:   size,
	}

	// Update offset
	s.offset += size

	return obj, nil
}

func (s *ByteStore) Get(obj Object) []byte {
	// There are no pre-checks here - if you pass in a malformed Object
	// this method may return nonsense or just panic
	return s.chunks[obj.chunk][obj.offset:obj.size]
}
