// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"fmt"
	"unsafe"
)

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
	address taggedAddress
}

func NewReference(dataAddress, metaAddress uintptr) RefPointer {
	if dataAddress == (uintptr)(unsafe.Pointer(nil)) {
		panic("cannot create new Reference with nil data pointer")
	}

	if metaAddress == (uintptr)(unsafe.Pointer(nil)) {
		panic("cannot create new Reference with nil metadata pointer")
	}

	r := RefPointer{
		address: newTaggedAddress(metaAddress),
	}

	// Set the dataAddressAndGen in this reference's metadata.
	//
	// In one of the next steps we will use this to look up the actual data
	// _and_ verify the generation of the reference.
	meta := r.metadata()
	meta.dataAddressAndGen = newTaggedAddress(dataAddress)
	// The value defaults to false, but we write it here for readability
	meta.setNotFree()

	return r
}

//gcassert:noescape
func (r *RefPointer) free() {
	meta := r.metadata()

	if meta.isFree() {
		// NB: The odd-looking *r here actually prevents an allocation.
		// Fuller explanation found in DataPtr()
		panic(fmt.Errorf("attempted to Free freed allocation %v", *r))
	}

	meta.checkReference(r)

	// Mark the object as free
	meta.setFree()

	// Increment the generation for the allocation and set that generation in
	// the Metadata and Reference
	gen := meta.gen()
	gen++
	meta.setGen(gen)
}

// NB: In the future this method should return a new RefPointer with updated
// generation tag instead of modifying the method receiver.
//
//gcassert:noescape
func (r *RefPointer) allocFromFree() {
	meta := r.metadata()

	// This object is now allocated and is no long free
	meta.setNotFree()
	// Update this reference so it's generatation tag matches the metadata
	r.setGen(meta.gen())
}

// This method re-allocates the memory location. When this method returns r
// will no longer be a valid reference.  The reference returned _will_ be a
// valid reference to the same location.
//
//gcassert:noescape
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

//gcassert:noescape
func (r *RefPointer) IsNil() bool {
	return r.address.isNil()
}

//gcassert:noescape
func (r *RefPointer) DataPtr() uintptr {
	return r.dataAddress().address()
}

// Convenient method to retrieve raw data of an allocation
//
//gcassert:noescape
func (r *RefPointer) Bytes(size int) []byte {
	return r.dataAddress().bytes(size)
}

// Calling this method
//
// 1: looks up the metadata for this reference
// 2: Verifies that the reference is _allowed_ to access this data
// 3: Returns the taggedAddress pointing to the actual data
//
//gcassert:noescape
func (r *RefPointer) dataAddress() taggedAddress {
	meta := r.metadata()

	if meta.isFree() {
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

	return meta.dataAddressAndGen
}

//gcassert:noescape
func (r *RefPointer) metaPtr() uintptr {
	return r.metaAddress().address()
}

//gcassert:noescape
func (r *RefPointer) metadata() *metadata {
	return (*metadata)(unsafe.Pointer(r.metaAddress().address()))
}

//gcassert:noescape
func (r *RefPointer) metaAddress() taggedAddress {
	return r.address
}

//gcassert:noescape
func (r *RefPointer) Gen() uint8 {
	return r.address.gen()
}

//gcassert:noescape
func (r *RefPointer) setGen(gen uint8) {
	r.address = r.address.withGen(gen)
}
