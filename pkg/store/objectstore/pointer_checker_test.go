package objectstore

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type deepBadStruct struct {
	goodField int
	badField  badStruct
}

type badStruct struct {
	badField string
}

type stringSmugglerStruct struct {
	reference Reference[string]
}

func TestBadTypes(t *testing.T) {
	// No arrays with pointers in them
	assert.NotNil(t, containsNoPointers[[32]deepBadStruct]())
	// No channels
	assert.NotNil(t, containsNoPointers[chan int]())
	// No functions
	assert.NotNil(t, containsNoPointers[func(int) int]())
	// No interfaces
	assert.NotNil(t, containsNoPointers[any]())
	// No maps
	assert.NotNil(t, containsNoPointers[map[int]int]())
	// No pointers
	assert.NotNil(t, containsNoPointers[*int]())
	// No slices
	assert.NotNil(t, containsNoPointers[[]int]())
	// No strings
	assert.NotNil(t, containsNoPointers[string]())
	// No structs with any pointerful fields
	assert.NotNil(t, containsNoPointers[badStruct]())
	assert.NotNil(t, containsNoPointers[deepBadStruct]())
	//assert.NotNil(t, containsNoPointers[stringSmugglerStruct]())
	// No unsafe pointers
	assert.NotNil(t, containsNoPointers[unsafe.Pointer]())
}

type deepGoodStruct struct {
	boolField bool
	deepField goodStruct
}

type goodStruct struct {
	intField       int
	floatField     float64
	referenceField Reference[goodStruct]
}

func TestGoodTypes(t *testing.T) {
	// bool is fine
	assert.Nil(t, containsNoPointers[bool]())
	// ints are fine
	assert.Nil(t, containsNoPointers[int]())
	// uints are fine
	assert.Nil(t, containsNoPointers[uint]())
	// floats are fine
	assert.Nil(t, containsNoPointers[float32]())
	// complex numbers are fine
	assert.Nil(t, containsNoPointers[complex64]())
	// arrays are fine
	assert.Nil(t, containsNoPointers[[32]int]())
	// structs with no pointerful fields are fine
	assert.Nil(t, containsNoPointers[goodStruct]())
	assert.Nil(t, containsNoPointers[deepGoodStruct]())
}
