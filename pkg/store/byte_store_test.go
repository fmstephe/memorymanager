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
		// Allocate the slice
		p, err := bs.Alloc(8)
		require.NoError(t, err)

		// Get the slice and write some data into it
		bytes := bs.Get(p)
		binary.LittleEndian.PutUint64(bytes, uint64(i))

		// Collect the pointer
		pointers[i] = p
	}

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		// Get the slice
		bytes := bs.Get(p)
		// Read its data
		value := int(binary.LittleEndian.Uint64(bytes))
		assert.Equal(t, i, value)
	}

	// Free all the allocated slices
	for _, p := range pointers {
		bs.Free(p)
	}

	// Get the bytes from the store and write data into it
	for i := range pointers {
		// Allocate the slice
		p, err := bs.Alloc(8)
		require.NoError(t, err)

		// Get the slice and write some data into it, make sure the
		// data is different from anything written earlier
		bytes := bs.Get(p)
		binary.LittleEndian.PutUint64(bytes, uint64(i<<10))

		// Collect the pointer
		pointers[i] = p
	}

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		// Get the slice
		bytes := bs.Get(p)
		// Read its data
		value := int(binary.LittleEndian.Uint64(bytes))
		assert.Equal(t, i<<10, value)
	}
}

// This small test is in response to a bug found in the free implementation The
// bug was that there is a loop in the `nextFree` of the last freed slot in
// each byteSlab.  This is because a freed slot must always have a non-nil
// `nextFree` pointer in its meta-data.  However because we weren't checking
// for this exact case the last freed slot would re-add itself to the root
// `nextFree` pointer in the byteSlab.  This means that in this case calls to
// `Alloc()` would allocate the same slot over and over, meaning multiple
// independently allocated pointers would all point to the same slot.
func TestFreeThenAllocTwice(t *testing.T) {
	bs := NewByteStore()

	// allocate and immediately free a slice
	p, err := bs.Alloc(8)
	require.NoError(t, err)
	bs.Free(p)

	// Allocate two new slices
	p1, err := bs.Alloc(8)
	require.NoError(t, err)
	p2, err := bs.Alloc(8)
	require.NoError(t, err)

	// assert that the pointers are independent
	assert.NotEqual(t, p1, p2)

	// Write different values to the two newly allocated slices
	bytes1 := bs.Get(p1)
	binary.LittleEndian.PutUint64(bytes1, uint64(1))
	bytes2 := bs.Get(p2)
	binary.LittleEndian.PutUint64(bytes2, uint64(2))

	// assert that the writes are in the two, separate, slices
	value1 := int(binary.LittleEndian.Uint64(bytes1))
	value2 := int(binary.LittleEndian.Uint64(bytes2))
	assert.Equal(t, 1, value1)
	assert.Equal(t, 2, value2)
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
		p, err := bs.Alloc(size)
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
	p, _ := os.Alloc(8)
	os.Free(p)

	assert.Panics(t, func() { os.Get(p) })
}

// Demonstrate that we can create an object, then free it. If we try to Free()
// the freed object BytesStore panics
func Test_Bytes_NewFreeFree_Panic(t *testing.T) {
	os := NewByteStore()
	p, _ := os.Alloc(8)
	os.Free(p)

	assert.Panics(t, func() { os.Free(p) })
}

// Demonstrate that if we create a large number of objects, then free them,
// then allocate that same number again, we re-use the freed objects
func Test_Bytes_NewFreeNew_ReusesOldBytes(t *testing.T) {
	s := NewByteStore()

	sliceAllocations := 10_000

	// Create a large number of objects
	slices := make([]BytePointer, sliceAllocations)
	for i := range slices {
		p, _ := s.Alloc(uint32(i))
		slices[i] = p
	}

	// We have allocate one batch of objects
	stats := s.GetStats()
	assert.Equal(t, sliceAllocations, stats.TotalAllocs)
	// They are all live
	assert.Equal(t, sliceAllocations, stats.TotalLive)
	// Nothing has been freed
	assert.Equal(t, 0, stats.TotalFrees)
	// Nothing has been reused
	assert.Equal(t, 0, stats.TotalReused)

	chunks := stats.TotalChunks
	// Assert that there are _some_ chunks which have been used to
	// serve the allocations
	assert.Greater(t, chunks, 0)

	// Free all of those slices
	for _, p := range slices {
		s.Free(p)
	}

	// We have allocate one batch of slices
	stats = s.GetStats()
	assert.Equal(t, sliceAllocations, stats.TotalAllocs)
	// None are live
	assert.Equal(t, 0, stats.TotalLive)
	// We have freed one batch of slices
	assert.Equal(t, sliceAllocations, stats.TotalFrees)
	// Nothing has been reused
	assert.Equal(t, 0, stats.TotalReused)
	// The number of chunks hasn't changed
	assert.Equal(t, chunks, stats.TotalChunks)

	// Allocate the same number of slices again
	for i := range slices {
		s.Alloc(uint32(i))
	}

	// We have allocated 2 batches of slices
	stats = s.GetStats()
	assert.Equal(t, 2*sliceAllocations, stats.TotalAllocs)
	// We have freed one batch
	assert.Equal(t, sliceAllocations, stats.TotalLive)
	// One batch is live
	assert.Equal(t, sliceAllocations, stats.TotalFrees)
	// All the freed slots have been reused
	assert.Equal(t, sliceAllocations, stats.TotalReused)
	// The number of chunks hasn't changed, since the first set of
	// allocations
	assert.Equal(t, chunks, stats.TotalChunks)

	// Allocate the same number of slices again
	for i := range slices {
		s.Alloc(uint32(i))
	}

	// We have allocated 3 batches of slices
	stats = s.GetStats()
	assert.Equal(t, 3*sliceAllocations, stats.TotalAllocs)
	// Two batches are live
	assert.Equal(t, 2*sliceAllocations, stats.TotalLive)
	// We have freed one batch
	assert.Equal(t, sliceAllocations, stats.TotalFrees)
	// All the freed slots (one batch) have been reused
	assert.Equal(t, sliceAllocations, stats.TotalReused)
	// Some number of chunks will be allocated for the new allocations
	assert.Greater(t, stats.TotalChunks, chunks)
}

func intToBytes(value int, bytes []byte) []byte {
	return bytes
}

func bytesToInt(bytes []byte) int {
	return int(binary.LittleEndian.Uint64(bytes))
}
