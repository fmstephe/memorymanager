package bytestore

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A test which shows that each byteSlab sets itself up correctly.  The
// behaviour we expect is that a slab will _try_ to size its byte chunks to
// maxChunkSize, In the case where each allocation is larger than maxChunkSize
// then each chunk is just the allocation size.
// We also assert that a new slab has empty meta and bytes slices
func TestSlab_NewByteSlab(t *testing.T) {
	{
		// Zero-allocator has 0 sized slots, and 0 sized chunk-size
		slab := newByteSlab(0)
		require.Equal(t, uint32(0), slab.slotSize)
		require.Len(t, slab.meta, 0)
		require.Len(t, slab.bytes, 0)
		require.Equal(t, uint32(0), slab.chunkSize)
		require.Equal(t, uint32(1024), slab.slotCount)
	}

	for i := 0; i < 32; i++ {
		size := uint32(1) << i
		slab := newByteSlab(size)

		// Slot size is always the size passed in
		require.Equal(t, size, slab.slotSize)
		// meta and bytes are always empty after creation
		require.Len(t, slab.meta, 0)
		require.Len(t, slab.bytes, 0)

		if size < maxChunkSize {
			// If the size is less than the maxChunkSize, then the
			// slab's chunk size is maxChunkSize
			require.Equal(t, uint32(maxChunkSize), slab.chunkSize)
			require.Equal(t, uint32(maxChunkSize/size), slab.slotCount)
		} else {
			// If the size is greater than maxChunkSize, then the
			// slab's chunk size is the same as slot size
			require.Equal(t, size, slab.chunkSize)
			require.Equal(t, uint32(1), slab.slotCount)
		}
	}

	{
		// Max-allocator has non power of 2 sized slots, needs a special treament
		slab := newByteSlab(math.MaxUint32)
		require.Equal(t, uint32(math.MaxUint32), slab.slotSize)
		require.Len(t, slab.meta, 0)
		require.Len(t, slab.bytes, 0)
		require.Equal(t, uint32(math.MaxUint32), slab.chunkSize)
		require.Equal(t, uint32(1), slab.slotCount)
	}
}

// Demonstrate that the first allocation of a fresh slab creates the meta and
// bytes slices. We expect that the meta slice will have the same number of
// elements as slotCount and the bytes slice will have the same number of
// elements as the chunkSize.
func TestSlab_FirstAlloc(t *testing.T) {
	{
		// Zero-allocator has 0 sized slots, and 0 sized chunk-size
		slab := newByteSlab(0)
		assertFirstAlloc(t, &slab)
	}

	for i := 0; i <= 32; i++ {
		size := uint32(1) << i
		if i == 32 {
			size = math.MaxUint32
		}

		slab := newByteSlab(size)
		assertFirstAlloc(t, &slab)

	}
}

func assertFirstAlloc(t *testing.T, slab *byteSlab) {
	p, err := slab.alloc(slab.slotSize)
	require.NoError(t, err)
	bytes := slab.get(p)
	require.Len(t, bytes, int(slab.slotSize))
	require.Len(t, slab.meta, 1)
	require.Len(t, slab.bytes, 1)
	require.Len(t, slab.meta[0], int(slab.slotCount))
	require.Len(t, slab.bytes[0], int(slab.chunkSize))
}

func TestSlab_FullChunkAllocFreeRealloc(t *testing.T) {
	// Zero-allocator has 0 sized slots, and 0 sized chunk-size
	//slab := newByteSlab(0)
	//assertAllocateChunkFreeRealloc(t, &slab)

	minSlabSize := 0
	maxSlabSize := 32 // This is a power of two slab size
	// This test is quite slow, so we speed it up, by avoiding running some expensive parts of the test
	if testing.Short() {
		minSlabSize = 8  // Avoid testing small allocations - they are slow
		maxSlabSize = 24 // Avoid testing large allocations - they require a lot of memory
	}

	for i := minSlabSize; i <= maxSlabSize; i++ {
		size := uint32(1) << i
		if i == 32 {
			size = math.MaxUint32
		}

		slab := newByteSlab(size)
		assertAllocateChunkFreeRealloc(t, &slab)
	}
}

func assertAllocateChunkFreeRealloc(t *testing.T, slab *byteSlab) {
	pointers := doAllocations(t, slab, int(slab.slotCount+1))

	// Assert that we have fully allocated an entire chunk and started a second one
	assertBytesAndMeta(t, 2, slab)

	freeAllocations(t, slab, pointers)

	// Assert that the backing data in the slab hasn't changed after all
	// allocations have been freed
	assertBytesAndMeta(t, 2, slab)

	doAllocations(t, slab, int(slab.slotCount+1))

	// Assert that no new backing data is created when we re-allocate all
	// the objects again
	assertBytesAndMeta(t, 2, slab)

	doAllocations(t, slab, int(slab.slotCount+1))

	// Allocate a new batch and see that new backing data is created
	expectedCount := int(((slab.slotCount + 1) * 2) / slab.slotCount)
	if int(((slab.slotCount+1)*2)%slab.slotCount) != 0 {
		expectedCount++
	}
	assertBytesAndMeta(t, expectedCount, slab)
}

func doAllocations(t *testing.T, slab *byteSlab, num int) []Pointer {
	pointers := []Pointer{}
	for i := 0; i < num; i++ {
		size := getRandomSize(slab.slotSize)

		p, err := slab.alloc(size)
		require.NoError(t, err)

		// Use that pointer to get the slice we just allocated
		bytes := slab.get(p)
		// Ensure it's the size we asked for
		require.Len(t, bytes, int(size))

		pointers = append(pointers, p)
	}

	return pointers
}

func assertBytesAndMeta(t *testing.T, count int, slab *byteSlab) {
	assert.Len(t, slab.meta, count, "expected %d, found %d instead", count, len(slab.meta))
	for i := range slab.meta {
		require.Len(t, slab.meta[i], int(slab.slotCount))
	}

	require.Len(t, slab.bytes, count)
	for i := range slab.bytes {
		require.Len(t, slab.bytes[i], int(slab.chunkSize))
	}
}

// Get a random allocation size which is valid for this slotSize
func getRandomSize(slotSize uint32) uint32 {
	if slotSize == 0 {
		return 0
	}

	maxSize := slotSize
	minSize := maxSize / 2
	diff := int64(maxSize - minSize)
	val := uint32(rand.Int63n(diff))
	return val + minSize
}

func freeAllocations(t *testing.T, slab *byteSlab, pointers []Pointer) {
	for i := range pointers {
		slab.free(pointers[i])
	}
}
