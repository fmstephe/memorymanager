package objectstore

import (
	"fmt"
	"unsafe"
)

const maskShift = 56 // This leaves 8 bits for the generation data
const genMask = uint64(0xFF << maskShift)
const pointerMask = ^genMask

// The address field holds a pointer to an object[O], but also sneaks a
// generation value in the top 8 bits of the address field.
//
// The generation must be masked out to get a usable pointer value. The object
// pointed to must have the same generation value in order to access/free that
// object.
type Reference[O any] struct {
	address uint64
}

// If the object has a non-nil nextFree pointer then the object is currently
// free. Object's which have never been allocated are implicitly free, but have
// a nil nextFree.
//
// An object has a gen field. Only references with the same gen value can
// access/free objects they point to. This is a best-effort safety check to try
// to catch use-after-free type errors.
type object[O any] struct {
	nextFree Reference[O]
	gen      uint8
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

	// NB: The gen on a brand new Reference is always 0
	// So we don't set it
	return Reference[O]{
		address: maskedAddress,
	}
}

func (r *Reference[O]) allocFromFree() (nextFree Reference[O]) {
	// Grab the object for the slot and nil out the slot's nextFree pointer
	obj := r.getObject()
	nextFree = obj.nextFree
	obj.nextFree = Reference[O]{}

	// If the nextFree pointer points back to this Reference, then there
	// are no more freed slots available
	if nextFree == *r {
		nextFree = Reference[O]{}
	}

	// Increment the generation for the object and set that generation in
	// the Reference
	obj.gen++
	r.setGen(obj.gen)

	return nextFree
}

func (r *Reference[O]) free(oldFree Reference[O]) {
	obj := r.getObject()

	if !obj.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed object %v", r))
	}

	if obj.gen != r.getGen() {
		panic(fmt.Errorf("Attempt to free object (%d) using stale reference (%d)", obj.gen, r.getGen()))
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

	refGen := r.getGen()
	objGen := obj.gen
	if refGen != objGen {
		panic(fmt.Errorf("Attempt to get value (%d) using stale reference (%d)", objGen, refGen))
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

func (r *Reference[O]) getGen() uint8 {
	return (uint8)((r.address & genMask) >> maskShift)
}

func (r *Reference[O]) setGen(gen uint8) {
	r.address = (r.address & pointerMask) | (uint64(gen) << maskShift)
}
