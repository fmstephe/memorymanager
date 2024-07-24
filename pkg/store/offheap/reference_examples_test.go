package offheap_test

import (
	"fmt"

	"github.com/fmstephe/location-system/pkg/store/offheap"
)

// Here we store and retrieve a RefString using a Reference typed container. We
// should note that instantiating a Reference type requires us to define a type
// for T. This type is unused because RefString, unlike RefObject and RefSlice,
// is not a parameterised type. This is awkward and inelegant, but survivable.
//
// This example exists simply to illustrate how the Reference type can be used.
// I don't think it's obvious without at least one example.
//
// TODO this example is actually pretty bad. We somehow need to show the ReferenceMap
// implementation inside the example. Otherwise the most important and
// complicated part is missing from the example.
func ExampleReference_refString() {
	var refStringMap *ReferenceMap[byte, offheap.RefString] = NewReferenceMap[byte, offheap.RefString]()
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefString = offheap.ConcatStrings(store, "ref string reference")

	refStringMap.Add(1, ref1)

	var refOut offheap.RefString = refStringMap.Get(1)
	fmt.Printf("Stored and retrieved %q", refOut.Value())
	// Output: Stored and retrieved "ref string reference"
}

func ExampleReference_refString2() {
	var stringReference offheap.Reference[byte]
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefString = offheap.ConcatStrings(store, "ref string reference")

	stringReference = ref1

	fmt.Printf("Stored and retrieved %q", stringReference.value())
	// Output: Stored and retrieved "ref string reference"
	//
}

// Here we store and retrieve a RefSlice[byte] using a Reference typed
// container.
//
// This example exists simply to illustrate how the Reference type can be used.
// I don't think it's obvious without at least one example.
//
// TODO this example is actually pretty bad. We somehow need to show the ReferenceMap
// implementation inside the example. Otherwise the most important and
// complicated part is missing from the example.
func ExampleReference_refSlice() {
	var refSliceMap *ReferenceMap[byte, offheap.RefSlice[byte]] = NewReferenceMap[byte, offheap.RefSlice[byte]]()
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefSlice[byte] = offheap.ConcatSlices[byte](store, []byte("ref slice reference"))

	refSliceMap.Add(1, ref1)

	var refOut offheap.RefSlice[byte] = refSliceMap.Get(1)
	fmt.Printf("Stored and retrieved %q", refOut.Value())
	// Output: Stored and retrieved "ref slice reference"
}

// Here we store and retrieve a RefObject[int] using a Reference typed
// container.
//
// This example exists simply to illustrate how the Reference type can be used.
// I don't think it's obvious without at least one example.
//
// TODO this example is actually pretty bad. We somehow need to show the ReferenceMap
// implementation inside the example. Otherwise the most important and
// complicated part is missing from the example.
func ExampleReference_refObject() {
	var refObjectMap *ReferenceMap[int, offheap.RefObject[int]] = NewReferenceMap[int, offheap.RefObject[int]]()
	var store *offheap.Store = offheap.New()

	var ref1 offheap.RefObject[int] = offheap.AllocObject[int](store)
	var intValue *int = ref1.Value()
	*intValue = 127

	refObjectMap.Add(1, ref1)

	var refOut offheap.RefObject[int] = refObjectMap.Get(1)
	fmt.Printf("Stored and retrieved %v", *(refOut.Value()))
	// Output: Stored and retrieved 127
}

type ReferenceMap[T any, R offheap.Reference[T]] struct {
	rMap map[int]R
}

func NewReferenceMap[T any, R offheap.Reference[T]]() *ReferenceMap[T, R] {
	return &ReferenceMap[T, R]{
		rMap: make(map[int]R),
	}
}

func (f *ReferenceMap[T, R]) Add(hash int, value R) {
	f.rMap[hash] = value
}

func (f *ReferenceMap[T, R]) Get(hash int) R {
	return f.rMap[hash]
}
