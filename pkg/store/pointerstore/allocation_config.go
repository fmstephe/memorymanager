package pointerstore

import "github.com/fmstephe/flib/fmath"

type AllocationConfig struct {
	RequestedObjectSize     uint64
	RequestedObjectsPerSlab uint64
	RequestedSlabSize       uint64

	ActualObjectSize     uint64
	ActualObjectsPerSlab uint64
	ActualSlabSize       uint64
}

func NewAllocationConfigByObjects(requestedObjectSize uint64, requestedObjectsPerSlab uint64) AllocationConfig {
	actualObjectSize, actualObjectsPerSlab, actualSlabSize := allocConfWithObjectsPerSlab(requestedObjectSize, requestedObjectsPerSlab)
	return AllocationConfig{
		RequestedObjectSize:     requestedObjectSize,
		RequestedObjectsPerSlab: requestedObjectsPerSlab,
		RequestedSlabSize:       0,

		ActualObjectSize:     actualObjectSize,
		ActualObjectsPerSlab: actualObjectsPerSlab,
		ActualSlabSize:       actualSlabSize,
	}
}

func NewAllocationConfigBySize(requestedObjectSize uint64, requestedSlabSize uint64) AllocationConfig {
	actualObjectSize, actualObjectsPerSlab, actualSlabSize := allocConfWithSlabSize(requestedObjectSize, requestedSlabSize)
	return AllocationConfig{
		RequestedObjectSize:     requestedObjectSize,
		RequestedObjectsPerSlab: 0,
		RequestedSlabSize:       requestedSlabSize,

		ActualObjectSize:     actualObjectSize,
		ActualObjectsPerSlab: actualObjectsPerSlab,
		ActualSlabSize:       actualSlabSize,
	}
}

func allocConfWithObjectsPerSlab(requestedObjectSize, requestedObjectsPerSlab uint64) (actualObjectSize, actualObjectsPerSlab, actualSlabSize uint64) {
	// Include the metadata overhead for each object and then round up to power of two
	actualObjectSize = uint64(fmath.NxtPowerOfTwo(int64(requestedObjectSize) + int64(metadataSize)))
	// Round the objects count up to a power of two NB: 0 -> 1 with this function
	actualObjectsPerSlab = uint64(fmath.NxtPowerOfTwo(int64(requestedObjectsPerSlab)))

	actualSlabSize = actualObjectSize * actualObjectsPerSlab

	return actualObjectSize, actualObjectsPerSlab, actualSlabSize
}

func allocConfWithSlabSize(requestedObjectSize, requestedSlabSize uint64) (actualObjectSize, actualObjectsPerSlab, actualSlabSize uint64) {
	// Include the metadata overhead for each object and then round up to power of two
	actualObjectSize = uint64(fmath.NxtPowerOfTwo(int64(requestedObjectSize) + int64(metadataSize)))

	actualSlabSize = uint64(fmath.NxtPowerOfTwo(int64(requestedSlabSize)))

	if actualSlabSize < actualObjectSize {
		// If the slab is too small - we match the object size for one
		// allocation per slab
		actualSlabSize = actualObjectSize
	}

	actualObjectsPerSlab = actualSlabSize / actualObjectSize

	return actualObjectSize, actualObjectsPerSlab, actualSlabSize
}
