package objectstore

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type deepBadStruct struct {
	badInt       *int
	deepBadField badStruct
}

type badStruct struct {
	badField string
}

type stringSmugglerStruct struct {
	reference Reference[string]
}

type manyPointers struct {
	chanField      chan int
	funcField      func(int) int
	interfaceField any
	mapField       map[int]int
	pointerField   *int
	sliceField     []int
	stringField    string
}

func TestBadTypes(t *testing.T) {
	// No arrays with pointers in them
	assert.EqualError(t, containsNoPointers[[32]badStruct](), "found pointer(s): [32](objectstore.badStruct)badField<string>")
	// No channels
	assert.EqualError(t, containsNoPointers[chan int](), "found pointer(s): <chan int>")
	// No functions
	assert.EqualError(t, containsNoPointers[func(int) int](), "found pointer(s): <func(int) int>")
	// No interfaces
	assert.EqualError(t, containsNoPointers[any](), "found pointer(s): <interface {}>")
	// No maps
	assert.EqualError(t, containsNoPointers[map[int]int](), "found pointer(s): <map[int]int>")
	// No pointer(s)
	assert.EqualError(t, containsNoPointers[*int](), "found pointer(s): <*int>")
	// No slices
	assert.EqualError(t, containsNoPointers[[]int](), "found pointer(s): <[]int>")
	// No strings
	assert.EqualError(t, containsNoPointers[string](), "found pointer(s): <string>")
	// No structs with any pointerful fields
	assert.EqualError(t, containsNoPointers[badStruct](), "found pointer(s): (objectstore.badStruct)badField<string>")
	assert.EqualError(t, containsNoPointers[deepBadStruct](), "found pointer(s): (objectstore.deepBadStruct)badInt<*int>,(objectstore.deepBadStruct)deepBadField(objectstore.badStruct)badField<string>")
	//assert.EqualError(t, containsNoPointers[stringSmugglerStruct](), "found pointer(s): ")
	// No unsafe pointer(s)
	assert.EqualError(t, containsNoPointers[unsafe.Pointer](), "found pointer(s): <unsafe.Pointer>")
	// We should find all of the bad fields in this struct
	assert.EqualError(t, containsNoPointers[manyPointers](), "found pointer(s): "+
		"(objectstore.manyPointers)chanField<chan int>,"+
		"(objectstore.manyPointers)funcField<func(int) int>,"+
		"(objectstore.manyPointers)interfaceField<interface {}>,"+
		"(objectstore.manyPointers)mapField<map[int]int>,"+
		"(objectstore.manyPointers)pointerField<*int>,"+
		"(objectstore.manyPointers)sliceField<[]int>,"+
		"(objectstore.manyPointers)stringField<string>")
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
