package objectstore

import (
	"fmt"
	"unsafe"
)

const maskShift = 56 // This leaves 8 bits for the meta-data
const metaMask = uint64(0xFF << maskShift)
const pointerMask = uint64(^metaMask)

type Reference[O any] struct {
	address uint64
}

func newReference[O any](obj *object[O]) Reference[O] {
	if obj == nil {
		panic("cannot create new Reference with nil object")
	}

	address := (uint64)((uintptr)(unsafe.Pointer(obj)))
	maskedAddress := address & pointerMask

	if address != maskedAddress {
		panic(fmt.Errorf("The raw pointer (%d) uses more than %d bits", address, maskShift))
	}

	// NB: The meta on a brand new Reference is always 0
	// So we don't set it
	return Reference[O]{
		address: maskedAddress,
	}
}

func (r *Reference[O]) GetValue() *O {
	obj := r.getObject()
	if !obj.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to get freed object %v", r))
	}

	refMeta := r.getMetaByte()
	objMeta := obj.meta
	if refMeta != objMeta {
		panic(fmt.Errorf("Attempt to get value (%d) using stale reference (%d)", objMeta, refMeta))
	}

	return &obj.value
}

func (r *Reference[O]) getObject() *object[O] {
	obj := (*object[O])(unsafe.Pointer(r.getUintptr()))
	return obj
}

func (r *Reference[O]) IsNil() bool {
	return r.getUintptr() == 0
}

func (r *Reference[O]) getUintptr() uintptr {
	return (uintptr)(pointerMask & r.address)
}

func (r *Reference[O]) getMetaByte() byte {
	return (byte)((r.address & metaMask) >> maskShift)
}

func (r *Reference[O]) setMeta(meta byte) {
	r.address = (r.address & pointerMask) | (uint64(meta) << maskShift)
}
