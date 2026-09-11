// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap_test

import (
	"fmt"

	"github.com/fmstephe/memorymanager/offheap"
)

// Calling AllocStringFromString allocates a string and returns a RefString
// which acts like a conventional pointer through which you can retrieve the
// allocated string via RefString.Value()
func ExampleStore_AllocStringFromString() {
	var store *offheap.Store = offheap.New()

	var ref offheap.RefString = store.AllocStringFromString("allocated")

	// Set the first element in the allocated slice
	var s1 string = ref.Value()

	fmt.Printf("String of %q", s1)
	// Output: String of "allocated"
}

// Calling AllocStringFromBytes allocates a string and returns a RefString
// which acts like a conventional pointer through which you can retrieve the
// allocated string via RefString.Value()
func ExampleStore_AllocStringFromBytes() {
	var store *offheap.Store = offheap.New()

	var ref offheap.RefString = store.AllocStringFromBytes([]byte("allocated"))

	// Set the first element in the allocated slice
	var s1 string = ref.Value()

	fmt.Printf("String of %q", s1)
	// Output: String of "allocated"
}

// You can allocate a RefString by passing in a number of strings to be concatenated together
func ExampleStore_ConcatStrings() {
	var store *offheap.Store = offheap.New()

	var ref offheap.RefString = store.ConcatStrings("all", "oca", "ted")

	var s1 string = ref.Value()

	fmt.Printf("String of %q", s1)
	// Output: String of "allocated"
}

// You can append a string to an allocated RefString
func ExampleStore_AppendString() {
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefString = store.AllocStringFromString("allocated")

	// AppendString a second element to the string
	var ref2 offheap.RefString = store.AppendString(ref1, " and appended")

	var s2 string = ref2.Value()

	fmt.Printf("String of %q", s2)
	// Output: String of "allocated and appended"
}

// After call to append the old RefString is no longer valid for use
func ExampleStore_AppendString_oldRef() {
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefString = store.AllocStringFromString("allocated")

	// AppendString a second element to the string
	store.AppendString(ref1, " and appended")

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("The RefString passed into AppendString cannot be used after")
		}
	}()

	ref1.Value()
	// Output: The RefString passed into AppendString cannot be used after
}
