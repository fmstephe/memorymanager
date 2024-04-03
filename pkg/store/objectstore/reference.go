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

// If the object has a non-nil nextFree pointer then the object is currently
// free. Object's which have never been allocated are implicitly free, but have
// a nil nextFree
//
// An object has a meta field. Only references with the same meta value can
// access/free objects they point to. This is a best-effort safety check to try
// to catch use-after-free type errors.
type object[O any] struct {
	nextFree Reference[O]
	meta     byte
	value    O
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

func (r *Reference[O]) allocFromFree() (nextFree Reference[O]) {
	// Grab the meta-data for the slot and nil out the, now
	// allocated, slot's nextFree pointer
	obj := r.getObject()
	nextFree = obj.nextFree
	obj.nextFree = Reference[O]{}

	// If the nextFree pointer points to this Reference, then
	// there are no more freed slots available
	if nextFree == *r {
		nextFree = Reference[O]{}
	}

	// Increment the generation meta-data for the object
	// and set that meta value in the Reference
	obj.meta++
	r.setMeta(obj.meta)

	return nextFree
}

func (r *Reference[O]) free(oldFree Reference[O]) {
	obj := r.getObject()

	if !obj.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed object %v", r))
	}

	if obj.meta != r.getMeta() {
		panic(fmt.Errorf("Attempt to free object (%d) using stale reference (%d)", obj.meta, r.getMeta()))
	}

	if oldFree.IsNil() {
		obj.nextFree = *r
	} else {
		obj.nextFree = oldFree
	}
}

func (r *Reference[O]) GetValue() *O {
	obj := r.getObject()
	if !obj.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to get freed object %v", r))
	}

	refMeta := r.getMeta()
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

func (r *Reference[O]) getMeta() byte {
	return (byte)((r.address & metaMask) >> maskShift)
}

func (r *Reference[O]) setMeta(meta byte) {
	r.address = (r.address & pointerMask) | (uint64(meta) << maskShift)
}
