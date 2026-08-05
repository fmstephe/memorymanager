// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap

import (
	"reflect"
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

func TestBadTypes_containsNoPointers(t *testing.T) {
	// No arrays with pointers in them
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[[32]badStruct]()), "found pointer(s): [32](offheap.badStruct)badField<string>")
	// No channels
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[chan int]()), "found pointer(s): <chan int>")
	// No functions
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[func(int) int]()), "found pointer(s): <func(int) int>")
	// No interfaces
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[any]()), "found pointer(s): <interface {}>")
	// No maps
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[map[int]int]()), "found pointer(s): <map[int]int>")
	// No pointer(s)
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[*int]()), "found pointer(s): <*int>")
	// No slices
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[[]int]()), "found pointer(s): <[]int>")
	// No strings
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[string]()), "found pointer(s): <string>")
	// No structs with any pointerful fields
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[badStruct]()), "found pointer(s): (offheap.badStruct)badField<string>")
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[deepBadStruct]()), "found pointer(s): (offheap.deepBadStruct)badInt<*int>,(offheap.deepBadStruct)deepBadField(offheap.badStruct)badField<string>")
	// assert.EqualError(t, containsNoPointers(reflect.TypeFor[stringSmugglerStruct]()), "found pointer(s): ")
	// No unsafe pointer(s)
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[unsafe.Pointer]()), "found pointer(s): <unsafe.Pointer>")
	// We should find all of the bad fields in this struct
	assert.EqualError(t, containsNoPointers(reflect.TypeFor[manyPointers]()), "found pointer(s): "+
		"(offheap.manyPointers)chanField<chan int>,"+
		"(offheap.manyPointers)funcField<func(int) int>,"+
		"(offheap.manyPointers)interfaceField<interface {}>,"+
		"(offheap.manyPointers)mapField<map[int]int>,"+
		"(offheap.manyPointers)pointerField<*int>,"+
		"(offheap.manyPointers)sliceField<[]int>,"+
		"(offheap.manyPointers)stringField<string>")
}

func TestBadTypes_checkTypes(t *testing.T) {
	checker := newTypeChecker()

	// No arrays with pointers in them} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[[32]badStruct]()) })
	// No channels} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[chan int]()) })
	// No functions} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[func(int) int]()) })
	// No interfaces} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[any]()) })
	// No maps} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[map[int]int]()) })
	// No pointer(s)} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[*int]()) })
	// No slices} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[[]int]()) })
	// No strings} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[string]()) })
	// No structs with any pointerful fields} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[badStruct]()) })
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[deepBadStruct]()) })
	// assert.Panics(t, func () { checker.checkType(reflect.TypeFor[stringSmugglerStruct]())} )
	// No unsafe pointer(s)} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[unsafe.Pointer]()) })
	// We should find all of the bad fields in this struct} )
	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[manyPointers]()) })
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

func TestGoodTypes_containsNoPointers(t *testing.T) {
	// bool is fine
	assert.Nil(t, containsNoPointers(reflect.TypeFor[bool]()))
	// ints are fine
	assert.Nil(t, containsNoPointers(reflect.TypeFor[int]()))
	// uints are fine
	assert.Nil(t, containsNoPointers(reflect.TypeFor[uint]()))
	// floats are fine
	assert.Nil(t, containsNoPointers(reflect.TypeFor[float32]()))
	// complex numbers are fine
	assert.Nil(t, containsNoPointers(reflect.TypeFor[complex64]()))
	// arrays are fine
	assert.Nil(t, containsNoPointers(reflect.TypeFor[[32]int]()))
	// structs with no pointerful fields are fine
	assert.Nil(t, containsNoPointers(reflect.TypeFor[goodStruct]()))
	assert.Nil(t, containsNoPointers(reflect.TypeFor[deepGoodStruct]()))
}

func TestGoodTypes_checkType(t *testing.T) {
	checker := newTypeChecker()

	// bool is fine
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[bool]()) })
	// ints are fine
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[int]()) })
	// uints are fine
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[uint]()) })
	// floats are fine
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[float32]()) })
	// complex numbers are fine
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[complex64]()) })
	// arrays are fine
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[[32]int]()) })
	// structs with no pointerful fields are fine
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[goodStruct]()) })
	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[deepGoodStruct]()) })
}

func TestTypeCaching(t *testing.T) {
	checker := newTypeChecker()

	types := []reflect.Type{
		reflect.TypeFor[bool](),
		reflect.TypeFor[int](),
		reflect.TypeFor[uint](),
		reflect.TypeFor[float32](),
		reflect.TypeFor[complex64](),
		reflect.TypeFor[goodStruct](),
		reflect.TypeFor[deepGoodStruct](),
	}

	for _, typ := range types {
		// Initially the type is not in the cache
		assert.NotContains(t, checker.typeCache, typ)
		// Check the type
		checker.checkType(typ)
		// The type is now in the cache
		assert.Contains(t, checker.typeCache, typ)
		// Check the type again, still works
		checker.checkType(typ)
	}
}

type goodRefObject struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference1 RefObject[int]
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference2 RefSlice[int]
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference3 RefSlice[[32]int]
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference4 RefString
}

type badRefObject struct {
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference1 RefObject[string]
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference2 RefObject[[]int]
	//lint:ignore U1000 this field looks unused but is observed by reflection
	reference3 RefSlice[string]
}

func TestGoodRefObjects_checkTypes(t *testing.T) {
	checker := newTypeChecker()

	assert.NotPanics(t, func() { checker.checkType(reflect.TypeFor[goodRefObject]()) })
}

func TestBadRefObjects_checkTypes(t *testing.T) {
	checker := newTypeChecker()

	assert.Panics(t, func() { checker.checkType(reflect.TypeFor[badRefObject]()) })
}
