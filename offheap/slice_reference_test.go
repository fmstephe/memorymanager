// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package offheap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSizeRanges = []int{
	0,
	1,
	2,
	3,
	4,
	7,
	8,
	(1 << 5) - 1,
	1 << 5,
	(1 << 5) + 1,
	(1 << 14) - 1,
	1 << 14,
	(1 << 14) + 1,
}

// Test that when we allocate a slice, the correct value is stored and
// retrieved.
func Test_Slice_AllocateModifyAndGet(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Allocate it
	ref := AllocSlice[MutableStruct](ss, 10, 20)
	value := ref.Value()

	// Assert that the len and cap are as expected
	assert.Equal(t, 10, len(value))
	assert.Equal(t, capacityForSlice(20), cap(value))

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

	for _, length := range testSizeRanges {
		t.Run(fmt.Sprintf("Allocate and get Slice %d", length), func(t *testing.T) {
			// Allocate it
			ref := AllocSlice[MutableStruct](ss, length, length)
			value := ref.Value()

			// Assert that the len and cap are as expected
			assert.Equal(t, length, len(value))
			assert.Equal(t, capacityForSlice(length), cap(value))

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
	ref := AllocSlice[MutableStruct](ss, 10, 10)
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
	ref := AllocSlice[MutableStruct](ss, 10, 10)
	FreeSlice(ss, ref)

	// Assert that calling FreeSlice() now panics
	assert.Panics(t, func() { FreeSlice(ss, ref) })
}

// Demonstrate that when we double free a re-allocated slice we still panic.
func Test_Slice_NewFreeAllocFree_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r := AllocSlice[MutableStruct](os, 10, 10)
	FreeSlice(os, r)
	// This will re-allocate the just-freed slice
	AllocSlice[MutableStruct](os, 10, 10)

	assert.Panics(t, func() { FreeSlice(os, r) })
}

// Demonstrate that when we call Value() on a re-allocated RefSlice we still panic
func Test_Slice_NewFreeAllocGet_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r := AllocSlice[MutableStruct](os, 10, 10)
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

			r1 := AllocSlice[MutableStruct](os, capacity, capacity)
			r2 := AllocSlice[MutableStruct](os, capacity, capacity)
			FreeSlice[MutableStruct](os, r1)
			r3 := AllocSlice[MutableStruct](os, capacity, capacity)
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

func Test_Slice_Append(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, length := range testSizeRanges {
		for _, extraCapacity := range testSizeRanges {

			doSliceAppendTest[[0]byte](
				t,
				os,
				length,
				extraCapacity,
				func() [0]byte { return [0]byte{} },
				func() [0]byte { return [0]byte{} })

			doSliceAppendTest[[1]byte](
				t,
				os,
				length,
				extraCapacity,
				func() [1]byte { return [1]byte{0x11} },
				func() [1]byte { return [1]byte{0x22} })

			doSliceAppendTest[[2]byte](
				t,
				os,
				length,
				extraCapacity,
				func() [2]byte { return [2]byte{0x11, 0x11} },
				func() [2]byte { return [2]byte{0x22, 0x22} })

			doSliceAppendTest[[3]byte](
				t,
				os,
				length,
				extraCapacity,
				func() [3]byte { return [3]byte{0x11, 0x11, 0x11} },
				func() [3]byte { return [3]byte{0x22, 0x22, 0x22} })

			doSliceAppendTest[int64](
				t,
				os,
				length,
				extraCapacity,
				func() int64 { return 0x11 },
				func() int64 { return 0x22 })

		}
	}
}

func doSliceAppendTest[T any](t *testing.T, os *Store, length, extraCapacity int, initVal, appendVal func() T) {
	t.Run(fmt.Sprintf("type %T length %d extra capacity %d", *(new(T)), length, extraCapacity), func(t *testing.T) {
		capacity := length + extraCapacity

		refInit := AllocSlice[T](os, length, capacity)
		initSlice := refInit.Value()
		// Assert the allocated slice works properly
		require.Equal(t, length, len(initSlice))
		initCapacity := capacityForSlice(capacity)
		require.Equal(t, initCapacity, cap(initSlice))

		expectedSlice := make([]T, length, capacity)
		for i := range initSlice {
			initSlice[i] = initVal()
			expectedSlice[i] = initVal()
		}

		refAppend := Append[T](os, refInit, appendVal())
		expectedSlice = append(expectedSlice, appendVal())

		resultSlice := refAppend.Value()
		require.Equal(t, len(expectedSlice), len(resultSlice))
		// If the existing capacity is enough, it is
		// unchanged. If the capacity is not enough we
		// round up to a power of two which is large
		// enough
		expectedCapacity := max(initCapacity, capacityForSlice(length+1))
		require.Equal(t, expectedCapacity, cap(resultSlice))

		// Assert the contents of the slice is correct
		require.Equal(t, expectedSlice, resultSlice)

		// Assert that the original reference has been invalidated
		require.Panics(t, func() { refInit.Value() })
	})
}

func Test_Slice_AppendSlice(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, length := range testSizeRanges {
		for _, extraCapacity := range testSizeRanges {
			for _, appendSize := range testSizeRanges {
				doSliceAppendSliceTest[[0]byte](
					t,
					os,
					length,
					extraCapacity,
					appendSize,
					func() [0]byte { return [0]byte{} },
					func() [0]byte { return [0]byte{} })

				doSliceAppendSliceTest[[1]byte](
					t,
					os,
					length,
					extraCapacity,
					appendSize,
					func() [1]byte { return [1]byte{0x11} },
					func() [1]byte { return [1]byte{0x22} })
				doSliceAppendSliceTest[[2]byte](
					t,
					os,
					length,
					extraCapacity,
					appendSize,
					func() [2]byte { return [2]byte{0x11, 0x11} },
					func() [2]byte { return [2]byte{0x22, 0x22} })

				doSliceAppendSliceTest[[3]byte](
					t,
					os,
					length,
					extraCapacity,
					appendSize,
					func() [3]byte { return [3]byte{0x11, 0x11, 0x11} },
					func() [3]byte { return [3]byte{0x22, 0x22, 0x22} })

				doSliceAppendSliceTest[int64](
					t,
					os,
					length,
					extraCapacity,
					appendSize,
					func() int64 { return 0x11 },
					func() int64 { return 0x22 })

			}
		}
	}
}

func doSliceAppendSliceTest[T any](t *testing.T, os *Store, length, extraCapacity, appendSize int, initVal, appendVal func() T) {
	t.Run(fmt.Sprintf("type %T length %d append %d extra capacity %d", *(new(T)), length, appendSize, extraCapacity), func(t *testing.T) {
		capacity := length + extraCapacity

		refInit := AllocSlice[T](os, length, capacity)
		initSlice := refInit.Value()
		initCapacity := capacityForSlice(capacity)
		// Assert the allocated slice works properly
		require.Equal(t, length, len(initSlice))
		require.Equal(t, initCapacity, cap(initSlice))

		appendSlice := make([]T, appendSize)
		for i := range appendSlice {
			appendSlice[i] = appendVal()
		}

		expectedSlice := make([]T, length, capacity)
		for i := range initSlice {
			initSlice[i] = initVal()
			expectedSlice[i] = initVal()
		}

		refResult := AppendSlice[T](os, refInit, appendSlice)
		expectedSlice = append(expectedSlice, appendSlice...)

		// Assert that refAppend contains the new value
		resultSlice := refResult.Value()
		require.Equal(t, len(expectedSlice), len(resultSlice))

		expectedCapacity := max(initCapacity, capacityForSlice(length+appendSize))
		require.Equal(t, expectedCapacity, cap(resultSlice))

		require.Equal(t, expectedSlice, resultSlice)

		// Assert that the original reference has been invalidated
		require.Panics(t, func() { refInit.Value() })
	})
}

func Test_Slice_ConcatSlices(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, testCase := range []struct {
		slices [][]int64
	}{
		// Empty cases
		{nil},
		{[][]int64{}},
		// Single slice cases
		{[][]int64{
			[]int64{1},
		}},
		{[][]int64{
			[]int64{1, 2},
		}},
		{[][]int64{
			[]int64{1, 2, 3},
		}},
		{[][]int64{
			[]int64{1, 2, 3, 4},
		}},
		// Multi slice cases
		{[][]int64{
			[]int64{1, 2},
			[]int64{1},
		}},
		{[][]int64{
			[]int64{1},
			[]int64{1, 2, 3},
			[]int64{1, 2},
		}},
		{[][]int64{
			[]int64{1},
			[]int64{1, 2, 3},
			[]int64{1, 2},
			[]int64{1, 2, 3, 4},
		}},
		{[][]int64{
			[]int64{1},
			[]int64{1, 2, 3},
			[]int64{1, 2},
			[]int64{1, 2, 3, 4, 5},
			[]int64{1, 2, 3, 4},
		}},
	} {
		expectedSlice := []int64{}
		for _, slice := range testCase.slices {
			expectedSlice = append(expectedSlice, slice...)
		}

		r := ConcatSlices[int64](os, testCase.slices...)
		resultSlice := r.Value()

		assert.Equal(t, expectedSlice, resultSlice)
		assert.Equal(t, expectedSlice, r.Value())
	}
}
