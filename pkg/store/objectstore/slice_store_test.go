package objectstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that when we allocate a slice, the correct value is stored and
// retrieved.
func Test_Slice_AllocateModifyAndGet(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Allocate it
	ref, value := AllocSlice[MutableStruct](ss, 10, 20)

	// Assert that the len and cap are as expected
	assert.Equal(t, 10, len(value))
	assert.Equal(t, 20, cap(value))

	// Mutate the elements of the slice, and the copied slice
	for i := range value {
		value[i].Field = i
	}

	// Copy slice - to allow for comparison
	valueCopy := make([]MutableStruct, 10, 20)
	copy(valueCopy, value)

	// Assert that the mutations are visible to the next call to ref.Value()
	assert.Equal(t, valueCopy, ref.Value())
	// Assert that the original value points to the same memory location as
	// ref.Value()
	assert.Equal(t, &value[0], &(ref.Value())[0])
}

func Test_Slice_AllocateModifyAndGet_ManySizes(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	for _, length := range []int{
		0,
		1,
		2,
		3,
		4,
		(1 << 5) - 1,
		1 << 5,
		(1 << 5) + 1,
		(1 << 9) - 1,
		1 << 9,
		(1 << 9) + 1,
		(1 << 14) - 1,
		1 << 14,
		(1 << 14) + 1,
	} {
		t.Run(fmt.Sprintf("Allocate and get Slice %d", length), func(t *testing.T) {
			// Allocate it
			ref, value := AllocSlice[MutableStruct](ss, length, length)

			// Assert that the len and cap are as expected
			assert.Equal(t, length, len(value))
			assert.Equal(t, length, cap(value))

			// Mutate the elements of the slice, and the copied slice
			for i := range value {
				value[i].Field = i
			}

			// Copy slice - to allow for comparison
			valueCopy := make([]MutableStruct, length)
			copy(valueCopy, value)

			// Assert that the mutations are visible to the next call to ref.Value()
			assert.Equal(t, valueCopy, ref.Value())
			// Assert that the original value points to the same memory location as
			// ref.Value()
			if length > 0 {
				assert.Equal(t, &value[0], &(ref.Value())[0])
			}
		})
	}
}

// Demonstrate that we can create a slice then free it. If we call Value()
// on the freed RefStr call will panic
func Test_Slice_NewFreeGet_Panic(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Allocate and free a slice value
	ref, _ := AllocSlice[MutableStruct](ss, 10, 10)
	FreeSlice(ss, ref)

	// Assert that calling Value() now panics
	assert.Panics(t, func() { ref.Value() })
}

// Demonstrate that we can create a slice then free it twice. The second Free
// call will panic.
func Test_Slice_NewFreeFree_Panic(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Allocate and free a slice value
	ref, _ := AllocSlice[MutableStruct](ss, 10, 10)
	FreeSlice(ss, ref)

	// Assert that calling FreeSlice() now panics
	assert.Panics(t, func() { FreeSlice(ss, ref) })
}

// Demonstrate that when we double free a re-allocated slice we still panic.
func Test_Slice_NewFreeAllocFree_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r, _ := AllocSlice[MutableStruct](os, 10, 10)
	FreeSlice(os, r)
	// This will re-allocate the just-freed slice
	AllocSlice[MutableStruct](os, 10, 10)

	assert.Panics(t, func() { FreeSlice(os, r) })
}

// Demonstrate that when we call Value() on a re-allocated RefSlice we still panic
func Test_Slice_NewFreeAllocGet_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r, _ := AllocSlice[MutableStruct](os, 10, 10)
	FreeSlice(os, r)
	// This will re-allocate the just-freed slice
	AllocSlice[MutableStruct](os, 10, 10)

	assert.Panics(t, func() { r.Value() })
}
