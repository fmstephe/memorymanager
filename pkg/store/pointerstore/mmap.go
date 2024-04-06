package pointerstore

import (
	"fmt"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
	"golang.org/x/sys/unix"
)

func MmapSlab(objectSize, objectsPerSlab int64) []uintptr {
	// Include the metadata overhead for each object and then round up to power of two
	objectSize = fmath.NxtPowerOfTwo(objectSize + int64(metadataSize))
	// Round the objects count up to a power of two
	objectsPerSlab = fmath.NxtPowerOfTwo(objectsPerSlab)

	slabSize := int(objectSize * objectsPerSlab)

	data, err := unix.Mmap(-1, 0, slabSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_ANON|unix.MAP_PRIVATE)
	if err != nil {
		panic(fmt.Errorf("cannot allocate %d bytes via mmap for %d many objects sized %d %s", slabSize, objectsPerSlab, objectSize, err))
	}

	slotPointers := make([]uintptr, objectsPerSlab)
	for i := range slotPointers {
		slotPointers[i] = (uintptr)((unsafe.Pointer)(&data[int64(i)*objectSize]))
	}

	return slotPointers
}
