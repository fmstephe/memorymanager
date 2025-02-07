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
type RefPointer struct {
	address taggedAddress
}

func NewReference(dataAddress, metaAddress uintptr) RefPointer {
	if dataAddress == nilPtr {
		panic("cannot create new Reference with nil data pointer")
	}

	if metaAddress == nilPtr {
		panic("cannot create new Reference with nil metadata pointer")
	}

	r := RefPointer{
		address: newTaggedAddress(metaAddress),
	}

	// Set the dataAddressAndGen in this reference's metadata.
	meta := r.metadata()
	meta.dataAddressAndGen = newTaggedAddress(dataAddress)
	// The value defaults to false, but we write it here for readability
	meta.setNotFree()

	return r
}

func (r RefPointer) free() {
	// Check that this reference can access the allocation to free it
	r.accessibleActiveAddress()

	meta := r.metadata()

	// Mark the object as free
	meta.setFree()

	// Increment the generation for the allocation's metadata
	meta.incGen()
}

// Sets the RefPointer's metadata to not-free and creates a new RefPointer with
// the correct generation tag.
func (r RefPointer) allocFromFree() RefPointer {
	meta := r.metadata()

	if !meta.isFree() {
		panic(fmt.Errorf("attempt to alloc-from-free active allocation %v", r))
	}

	// This object is now allocated and is no long free
	meta.setNotFree()

	// Get the metadata generation tag
	gen := meta.gen()

	// Create new RefPointer with correct generation tag
	return r.withGen(gen)
}

// This method re-allocates the memory location. When this method returns r
// will no longer be a valid reference.  The reference returned _will_ be a
// valid reference to the same location.
func (r RefPointer) Realloc() RefPointer {
	// Test that this reference is actually allowed to access the allocation
	r.accessibleActiveAddress()

	meta := r.metadata()

	// Get and increment the metadata generation tag
	gen := meta.incGen()

	// Set the new generation tag in the reference
	return r.withGen(gen)
}

func (r RefPointer) DataPtr() uintptr {
	return r.accessibleActiveAddress().pointer()
}

// Convenient method to retrieve raw data of an allocation
func (r RefPointer) Bytes(size int) []byte {
	return r.accessibleActiveAddress().bytes(size)
}

func (r RefPointer) IsNil() bool {
	return r.address.isNil()
}

// Calling this method
//
// 1: looks up the metadata for this reference
// 2: Verifies that the reference is _allowed_ to access this data
// 3: Returns the taggedAddress pointing to the actual data
func (r RefPointer) accessibleActiveAddress() taggedAddress {
	meta := r.metadata()

	if meta.isFree() {
		panic(fmt.Errorf("attempt to access freed allocation %v", r))
	}

	if meta.gen() != r.Gen() {
		panic(fmt.Errorf("generation mismatch between metadata (%d) and reference (%d)", meta.gen(), r.Gen()))
	}

	return meta.dataAddressAndGen
}

func (r RefPointer) metadata() *metadata {
	return (*metadata)(unsafe.Pointer(r.address.pointer()))
}

func (r RefPointer) Gen() uint8 {
	return r.address.gen()
}

func (r RefPointer) withGen(gen uint8) RefPointer {
	return RefPointer{
		address: r.address.withGen(gen),
	}
}
