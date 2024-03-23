package objectstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNil(t *testing.T) {
	{
		// Zero value of reference
		r := Reference[int]{}
		assert.True(t, r.IsNil())
	}
	{
		// new Reference with nil value
		r := newReference[int]((*object[int])(nil))
		assert.True(t, r.IsNil())
	}
}

// Demonstrate that a pointer with any non-0 field is not nil
func TestIsNotNil(t *testing.T) {
	slab := mmapSlab[int]()
	for i := range *slab {
		object := &slab[i]
		r := newReference[int](object)
		assert.False(t, r.IsNil())
		assert.Equal(t, r.GetValue(), &object.value)
	}
}
