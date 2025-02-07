// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// Zero value of Reference returns true for IsNil()
func TestIsNil(t *testing.T) {
	r := RefPointer{}
	assert.True(t, r.IsNil())
}

// Calling newReference() with any nil uintptr parameter will panic
func TestNewReferenceWithNilPanics(t *testing.T) {
	nilPtr := uintptr(unsafe.Pointer(nil))

	allocConfig := NewAllocConfigBySize(8, 8)
	objects, metadata := MmapSlab(allocConfig)

	assert.Panics(t, func() { NewReference(objects[0], nilPtr) })
	assert.Panics(t, func() { NewReference(nilPtr, metadata[0]) })
	assert.Panics(t, func() { NewReference(nilPtr, nilPtr) })
}

// Calling newReference() with any uintptr parameter with generation tag bits set will panic
func TestNewReferenceWithGenerationBitsSetPanics(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 8)
	objects, metadata := MmapSlab(allocConfig)

	badObject := objects[0] | (1 << maskShift)
	badMeta := metadata[0] | (1 << maskShift)

	assert.Panics(t, func() { NewReference(objects[0], badMeta) })
	assert.Panics(t, func() { NewReference(badObject, metadata[0]) })
	assert.Panics(t, func() { NewReference(badObject, badMeta) })
}

// Demonstrate that a pointer with any non-0 field is not nil
func TestNewReference(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadata := MmapSlab(allocConfig)
	for i := range objects {
		r := NewReference(objects[i], metadata[i])
		// The object is not nil
		assert.False(t, r.IsNil())
		// Data pointer points to the correct location
		assert.Equal(t, objects[i], r.DataPtr())
		// Bytes points to data at the correct location
		assert.Equal(t, objects[i], uintptr(unsafe.Pointer(&r.Bytes(8)[0])))
		// Metadata pointer points to the correct location
		assert.Equal(t, metadata[i], r.address.pointer())
		// Generation of a new Reference is always 0
		assert.Equal(t, uint8(0), r.Gen())
		// The data should be accessible through this reference
		assert.NotPanics(t, func() { r.DataPtr() })
		assert.NotPanics(t, func() { r.Bytes(8) })
	}
}

// Demonstrate that the generation tag does not appear in either of the data or
// metadata pointers.
//
// It might be expected that if either of these pointer values did have the
// generation tag in them they wouldn't work, and all our conventional tests
// would reveal this. But our lived experience indicates that some platforms
// (anecdotally OSX M2 CPU running Ubunutu in a VM) work perfectly fine with
// garbage data in the high bits of the address. While other platforms
// (anecdotally Ubuntu running on whatever github uses) produce segfaults and
// seem to weirdly claim the address was 0x0.
//
// So now we have this test, which should catch us if we don't strip the
// generation tag out of wherever we've hidden it (at time of writing it's
// hidden in the top 8 bits of the object address pointer).
func TestGenerationDoesNotAppearInOtherFields(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadatas := MmapSlab(allocConfig)

	r := NewReference(objects[0], metadatas[0])
	dataPtr := r.DataPtr()
	metadata := r.metadata()

	metadata.setGen(maxGen)
	r = r.withGen(maxGen)

	assert.Equal(t, dataPtr, r.DataPtr())
	assert.Equal(t, metadata, r.metadata())
	assert.Equal(t, uint8(maxGen), r.Gen())
}

func TestFree(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadatas := MmapSlab(allocConfig)

	r := NewReference(objects[0], metadatas[0])
	meta := r.metadata()

	assert.False(t, meta.isFree())
	assert.Equal(t, uint8(0), meta.dataAddressAndGen.gen())
	assert.Equal(t, uint8(0), r.Gen())

	r.free()

	// The reference still points to the same metadata location
	assert.Equal(t, meta, r.metadata())

	// The metadata is now marked as free
	assert.True(t, meta.isFree())
	// After free is called the metadata for this reference has a new
	// generation tag, while the reference has the same old generation tag
	assert.Equal(t, uint8(1), meta.dataAddressAndGen.gen())
	assert.Equal(t, uint8(0), r.Gen())

	// Accessng the data of a freed reference will panic
	assert.Panics(t, func() { r.DataPtr() })
	assert.Panics(t, func() { r.Bytes(8) })

	// Reallocing a freed reference will panic
	assert.Panics(t, func() { r.Realloc() })

	// Freeing a freed reference will panic
	assert.Panics(t, func() { r.free() })
}

func TestAllocFromFree(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadatas := MmapSlab(allocConfig)

	r := NewReference(objects[0], metadatas[0])
	meta := r.metadata()

	assert.False(t, meta.isFree())
	assert.Equal(t, uint8(0), meta.dataAddressAndGen.gen())
	assert.Equal(t, uint8(0), r.Gen())

	r.free()
	r = r.allocFromFree()

	// The reference still points to the same metadata location
	assert.Equal(t, meta, r.metadata())

	// The metadata is now marked as not free
	assert.False(t, meta.isFree())
	// After allocFromFree is called the reference's generation will match the metadata's generation
	assert.Equal(t, uint8(1), meta.dataAddressAndGen.gen())
	assert.Equal(t, uint8(1), r.Gen())

	// Accessng the data of an allocated-from-free reference will not panic
	assert.NotPanics(t, func() { r.DataPtr() })
	assert.NotPanics(t, func() { r.Bytes(8) })

	// Freeing an allocated-from-free reference will not panic
	assert.NotPanics(t, func() { r.free() })
}

func TestAllocFromFree_NoFree(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadatas := MmapSlab(allocConfig)

	r := NewReference(objects[0], metadatas[0])

	assert.Panics(t, func() { r.allocFromFree() })
}

func TestRealloc(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadatas := MmapSlab(allocConfig)

	r1 := NewReference(objects[0], metadatas[0])
	meta1 := r1.metadata()

	dataPtr := r1.DataPtr()
	metadata := r1.metadata()
	gen := r1.Gen()

	r2 := r1.Realloc()
	meta2 := r2.metadata()

	// The metadata of r1 and r2 is the same location
	assert.Equal(t, meta1, meta2)
	// The generation matches in both the r2 and the metadata
	assert.Equal(t, r2.Gen(), meta2.gen())
	// The generation does not match between r1 and the metadata
	assert.NotEqual(t, r1.Gen(), meta2.gen())

	// Assert that the data/metadata pointed to by r1 and r2 is the same
	assert.Equal(t, dataPtr, r2.DataPtr())
	assert.Equal(t, metadata, r2.metadata())

	// Assert that r2 has a different generation than r1
	assert.NotEqual(t, gen, r2.Gen())

	// Assert that data is no longer accessible via r1
	assert.Panics(t, func() { r1.DataPtr() })
	assert.Panics(t, func() { r1.Bytes(8) })

	// Assert that data is accessible via r2
	assert.NotPanics(t, func() { r2.DataPtr() })
	assert.NotPanics(t, func() { r2.Bytes(8) })
}
