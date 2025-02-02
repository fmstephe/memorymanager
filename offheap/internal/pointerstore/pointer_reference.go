// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"fmt"
	"unsafe"
)

const maskShift = 56 // This leaves 8 bits for the generation data
const genMask = uint64(0xFF << maskShift)
const pointerMask = ^genMask

// The address field holds a pointer to an object, but also sneaks a generation
// value in the top 8 bits of the dataAddress field.
//
// The generation must be masked out to get a usable pointer value. The object
// pointed to must have the same generation value in order to access/free that
// object.
//
// Because the Refpointer struct is two words (on 64 bit systems) in size, all
// method receivers are pointers to avoid copying two words in method calls.
// However, the RefPointer is _always_ used as a value. This means that having
// a pointer receiver could potentially cause the RefPointer to be allocated if
// its receiver escapes to the heap (according to escape analysis). We assert
// that all methods do not allow their receiver variable to escape to the heap.
//
// This is all very nice and good, but the idea that avoiding copying two words
// is a meaningful improvement are speculative and haven't been tested. This
// would be a good target for future performance testing.
type RefPointer struct {
	dataAddressAndGen uint64
	metaAddress       uint64
}

// If the object's metadata has a non-nil nextFree pointer then the object is
// currently free. Object's which have never been allocated are implicitly
// free, but have a nil nextFree.
//
// An object's metadata has a gen field. Only references with the same gen
// value can access/free objects they point to. This is a best-effort safety
// check to try to catch use-after-free type errors.
type metadata struct {
	nextFree          RefPointer
	dataAddressAndGen uint64
}

//gcassert:noescape
func (m *metadata) gen() uint8 {
	return (uint8)((m.dataAddressAndGen & genMask) >> maskShift)
}

//gcassert:noescape
func (m *metadata) setGen(gen uint8) {
	m.dataAddressAndGen = (m.dataAddressAndGen & pointerMask) | (uint64(gen) << maskShift)
}

// Check that the metadata for a reference agrees with the generation tag and the data address.
// Failure results in a panic.
//
//gcassert:noescape
func (m *metadata) checkReference(r *RefPointer) {
	if m.gen() != r.Gen() {
		panic(fmt.Errorf("attempt to get value (%d) using stale reference (%d)", m.gen(), r.Gen()))
	}

	if m.dataAddressAndGen != r.dataAddressAndGen {
		panic(fmt.Errorf("attempt to get value where reference's data-address (%d) and metadata's data-address (%d) are different", r.dataAddressAndGen, m.dataAddressAndGen))
	}
}

func NewReference(dataAddress, metaAddress uintptr) RefPointer {
	if dataAddress == (uintptr)(unsafe.Pointer(nil)) {
		panic("cannot create new Reference with nil data pointer")
	}

	r := RefPointer{
		dataAddressAndGen: uint64(dataAddress),
		metaAddress:       uint64(metaAddress),
	}

	if r.Gen() != 0 {
		panic(fmt.Errorf("the data pointer (%d) contains a non-zero generation tag (%d)", dataAddress, r.Gen()))
	}

	// Set the dataAddressAndGen in this reference's metadata.
	//
	// In one of the next steps we will use this to look up the actual data
	// _and_ verify the generation of the reference.
	meta := r.metadata()
	meta.dataAddressAndGen = r.dataAddressAndGen

	return r
}

//gcassert:noescape
func (r *RefPointer) AllocFromFree() (nextFree RefPointer) {
	// Grab the nextFree reference, and nil it for this metadata
	meta := r.metadata()
	nextFree = meta.nextFree
	meta.nextFree = RefPointer{}

	// If the nextFree pointer points back to this Reference, then there
	// are no more freed slots available
	if nextFree == *r {
		nextFree = RefPointer{}
	}

	// Increment the generation for the allocation and set that generation in
	// the Metadata and Reference
	gen := meta.gen()
	gen++
	meta.setGen(gen)
	r.setGen(gen)

	return nextFree
}

//gcassert:noescape
func (r *RefPointer) Free(oldFree RefPointer) {
	meta := r.metadata()

	if !meta.nextFree.IsNil() {
		// NB: The odd-looking *r here actually prevents an allocation.
		// Fuller explanation found in DataPtr()
		panic(fmt.Errorf("attempted to Free freed allocation %v", *r))
	}

	meta.checkReference(r)

	if oldFree.IsNil() {
		meta.nextFree = *r
	} else {
		meta.nextFree = oldFree
	}
}

//gcassert:noescape
func (r *RefPointer) IsNil() bool {
	return r.metadataPtr() == 0
}

//gcassert:noescape
func (r *RefPointer) DataPtr() uintptr {
	meta := r.metadata()

	if !meta.nextFree.IsNil() {
		// NB: We make a copy of r here - otherwise the compiler
		// believes that r itself escapes to the heap (not strictly
		// wrong) and will allocate it to the heap, even if this path
		// is not taken. This panic path _does_ allocate due to the fmt
		// call, but if we don't take a copy of r in the fmt call, then
		// every call will allocate regardless of whether the method
		// panics or not
		panic(fmt.Errorf("attempt to get freed allocation %v", *r))
	}

	meta.checkReference(r)

	return (uintptr)(r.dataAddressAndGen & pointerMask)
}

// Convenient method to retrieve raw data of an allocation
//
//gcassert:noescape
func (r *RefPointer) Bytes(size int) []byte {
	ptr := r.DataPtr()
	return pointerToBytes(ptr, size)
}

//gcassert:noescape
func (r *RefPointer) metadataPtr() uintptr {
	return (uintptr)(r.metaAddress & pointerMask)
}

//gcassert:noescape
func (r *RefPointer) metadata() *metadata {
	return (*metadata)(unsafe.Pointer(r.metadataPtr()))
}

//gcassert:noescape
func (r *RefPointer) Gen() uint8 {
	return (uint8)((r.dataAddressAndGen & genMask) >> maskShift)
}

//gcassert:noescape
func (r *RefPointer) setGen(gen uint8) {
	r.dataAddressAndGen = (r.dataAddressAndGen & pointerMask) | (uint64(gen) << maskShift)
}

// This method re-allocates the memory location. When this method returns r
// will no longer be a valid reference.  The reference returned _will_ be a
// valid reference to the same location.
func (r *RefPointer) Realloc() RefPointer {
	newRef := *r

	// Get metadata generation tag and increment it
	meta := r.metadata()
	gen := meta.gen()
	gen++

	// Set the new generation tag in both the metadata and reference
	meta.setGen(gen)
	newRef.setGen(gen)
	return newRef
}
