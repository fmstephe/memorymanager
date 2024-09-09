// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap_test

import (
	"fmt"

	"github.com/fmstephe/memorymanager/offheap"
)

// Here we store and retrieve a RefString using a Reference typed container. We
// should note that instantiating a Reference type requires us to define a type
// for T. This type is unused because RefString, unlike RefObject and RefSlice,
// is not a parameterised type. This is awkward and inelegant, but survivable.
//
// This example exists simply to illustrate how the Reference type can be used.
// I don't think it's obvious without at least one example.
func ExampleReference_refString() {
	// The ReferenceMap type is capable of storing either RefString,
	// RefSlice or RefObject types
	type ReferenceMap[T any, R offheap.Reference[T]] struct {
		rMap map[int]R
	}

	// Here we instantiate a ReferenceMap which can store RefString references
	var refStringMap ReferenceMap[byte, offheap.RefString] = ReferenceMap[byte, offheap.RefString]{
		rMap: make(map[int]offheap.RefString),
	}

	// Create a RefString
	var store *offheap.Store = offheap.New()
	var ref offheap.RefString = offheap.ConcatStrings(store, "ref string reference")

	// Store and retrieve that RefString
	refStringMap.rMap[1] = ref
	var refOut offheap.RefString = refStringMap.rMap[1]

	fmt.Printf("Stored and retrieved %q", refOut.Value())
	// Output: Stored and retrieved "ref string reference"
}

// Here we store and retrieve a RefSlice[byte] using a Reference typed
// container.
//
// This example exists simply to illustrate how the Reference type can be used.
// I don't think it's obvious without at least one example.
func ExampleReference_refSlice() {
	// The ReferenceMap type is capable of storing either RefString,
	// RefSlice or RefObject types
	type ReferenceMap[T any, R offheap.Reference[T]] struct {
		rMap map[int]R
	}

	// Here we instantiate a ReferenceMap which can store RefSlice references
	var refSliceMap ReferenceMap[byte, offheap.RefSlice[byte]] = ReferenceMap[byte, offheap.RefSlice[byte]]{
		rMap: make(map[int]offheap.RefSlice[byte]),
	}

	// Create a RefSlice
	var store *offheap.Store = offheap.New()
	var ref offheap.RefSlice[byte] = offheap.ConcatSlices[byte](store, []byte("ref slice reference"))

	// Store and retrieve that RefSlice
	refSliceMap.rMap[1] = ref
	var refOut offheap.RefSlice[byte] = refSliceMap.rMap[1]

	fmt.Printf("Stored and retrieved %q", refOut.Value())
	// Output: Stored and retrieved "ref slice reference"
}

// Here we store and retrieve a RefObject[int] using a Reference typed
// container.
//
// This example exists simply to illustrate how the Reference type can be used.
// I don't think it's obvious without at least one example.
func ExampleReference_refObject() {
	// The ReferenceMap type is capable of storing either RefString,
	// RefSlice or RefObject types
	type ReferenceMap[T any, R offheap.Reference[T]] struct {
		rMap map[int]R
	}

	// Here we instantiate a ReferenceMap which can store RefObject references
	var refObjectMap ReferenceMap[int, offheap.RefObject[int]] = ReferenceMap[int, offheap.RefObject[int]]{
		rMap: make(map[int]offheap.RefObject[int]),
	}

	// Create a RefObject
	var store *offheap.Store = offheap.New()
	var ref offheap.RefObject[int] = offheap.AllocObject[int](store)
	var intValue *int = ref.Value()
	*intValue = 127

	// Store and retrieve that RefObject
	refObjectMap.rMap[1] = ref
	var refOut offheap.RefObject[int] = refObjectMap.rMap[1]

	fmt.Printf("Stored and retrieved %v", *(refOut.Value()))
	// Output: Stored and retrieved 127
}
