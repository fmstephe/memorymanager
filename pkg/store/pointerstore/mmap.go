package pointerstore

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

func MmapSlab(allocConf AllocationConfig) []uintptr {
	slabSize := allocConf.ActualSlabSize
	objectSize := allocConf.ActualObjectSize
	objectsPerSlab := allocConf.ActualObjectsPerSlab

	data, err := unix.Mmap(-1, 0, int(slabSize), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_ANON|unix.MAP_PRIVATE)
	if err != nil {
		panic(fmt.Errorf("cannot allocate %d bytes via mmap for %d many objects sized %d %s", slabSize, objectsPerSlab, objectSize, err))
	}

	slotPointers := make([]uintptr, objectsPerSlab)
	for i := range slotPointers {
		slotPointers[i] = (uintptr)((unsafe.Pointer)(&data[uint64(i)*objectSize]))
	}

	return slotPointers
}
