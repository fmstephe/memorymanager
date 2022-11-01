package store

import (
	"math"
	"testing"

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

func TestSlab_FullChunkAlloc(t *testing.T) {
	minSlabSize := 0
	maxSlabSize := 32 // This is a power of two slab size

	// This test is quite slow, so we speed it up, by avoiding running some expensive parts of the test
	if testing.Short() {
		minSlabSize = 4  // Avoid testing small allocations - they are slow
		maxSlabSize = 20 // Avoid testing large allocations - they require a lot of memory
	}

	{
		// Zero-allocator has 0 sized slots, and 0 sized chunk-size
		slab := newByteSlab(0)
		assertFirstChunkAlloc(t, &slab)
	}

	for i := minSlabSize; i <= maxSlabSize; i++ {
		size := uint32(1) << i
		if i == 32 {
			size = math.MaxUint32
		}

		slab := newByteSlab(size)
		assertFirstChunkAlloc(t, &slab)
	}
}

func assertFirstChunkAlloc(t *testing.T, slab *byteSlab) {
	// Alloc more than one chunks worth of slices
	maxSize := slab.slotSize
	minSize := maxSize / 2

	size := minSize
	for i := 0; i < int(slab.slotCount+1); i++ {
		p, err := slab.alloc(size)
		require.NoError(t, err)

		// Use that pointer to get the slice we just allocated
		bytes := slab.get(p)
		// Ensure it's the size we asked for
		require.Len(t, bytes, int(size))

		// Change the allocation size, demonstrates that we can
		// allocate different sized slices with the same slab
		size++
		if size > maxSize {
			size = minSize
		}
	}

	require.Len(t, slab.meta, 2)
	require.Len(t, slab.bytes, 2)
	require.Len(t, slab.meta[0], int(slab.slotCount))
	require.Len(t, slab.bytes[0], int(slab.chunkSize))
	require.Len(t, slab.meta[1], int(slab.slotCount))
	require.Len(t, slab.bytes[1], int(slab.chunkSize))
}
