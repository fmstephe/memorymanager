package store

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Demonstrate that we can create bytes, then get those bytes and modify them
// we can then get those bytes again and will see the modification
// We ensure that we allocate so many bytes that we will need more than one chunk
// to store all bytes.
func Test_Bytes_GetModifyGet(t *testing.T) {
	const chunkSize = 1024
	bs := NewByteStore(chunkSize)

	// Create all the byte slices
	pointers := make([]BytePointer, chunkSize*3)
	for i := range pointers {
		p, err := bs.New(8)
		require.NoError(t, err)
		pointers[i] = p
	}

	// Get the bytes from the store and write data into it
	for i, p := range pointers {
		bytes := bs.Get(p)
		intBytes := intToBytes(i)
		copy(bytes, intBytes)
	}

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		bytes := bs.Get(p)
		value := bytesToInt(bytes)
		assert.Equal(t, i, value)
	}
}

// Demonstrate that we can create bytes, then get those bytes and modify them
// we can then get those bytes again and will see the modification
// We ensure that we allocate so many bytes that we will need more than one chunk
// to store all bytes.
func Test_Bytes_GetModifyGet_OddSizing(t *testing.T) {
	const chunkSize = 1024
	bs := NewByteStore(chunkSize)

	// Create all the byte slices
	pointers := make([]BytePointer, chunkSize*3)
	size := int32(0)
	for i := range pointers {
		p, err := bs.New(size)
		size++
		if size > chunkSize {
			size = 0
		}
		require.NoError(t, err)
		pointers[i] = p
	}

	// Get the bytes from the store and write data into it
	for i, p := range pointers {
		bytes := bs.Get(p)

		// Write value into the bytes
		for j := range bytes {
			bytes[j] = byte(i)
		}
	}

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		bytes := bs.Get(p)

		// Write value into the expected bytes
		expectedBytes := make([]byte, len(bytes))
		for j := range expectedBytes {
			expectedBytes[j] = byte(i)
		}

		assert.Equal(t, expectedBytes, bytes)
	}
}

func intToBytes(value int) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(value))
	return bytes
}

func bytesToInt(bytes []byte) int {
	return int(binary.LittleEndian.Uint64(bytes))
}
