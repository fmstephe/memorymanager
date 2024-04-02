package objectstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Zero value of Reference returns true for IsNil()
func TestIsNil(t *testing.T) {
	r := Reference[int]{}
	assert.True(t, r.IsNil())
}

// Calling newReference() with nil will panic
func TestNewReferenceWithNilPanics(t *testing.T) {
	assert.Panics(t, func() { newReference[int](nil) })
}

// Demonstrate that a pointer with any non-0 field is not nil
func TestIsNotNil(t *testing.T) {
	slab := mmapSlab[int]()
	for i := range *slab {
		object := &slab[i]
		r := newReference[int](object)
		// The object is not nil
		assert.False(t, r.IsNil())
		// The value pointed to, is the value in object
		assert.Equal(t, r.GetValue(), &object.value)
		// The meta of the object is the same as the meta in the reference
		assert.Equal(t, object.meta, r.getMetaByte())
	}
}
