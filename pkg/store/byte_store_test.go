package store

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Demonstrate that we are initialising the slabs with the correct indices i.e.
// that the indexForSize(slotSize) returns the correct index for that slab
func TestSlabIndices_CorrectlyOrdered(t *testing.T) {
	slabs := initialiseSlabs()
	for i, slab := range slabs {
		assert.Equal(t, i, indexForSize(slab.slotSize))
	}
}

// Demonstrate that allocations which are not the same as the exact slot size
// of a slab are still indexed to the correct slab.  An allocation is indexed
// correctly if it is less than or equal to the slot size of the slab it was
// indexed to, as well as being strictly larger than the slab of a lower index.
func TestSlabIndices_SmallerAllocations(t *testing.T) {
	slabs := initialiseSlabs()
	sizes := generateAllocationSizes()
	for _, size := range sizes {
		idx := indexForSize(size)
		slab := slabs[idx]
		// The size must be less than or equal to the slot size, or this slab can't allocate it
		assert.LessOrEqual(t, size, slab.slotSize, "%d, %d, %#v", size, idx, slab)
		if idx > 0 {
			smallerSlab := slabs[idx-1]
			assert.Greater(t, size, smallerSlab.slotSize)
		}
	}
}

// Because evenly distributed random numbers tend to mostly be very large we
// artificially generate small ones here to test the full range of allocation
// sizes
func generateAllocationSizes() []uint32 {
	sizes := []uint32{}
	for i := 0; i < 1000; i++ {
		size := rand.Uint32()
		sizes = append(sizes, size)
		for ; size > 0; size = size >> 1 {
			sizes = append(sizes, size)
		}
	}
	return sizes
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

	sliceAllocations := 10_000

	// Create a large number of objects
	slices := make([]BytePointer, sliceAllocations)
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
