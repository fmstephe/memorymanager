package objectstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Demonstrate that a pointer with all 0 fields is nil
func TestIsNil(t *testing.T) {
	p := Pointer[string]{
		chunk:  0,
		offset: 0,
	}
	assert.True(t, p.IsNil())
}

// Demonstrate that a pointer with any non-0 field is not nil
func TestIsNotNil(t *testing.T) {
	p := Pointer[string]{
		chunk:  0,
		offset: 1,
	}
	assert.False(t, p.IsNil())

	p = Pointer[string]{
		chunk:  1,
		offset: 0,
	}
	assert.False(t, p.IsNil())

	p = Pointer[string]{
		chunk:  1,
		offset: 1,
	}
	assert.False(t, p.IsNil())
}
