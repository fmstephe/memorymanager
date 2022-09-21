package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MutableStruct struct {
	Field int
}

// Demonstrate that we can create an object, modify that object and when we get
// that object from the store we can see the modifications
// We ensure that we allocate so many objects that we will need more than one chunk
// to store all objects.
func Test_Object_NewModifyGet(t *testing.T) {
	os := NewObjectStore[MutableStruct]()

	// Create all the objects and modify field
	pointers := make([]ObjectPointer[MutableStruct], objectChunkSize*3)
	for i := range pointers {
		p, s := os.New()
		s.Field = i
		pointers[i] = p
	}

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		s := os.Get(p)
		assert.Equal(t, i, s.Field)
	}
}

// Demonstrate that we can create an object, then get that object and modify it
// we can then get that object again and will see the modification
// We ensure that we allocate so many objects that we will need more than one chunk
// to store all objects.
func Test_Object_GetModifyGet(t *testing.T) {
	os := NewObjectStore[MutableStruct]()

	// Create all the objects
	pointers := make([]ObjectPointer[MutableStruct], objectChunkSize*3)
	for i := range pointers {
		p, _ := os.New()
		pointers[i] = p
	}

	// Get each object and modify field
	for i, p := range pointers {
		s := os.Get(p)
		s.Field = i * 2
	}

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		s := os.Get(p)
		assert.Equal(t, i*2, s.Field)
	}
}
