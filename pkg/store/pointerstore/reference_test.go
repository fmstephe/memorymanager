package pointerstore

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// Zero value of Reference returns true for IsNil()
func TestIsNil(t *testing.T) {
	r := Reference{}
	assert.True(t, r.IsNil())
}

// Calling newReference() with nil will panic
func TestNewReferenceWithNilPanics(t *testing.T) {
	assert.Panics(t, func() { NewReference(0) })
}

// Demonstrate that a pointer with any non-0 field is not nil
func TestIsNotNil(t *testing.T) {
	slab := MmapSlab(int64(unsafe.Sizeof(uint32(1))), 32)
	for _, ptr := range slab {
		r := NewReference(ptr)
		// The object is not nil
		assert.False(t, r.IsNil())
		// The metadata pointer is at the start of the object (i.e. where the original pointer pointed)
		assert.Equal(t, ptr, r.getMetadataPtr())
		// The data pointer is after the metadata object
		assert.Equal(t, ptr+metadataSize, r.GetDataPtr())
		// Generation of a new Reference is always 0
		assert.Equal(t, uint8(0), r.GetGen())
	}
}
