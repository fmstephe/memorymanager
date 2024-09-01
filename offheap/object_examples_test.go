package offheap_test

import (
	"fmt"

	"github.com/fmstephe/offheap/offheap"
)

// Calling AllocObject allocates an object and returns a RefObject which acts
// like a conventional pointer through which you can retrieve the allocated
// object via RefObject.Value()
func ExampleAllocObject() {
	var store *offheap.Store = offheap.New()

	var ref offheap.RefObject[int] = offheap.AllocObject[int](store)
	var i1 *int = ref.Value()

	var i2 *int = ref.Value()

	if i1 == i2 {
		fmt.Println("This is correct, i1 and i2 are pointers to the same int location")
	}
	// Output: This is correct, i1 and i2 are pointers to the same int location
}

// You can free memory used by an an allocated object by calling
// FreeObject(...). The RefObject can no longer be used, and the use of the
// actual object pointed to will have unpredicatable results.
func ExampleFreeObject() {
	var store *offheap.Store = offheap.New()

	var ref offheap.RefObject[int] = offheap.AllocObject[int](store)

	offheap.FreeObject(store, ref)
	// You must never use ref again
}

// You can free memory used by an an allocated object by calling
// FreeObject(...). The RefObject can no longer be used, and the use of the
// actual object pointed to will have unpredicatable results.
func ExampleFreeObject_useAfterFreePanics() {
	var store *offheap.Store = offheap.New()

	var ref offheap.RefObject[int] = offheap.AllocObject[int](store)

	offheap.FreeObject(store, ref)
	// You must never use ref again

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Use after free panics")
		}
	}()

	ref.Value()
	// Output: Use after free panics
}

// You can allocate objects of complex types, including types with fields which
// are also of type RefObject. This allows us to build large datastructures,
// like trees in this example.
func ExampleAllocObject_complexType() {
	type Node struct {
		left  offheap.RefObject[Node]
		right offheap.RefObject[Node]
	}

	var store *offheap.Store = offheap.New()

	var refParent offheap.RefObject[Node] = offheap.AllocObject[Node]((store))

	var parent *Node = refParent.Value()

	var refLeft offheap.RefObject[Node] = offheap.AllocObject[Node]((store))

	parent.left = refLeft

	var refRight offheap.RefObject[Node] = offheap.AllocObject[Node]((store))

	parent.right = refRight

	// Re-get the parent pointer
	var reGetParent *Node = refParent.Value()

	if reGetParent.left == refLeft && reGetParent.right == refRight {
		fmt.Println("The mutations of the parent Node are visible via the reference")
	}
	// Output: The mutations of the parent Node are visible via the reference
}

// You cannot allocate a string type. Strings contain pointers interally and
// are not allowed.
func ExampleAllocObject_badTypeString() {
	type BadStruct struct {
		//lint:ignore U1000 this field looks unused but is observed by reflection
		stringsHavePointers string
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Can't allocate strings")
		}
	}()

	var store *offheap.Store = offheap.New()
	offheap.AllocObject[BadStruct](store)
	// Output: Can't allocate strings
}

// You cannot allocate a map type. Maps contain pointers interally and
// are not allowed.
func ExampleAllocObject_badTypeMap() {
	type BadStruct struct {
		//lint:ignore U1000 this field looks unused but is observed by reflection
		mapsHavePointers map[int]int
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Can't allocate maps")
		}
	}()

	var store *offheap.Store = offheap.New()
	offheap.AllocObject[BadStruct](store)
	// Output: Can't allocate maps
}

// You cannot allocate a slice type. Slices contain pointers interally and
// are not allowed.
func ExampleAllocObject_badTypeSlice() {
	type BadStruct struct {
		//lint:ignore U1000 this field looks unused but is observed by reflection
		slicesHavePointers []int
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Can't allocate slices (as an object)")
		}
	}()

	var store *offheap.Store = offheap.New()
	offheap.AllocObject[BadStruct](store)
	// Output: Can't allocate slices (as an object)
}

// You cannot allocate a pointer type (obviously).
func ExampleAllocObject_badTypePointer() {
	type BadStruct struct {
		//lint:ignore U1000 this field looks unused but is observed by reflection
		pointersHavePointers *int
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Can't allocate pointers")
		}
	}()

	var store *offheap.Store = offheap.New()
	offheap.AllocObject[BadStruct](store)
	// Output: Can't allocate pointers
}
