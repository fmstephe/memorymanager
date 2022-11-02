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
		p, s := os.Alloc()
		s.Field = i
		pointers[i] = p
	}

	assert.Equal(t, len(pointers), os.AllocCount())
	assert.Equal(t, len(pointers), os.LiveCount())
	assert.Equal(t, 0, os.FreeCount())

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
		p, _ := os.Alloc()
		pointers[i] = p
	}

	assert.Equal(t, len(pointers), os.AllocCount())
	assert.Equal(t, len(pointers), os.LiveCount())
	assert.Equal(t, 0, os.FreeCount())

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

// Demonstrate that we can create an object, then free it. If we try to Get()
// the freed object ObjectStore panics
func Test_Object_NewFreeGet_Panic(t *testing.T) {
	os := NewObjectStore[MutableStruct]()
	p, _ := os.Alloc()
	os.Free(p)

	assert.Panics(t, func() { os.Get(p) })
}

// Demonstrate that we can create an object, then free it. If we try to Free()
// the freed object ObjectStore panics
func Test_Object_NewFreeFree_Panic(t *testing.T) {
	os := NewObjectStore[MutableStruct]()
	p, _ := os.Alloc()
	os.Free(p)

	assert.Panics(t, func() { os.Free(p) })
}

// Demonstrate that if we create a large number of objects, then free them,
// then allocate that same number again, we re-use the freed objects
func Test_Object_NewFreeNew_ReusesOldObjects(t *testing.T) {
	os := NewObjectStore[MutableStruct]()

	objectAllocations := objectChunkSize * 3

	// Create a large number of objects
	pointers := make([]ObjectPointer[MutableStruct], objectChunkSize*3)
	for i := range pointers {
		p, _ := os.Alloc()
		pointers[i] = p
	}

	// We have allocate one batch of objects
	assert.Equal(t, objectAllocations, os.AllocCount())
	// They are all live
	assert.Equal(t, objectAllocations, os.LiveCount())
	// Nothing has been freed
	assert.Equal(t, 0, os.FreeCount())
	// Internally 4 chunks have been created
	assert.Equal(t, 4, os.Chunks())

	// Free all of those objects
	for _, p := range pointers {
		os.Free(p)
	}

	// We have allocate one batch of objects
	assert.Equal(t, objectAllocations, os.AllocCount())
	// None are live
	assert.Equal(t, 0, os.LiveCount())
	// We have freed one batch of objects
	assert.Equal(t, objectAllocations, os.FreeCount())
	// Internally 4 chunks have been created
	assert.Equal(t, 4, os.Chunks())

	// Allocate the same number of objects again
	for range pointers {
		os.Alloc()
	}

	// We have allocated 2 batches of objects
	assert.Equal(t, 2*objectAllocations, os.AllocCount())
	// We have freed one batch
	assert.Equal(t, objectAllocations, os.LiveCount())
	// One batch is live
	assert.Equal(t, objectAllocations, os.FreeCount())
	// Internally 4 chunks have been created
	assert.Equal(t, 4, os.Chunks())
}
