package offheap

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type deepBadStruct struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	badInt *int
	//lint:ignore U1000 this field looks unused but is observed by reflection
	deepBadField badStruct
}

type badStruct struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	badField string
}

// We will create a test which uses this struct in the future
//
//lint:ignore U1000 this struct actually is unused - but it represents a real bug in our code
type stringSmugglerStruct struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference RefObject[string]
}

type manyPointers struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	chanField chan int
	//lint:ignore U1000 this field looks unused but is observed by reflection
	funcField func(int) int
	//lint:ignore U1000 this field looks unused but is observed by reflection
	interfaceField any
	//lint:ignore U1000 this field looks unused but is observed by reflection
	mapField map[int]int
	//lint:ignore U1000 this field looks unused but is observed by reflection
	pointerField *int
	//lint:ignore U1000 this field looks unused but is observed by reflection
	sliceField []int
	//lint:ignore U1000 this field looks unused but is observed by reflection
	stringField string
}

func TestBadTypes(t *testing.T) {
	// No arrays with pointers in them
	assert.EqualError(t, containsNoPointers[[32]badStruct](), "found pointer(s): [32](offheap.badStruct)badField<string>")
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
	assert.EqualError(t, containsNoPointers[badStruct](), "found pointer(s): (offheap.badStruct)badField<string>")
	assert.EqualError(t, containsNoPointers[deepBadStruct](), "found pointer(s): (offheap.deepBadStruct)badInt<*int>,(offheap.deepBadStruct)deepBadField(offheap.badStruct)badField<string>")
	// assert.EqualError(t, containsNoPointers[stringSmugglerStruct](), "found pointer(s): ")
	// No unsafe pointer(s)
	assert.EqualError(t, containsNoPointers[unsafe.Pointer](), "found pointer(s): <unsafe.Pointer>")
	// We should find all of the bad fields in this struct
	assert.EqualError(t, containsNoPointers[manyPointers](), "found pointer(s): "+
		"(offheap.manyPointers)chanField<chan int>,"+
		"(offheap.manyPointers)funcField<func(int) int>,"+
		"(offheap.manyPointers)interfaceField<interface {}>,"+
		"(offheap.manyPointers)mapField<map[int]int>,"+
		"(offheap.manyPointers)pointerField<*int>,"+
		"(offheap.manyPointers)sliceField<[]int>,"+
		"(offheap.manyPointers)stringField<string>")
}

type deepGoodStruct struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	boolField bool
	//lint:ignore U1000 this field looks unused but is observed by reflection
	deepField goodStruct
}

type goodStruct struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	intField int
	//lint:ignore U1000 this field looks unused but is observed by reflection
	floatField float64
	//lint:ignore U1000 this field looks unused but is observed by reflection
	referenceField RefObject[goodStruct]
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
