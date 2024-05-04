package pointerstore

import (
	"unsafe"

	"github.com/fmstephe/flib/fmath"
)

type AllocationConfig struct {
	RequestedObjectSize uint64
	RequestedSlabSize   uint64
	//
	ObjectsPerSlab    uint64
	ObjectSize        uint64
	TotalObjectSize   uint64
	MetadataSize      uint64
	TotalMetadataSize uint64
	TotalSlabSize     uint64
}

func NewAllocationConfigBySize(requestedObjectSize uint64, requestedSlabSize uint64) AllocationConfig {
	objectSize := uint64(fmath.NxtPowerOfTwo(int64(requestedObjectSize)))

	totalObjectSize := uint64(fmath.NxtPowerOfTwo(int64(requestedSlabSize)))

	if totalObjectSize < objectSize {
		// If the slab is too small - we match the object size for one
		// allocation per slab
		totalObjectSize = objectSize
	}

	objectsPerSlab := totalObjectSize / objectSize

	// TODO have a think about this - we don't strictly _need_ the metadata
	// to be aligned by a power of 2 (do we?)
	metadataSize := uint64(fmath.NxtPowerOfTwo(int64(unsafe.Sizeof(metadata{}))))

	totalMetadataSize := metadataSize * objectsPerSlab

	totalSlabSize := totalObjectSize + totalMetadataSize

	return AllocationConfig{
		RequestedObjectSize: requestedObjectSize,
		RequestedSlabSize:   requestedSlabSize,

		ObjectsPerSlab:    objectsPerSlab,
		ObjectSize:        objectSize,
		TotalObjectSize:   totalObjectSize,
		MetadataSize:      metadataSize,
		TotalMetadataSize: totalMetadataSize,
		TotalSlabSize:     totalSlabSize,
	}
}
