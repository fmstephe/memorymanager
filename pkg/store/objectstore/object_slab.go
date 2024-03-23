package objectstore

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// TODO this is sized by objectChunkSize, but really it should be size by some fixed number of pages
// We then determine the number of objects that can fit into each allocation dynamically.
// This can easily be done in the future
func mmapSlab[O any]() *[objectChunkSize]object[O] {
	o := new(object[O])
	objectSize := uint64(unsafe.Sizeof(*o))

	data, err := unix.Mmap(-1, 0, int(objectSize*objectChunkSize), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_ANON|unix.MAP_PRIVATE)
	if err != nil {
		panic(fmt.Errorf("cannot allocate %d bytes via mmap for %T: %s", objectSize*objectChunkSize, o.value, err))
	}

	return (*[objectChunkSize]object[O])(unsafe.Pointer(&data[0]))
}