package pointerstore

import (
	"fmt"
	"unsafe"
)

const maskShift = 56 // This leaves 8 bits for the generation data
const genMask = uint64(0xFF << maskShift)
const pointerMask = ^genMask

// The address field holds a pointer to an object, but also sneaks a
// generation value in the top 8 bits of the address field.
//
// The generation must be masked out to get a usable pointer value. The object
// pointed to must have the same generation value in order to access/free that
// object.
type Reference struct {
	address  uint64
	metadata uint64
}

const metadataSize = unsafe.Sizeof(metadata{})

// If the object's metadata has a non-nil nextFree pointer then the object is
// currently free. Object's which have never been allocated are implicitly
// free, but have a nil nextFree.
//
// An object's metadata has a gen field. Only references with the same gen
// value can access/free objects they point to. This is a best-effort safety
// check to try to catch use-after-free type errors.
type metadata struct {
	nextFree Reference
	gen      uint8
}

func NewReference(pAddress, pMetadata uintptr) Reference {
	if pAddress == (uintptr)(unsafe.Pointer(nil)) {
		panic("cannot create new Reference with nil pointer")
	}

	address := uint64(pAddress)
	// This sets the generation to 0 by clearing the smuggled bits
	maskedAddress := address & pointerMask

	// Setting the generation 0 shouldn't actually change the address
	// If there were any 1s in the top part of the address our generation
	// smuggling system will break this pointer. This is an unrecoverable error.
	if address != maskedAddress {
		panic(fmt.Errorf("The raw pointer (%d) uses more than %d bits", address, maskShift))
	}

	// NB: The gen on a brand new Reference is always 0
	// So we don't set it
	return Reference{
		address:  maskedAddress,
		metadata: uint64(pMetadata),
	}
}

func (r *Reference) AllocFromFree() (nextFree Reference) {
	// Grab the object for the slot and nil out the slot's nextFree pointer
	obj := r.getMetadata()
	nextFree = obj.nextFree
	obj.nextFree = Reference{}

	// If the nextFree pointer points back to this Reference, then there
	// are no more freed slots available
	if nextFree == *r {
		nextFree = Reference{}
	}

	// Increment the generation for the object and set that generation in
	// the Reference
	obj.gen++
	r.setGen(obj.gen)

	return nextFree
}

func (r *Reference) Free(oldFree Reference) {
	meta := r.getMetadata()

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed allocation %v", r))
	}

	if meta.gen != r.GetGen() {
		panic(fmt.Errorf("Attempt to free allocation (%d) using stale reference (%d)", meta.gen, r.GetGen()))
	}

	if oldFree.IsNil() {
		meta.nextFree = *r
	} else {
		meta.nextFree = oldFree
	}
}

func (r *Reference) IsNil() bool {
	return r.getMetadataPtr() == 0
}

func (r *Reference) GetDataPtr() uintptr {
	meta := r.getMetadata()

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to get freed allocation %v", r))
	}

	if meta.gen != r.GetGen() {
		panic(fmt.Errorf("Attempt to get value (%d) using stale reference (%d)", meta.gen, r.GetGen()))
	}
	return (uintptr)(r.address)
}

func (r *Reference) getMetadataPtr() uintptr {
	return (uintptr)(r.metadata)
}

func (r *Reference) getMetadata() *metadata {
	return (*metadata)(unsafe.Pointer(r.getMetadataPtr()))
}

func (r *Reference) GetGen() uint8 {
	return (uint8)((r.address & genMask) >> maskShift)
}

func (r *Reference) setGen(gen uint8) {
	r.address = (r.address & pointerMask) | (uint64(gen) << maskShift)
}
