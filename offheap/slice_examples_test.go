// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package offheap_test

import (
	"fmt"

	"github.com/fmstephe/memorymanager/offheap"
)

// Calling AllocSlice allocates a slice and returns a RefSlice which acts like
// a conventional pointer through which you can retrieve the allocated slice
// via RefSlice.Value()
func ExampleAllocSlice() {
	var store *offheap.Store = offheap.New()

	var ref offheap.RefSlice[int] = offheap.AllocSlice[int](store, 1, 2)

	// Set the first element in the allocated slice
	var s1 []int = ref.Value()
	s1[0] = 1

	// Show that the first element write is visible to other reads
	var s2 []int = ref.Value()

	fmt.Printf("Slice of %v with length %d and capacity %d", s2, len(s2), cap(s1))
	// Output: Slice of [1] with length 1 and capacity 2
}

// You can allocate a RefSlice by passing in a number of slices to be concatenated together
func ExampleConcatSlices() {
	var store *offheap.Store = offheap.New()

	slice1 := []int{1, 2}
	slice2 := []int{3, 4}
	slice3 := []int{5, 6}

	var ref offheap.RefSlice[int] = offheap.ConcatSlices[int](store, slice1, slice2, slice3)

	var s1 []int = ref.Value()

	fmt.Printf("Slice of %v with length %d and capacity %d", s1, len(s1), cap(s1))
	// Output: Slice of [1 2 3 4 5 6] with length 6 and capacity 8
}

// You can append an element to an allocated RefSlice
func ExampleAppend() {
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefSlice[int] = offheap.AllocSlice[int](store, 1, 2)

	// Set the first element in the allocated slice
	var s1 []int = ref1.Value()
	s1[0] = 1

	// Append a second element to the slice
	var ref2 offheap.RefSlice[int] = offheap.Append(store, ref1, 2)

	var s2 []int = ref2.Value()

	fmt.Printf("Slice of %v with length %d and capacity %d", s2, len(s2), cap(s1))
	// Output: Slice of [1 2] with length 2 and capacity 2
}

// After call to append the old RefSlice is no longer valid for use
func ExampleAppend_oldRef() {
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefSlice[int] = offheap.AllocSlice[int](store, 1, 2)

	// Set the first element in the allocated slice
	var s1 []int = ref1.Value()
	s1[0] = 1

	// Append a second element to the slice
	offheap.Append(store, ref1, 2)

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("The RefSlice passed into Append cannot be used after")
		}
	}()

	ref1.Value()
	// Output: The RefSlice passed into Append cannot be used after
}

// You can append a slice to a RefSlice. This will create a new RefSlice with
// the original slice and the slice passed in appended together.
func ExampleAppendSlice() {
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefSlice[int] = offheap.AllocSlice[int](store, 1, 2)

	// Set the first element in the allocated slice
	var s1 []int = ref1.Value()
	s1[0] = 1

	// Append a second element to the slice
	var ref2 offheap.RefSlice[int] = offheap.AppendSlice(store, ref1, []int{2, 3})

	var s2 []int = ref2.Value()

	fmt.Printf("Slice of %v with length %d and capacity %d", s2, len(s2), cap(s2))
	// Output: Slice of [1 2 3] with length 3 and capacity 4
}

// After call to AppendSlice the old RefSlice is no longer valid for use.
func ExampleAppendSlice_oldRef() {
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefSlice[int] = offheap.AllocSlice[int](store, 1, 2)

	// Set the first element in the allocated slice
	var s1 []int = ref1.Value()
	s1[0] = 1

	// Append a second element to the slice
	offheap.AppendSlice(store, ref1, []int{2, 3})

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("The RefSlice passed into AppendSlice cannot be used after")
		}
	}()

	ref1.Value()
	// Output: The RefSlice passed into AppendSlice cannot be used after
}
