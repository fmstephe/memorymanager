package pointerstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Zero value of Reference returns true for IsNil()
func TestIsNil(t *testing.T) {
	r := Reference{}
	assert.True(t, r.IsNil())
}

// Calling newReference() with nil will panic
func TestNewReferenceWithNilPanics(t *testing.T) {
	assert.Panics(t, func() { NewReference(0, 0) })
}

// Demonstrate that a pointer with any non-0 field is not nil
func TestIsNotNil(t *testing.T) {
	allocConfig := NewAllocationConfigBySize(8, 32*8)
	objects, metadata := MmapSlab(allocConfig)
	for i := range objects {
		r := NewReference(objects[i], metadata[i])
		// The object is not nil
		assert.False(t, r.IsNil())
		// Data pointer points to the correct location
		assert.Equal(t, objects[i], r.GetDataPtr())
		// Metadata pointer points to the correct location
		assert.Equal(t, metadata[i], r.getMetadataPtr())
		// Generation of a new Reference is always 0
		assert.Equal(t, uint8(0), r.GetGen())
	}
}
