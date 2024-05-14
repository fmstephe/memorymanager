package objectstore

import (
	"fmt"
	"testing"

	"github.com/fmstephe/flib/fmath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// These tests are a bit fragile, as we have to _carefully_ only allocate
// objects of each size class only once. Because we track the number of slabs
// allocated as well as raw/reused allocations asserting the correct metrics
// quickly becomes difficult when we exercise the same size class multiple
// times.
func Test_Slice_SizedStats(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, capacity := range []int{
		0,
		1 << 1,
		1 << 2,
		1 << 5,
		(1 << 5) + 1,
		1 << 9,
		(1 << 9) + 1,
		1 << 14,
		(1 << 14) + 1,
	} {
		t.Run("", func(t *testing.T) {
			expectedStats := StatsForSlice[MutableStruct](os, capacity)

			r1, _ := AllocSlice[MutableStruct](os, capacity, capacity)
			r2, _ := AllocSlice[MutableStruct](os, capacity, capacity)
			FreeSlice[MutableStruct](os, r1)
			r3, _ := AllocSlice[MutableStruct](os, capacity, capacity)
			FreeSlice[MutableStruct](os, r2)
			FreeSlice[MutableStruct](os, r3)

			expectedStats.Allocs = 3
			expectedStats.Frees = 3
			expectedStats.RawAllocs = 2
			expectedStats.Reused = 1

			conf := ConfForSlice[MutableStruct](os, capacity)

			if conf.ObjectsPerSlab > 1 {
				// Only expect one slab to be allocated for smaller objects
				expectedStats.Slabs = 1
			} else {
				// Larger objects will require a slab per allocation
				expectedStats.Slabs = 2
			}

			actualStats := StatsForSlice[MutableStruct](os, capacity)

			assert.Equal(t, expectedStats, actualStats, "Bad stats for %d sized slice", capacity)
		})
	}
}

func Test_Slice_AppendInsufficientCapacity(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, length := range []int{
		0,
		1,
		2,
		3,
		1 << 1,
		1 << 2,
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
		refInit, initSlice := AllocSlice[int](os, length, length)
		require.Equal(t, length, len(initSlice))
		require.Equal(t, length, cap(initSlice))

		const appendValue = 99

		refAppend := Append[int](os, refInit, appendValue)

		// Assert that refAppend contains the new value
		appendSlice := refAppend.Value()
		require.Equal(t, length+1, len(appendSlice))
		require.Equal(t, int(fmath.NxtPowerOfTwo(int64(length+1))), cap(appendSlice))
		require.Equal(t, appendValue, appendSlice[len(appendSlice)-1])

		// Show that refInit has been freed
		require.Panics(t, func() { refInit.Value() })
	}
}

func Test_Slice_AppendSufficientCapacity(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, length := range []int{
		0,
		1,
		2,
		3,
		1 << 1,
		1 << 2,
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
		// Test with a range off spare slice capacity
		for _, extraCapacity := range []int{1, 2, 5, 7, 16} {
			capacity := length + extraCapacity

			refInit, initSlice := AllocSlice[int](os, length, capacity)
			// Assert the allocated slice works properly
			require.Equal(t, length, len(initSlice))
			require.Equal(t, capacity, cap(initSlice))

			const appendValue = 99

			refAppend := Append[int](os, refInit, appendValue)

			// Assert that refAppend contains the new value
			appendSlice := refAppend.Value()
			require.Equal(t, length+1, len(appendSlice))
			require.Equal(t, capacity, cap(appendSlice))
			require.Equal(t, appendValue, appendSlice[len(appendSlice)-1])

			// Show that refInit has not been freed
			require.NotPanics(t, func() { refInit.Value() })
		}
	}
}
