package store

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Assert we can generate reliable byteSlab indexes for all sizes of
// allocation Once we have this size->index mapping we can build separate
// byteSlabs for each allocation size range.
func TestIndexForSize(t *testing.T) {
	// Allocations of zero and one sized slices are handled by the one sized slab
	assert.Equal(t, uint32(0), indexForSize(0))
	assert.Equal(t, uint32(0), indexForSize(1))

	// Two is a lonely power of two group
	assert.Equal(t, uint32(1), indexForSize(2))

	// Four is also a small power of two group
	assert.Equal(t, uint32(2), indexForSize(3))
	assert.Equal(t, uint32(2), indexForSize(4))

	for i := uint32(5); i <= 8; i++ {
		assert.Equal(t, uint32(3), indexForSize(i))
	}

	for i := uint32(9); i <= 16; i++ {
		assert.Equal(t, uint32(4), indexForSize(i))
	}

	for i := uint32(17); i <= 32; i++ {
		assert.Equal(t, uint32(5), indexForSize(i))
	}

	for i := uint32(33); i <= 64; i++ {
		assert.Equal(t, uint32(6), indexForSize(i))
	}

	for i := uint32(65); i <= 128; i++ {
		assert.Equal(t, uint32(7), indexForSize(i))
	}

	for i := uint32(129); i <= 256; i++ {
		assert.Equal(t, uint32(8), indexForSize(i))
	}

	for i := uint32(257); i <= 512; i++ {
		assert.Equal(t, uint32(9), indexForSize(i))
	}

	for i := uint32(513); i <= 1024; i++ {
		assert.Equal(t, uint32(10), indexForSize(i))
	}

	// Hopefully these are enough cases, it's hard to write a generic
	// test without just writing out the indexForSize method again
}

// Demonstrate that we can create bytes, then get those bytes and modify them
// we can then get those bytes again and will see the modification We ensure
// that we allocate so many bytes that we will need more than one chunk to
// store all bytes.
func Test_Bytes_GetModifyGet(t *testing.T) {
	const chunkSize = 1024
	bs := NewByteStore()

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
// we can then get those bytes again and will see the modification.  We ensure
// that we allocate so many bytes that we will need more than one chunk to
// store all bytes.
func Test_Bytes_GetModifyGet_OddSizing(t *testing.T) {
	const chunkSize = 1024
	bs := NewByteStore()

	// Create all the byte slices
	pointers := make([]BytePointer, chunkSize*3)
	size := uint32(0)
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

// Demonstrate that we can create an object, then free it. If we try to Get()
// the freed object BytesStore panics
func Test_Bytes_NewFreeGet_Panic(t *testing.T) {
	os := NewByteStore()
	p, _ := os.New(8)
	os.Free(p)

	assert.Panics(t, func() { os.Get(p) })
}

// Demonstrate that we can create an object, then free it. If we try to Free()
// the freed object BytesStore panics
func Test_Bytes_NewFreeFree_Panic(t *testing.T) {
	os := NewByteStore()
	p, _ := os.New(8)
	os.Free(p)

	assert.Panics(t, func() { os.Free(p) })
}

// Demonstrate that if we create a large number of objects, then free them,
// then allocate that same number again, we re-use the freed objects
func Test_Bytes_NewFreeNew_ReusesOldBytes(t *testing.T) {
	os := NewByteStore()

	sliceAllocations := slotCountSize * 3

	// Create a large number of objects
	slices := make([]BytePointer, slotCountSize*3)
	for i := range slices {
		p, _ := os.New(uint32(i))
		slices[i] = p
	}

	// We have allocate one batch of objects
	assert.Equal(t, sliceAllocations, os.AllocCount())
	// They are all live
	assert.Equal(t, sliceAllocations, os.LiveCount())
	// Nothing has been freed
	assert.Equal(t, 0, os.FreeCount())

	// Free all of those slices
	for _, p := range slices {
		os.Free(p)
	}

	// We have allocate one batch of slices
	assert.Equal(t, sliceAllocations, os.AllocCount())
	// None are live
	assert.Equal(t, 0, os.LiveCount())
	// We have freed one batch of slices
	assert.Equal(t, sliceAllocations, os.FreeCount())

	// Allocate the same number of slices again
	for i := range slices {
		os.New(uint32(i))
	}

	// We have allocated 2 batches of slices
	assert.Equal(t, 2*sliceAllocations, os.AllocCount())
	// We have freed one batch
	assert.Equal(t, sliceAllocations, os.LiveCount())
	// One batch is live
	assert.Equal(t, sliceAllocations, os.FreeCount())
}

func intToBytes(value int) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(value))
	return bytes
}

func bytesToInt(bytes []byte) int {
	return int(binary.LittleEndian.Uint64(bytes))
}
