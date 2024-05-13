package pointerstore

import (
	"fmt"
	"reflect"
	"unsafe"

	"golang.org/x/sys/unix"
)

func MmapSlab(conf AllocConfig) (objects, metadata []uintptr) {
	data, err := unix.Mmap(-1, 0, int(conf.TotalSlabSize), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_ANON|unix.MAP_PRIVATE)
	if err != nil {
		panic(fmt.Errorf("cannot allocate %#v via mmap because %s", conf, err))
	}

	// Collect pointers to each object allocation slot
	objects = make([]uintptr, conf.ObjectsPerSlab)
	for i := range objects {
		idx := uint64(i) * conf.ObjectSize
		objects[i] = (uintptr)((unsafe.Pointer)(&data[idx]))
	}

	// Collect pointers to each metadata slot
	metadata = make([]uintptr, conf.ObjectsPerSlab)
	for i := range metadata {
		idx := conf.TotalObjectSize + (uint64(i) * conf.MetadataSize)
		metadata[i] = (uintptr)((unsafe.Pointer)(&data[idx]))
	}

	return objects, metadata
}

func MunmapSlab(ptr uintptr, allocConf AllocConfig) error {
	b := pointerToBytes(ptr, int(allocConf.TotalSlabSize))
	return unix.Munmap(b)
}

func pointerToBytes(ptr uintptr, size int) (b []byte) {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sliceHeader.Data = ptr
	sliceHeader.Len = size
	sliceHeader.Cap = size
	return b
}
