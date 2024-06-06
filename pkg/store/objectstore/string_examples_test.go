package objectstore_test

import (
	"fmt"

	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

// Calling AllocStringFromString allocates a string and returns a RefString
// which acts like a conventional pointer through which you can retrieve the
// allocated string via RefString.Value()
func ExampleAllocFromString() {
	var store *objectstore.Store = objectstore.New()

	var ref objectstore.RefString = objectstore.AllocStringFromString(store, "allocated")

	// Set the first element in the allocated slice
	var s1 string = ref.Value()

	fmt.Printf("String of %q", s1)
	// Output: String of "allocated"
}

// Calling AllocStringFromBytes allocates a string and returns a RefString
// which acts like a conventional pointer through which you can retrieve the
// allocated string via RefString.Value()
func ExampleAllocStringFromBytes() {
	var store *objectstore.Store = objectstore.New()

	var ref objectstore.RefString = objectstore.AllocStringFromBytes(store, []byte("allocated"))

	// Set the first element in the allocated slice
	var s1 string = ref.Value()

	fmt.Printf("String of %q", s1)
	// Output: String of "allocated"
}

// You can allocate a RefString by passing in a number of strings to be concatenated together
func ExampleConcatStrings() {
	var store *objectstore.Store = objectstore.New()

	var ref objectstore.RefString = objectstore.ConcatStrings(store, "all", "oca", "ted")

	var s1 string = ref.Value()

	fmt.Printf("String of %q", s1)
	// Output: String of "allocated"
}

// You can append a string to an allocated RefString
func ExampleAppendString() {
	var store *objectstore.Store = objectstore.New()

	var ref1 objectstore.RefString = objectstore.AllocStringFromString(store, "allocated")

	// AppendString a second element to the string
	var ref2 objectstore.RefString = objectstore.AppendString(store, ref1, " and appended")

	var s2 string = ref2.Value()

	fmt.Printf("String of %q", s2)
	// Output: String of "allocated and appended"
}

// After call to append the old RefString is no longer valid for use
func ExampleAppendString_oldRef() {
	var store *objectstore.Store = objectstore.New()

	var ref1 objectstore.RefString = objectstore.AllocStringFromString(store, "allocated")

	// AppendString a second element to the string
	objectstore.AppendString(store, ref1, " and appended")

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("The RefString passed into AppendString cannot be used after")
		}
	}()

	ref1.Value()
	// Output: The RefString passed into AppendString cannot be used after
}
