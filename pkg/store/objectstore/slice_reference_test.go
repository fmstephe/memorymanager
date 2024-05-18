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

func Test_Slice_Append(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, length := range []int{
		0,
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
		9,
		10,
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
		for _, extraCapacity := range []int{0, 1, 2, 3, 4, 5, 6, 7, 16, 100} {
			t.Run(fmt.Sprintf("length %d extra capacity %d", length, extraCapacity), func(t *testing.T) {
				const initValue = 0x11
				const appendValue = 0x22

				capacity := length + extraCapacity

				refInit, initSlice := AllocSlice[int64](os, length, capacity)
				// Assert the allocated slice works properly
				require.Equal(t, length, len(initSlice))
				require.Equal(t, capacity, cap(initSlice))

				expectedSlice := make([]int64, length, capacity)
				for i := range initSlice {
					initSlice[i] = initValue
					expectedSlice[i] = initValue
				}

				refAppend := Append[int64](os, refInit, appendValue)
				expectedSlice = append(expectedSlice, appendValue)

				resultSlice := refAppend.Value()
				require.Equal(t, len(expectedSlice), len(resultSlice))
				// If the existing capacity is enough, it is
				// unchanged. If the capacity is not enough we
				// round up to a power of two which is large
				// enough
				requiredCapacity := int(fmath.NxtPowerOfTwo(int64(length + 1)))
				require.Equal(t, max(capacity, requiredCapacity), cap(resultSlice))

				// Assert the contents of the slice is correct
				require.Equal(t, expectedSlice, resultSlice)

				// Assert that the original reference has been invalidated
				require.Panics(t, func() { refInit.Value() })
			})
		}
	}
}

func Test_Slice_AppendSlice(t *testing.T) {
	os := New()
	defer os.Destroy()

	for _, length := range []int{
		0,
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
		9,
		10,
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
		for _, extraCapacity := range []int{0, 1, 2, 3, 4, 5, 7, 16, 100} {
			for _, appendSize := range []int{0, 1, 2, 3, 4, 5, 7, 16, 100} {
				t.Run(fmt.Sprintf("length %d append %d extra capacity %d", length, appendSize, extraCapacity), func(t *testing.T) {
					const initValue = 0x11
					const appendValue = 0x22

					capacity := length + extraCapacity

					refInit, initSlice := AllocSlice[int64](os, length, capacity)
					// Assert the allocated slice works properly
					require.Equal(t, length, len(initSlice))
					require.Equal(t, capacity, cap(initSlice))

					appendSlice := make([]int64, appendSize)
					for i := range appendSlice {
						appendSlice[i] = appendValue
					}

					expectedSlice := make([]int64, length, capacity)
					for i := range initSlice {
						initSlice[i] = initValue
						expectedSlice[i] = initValue
					}

					refResult := AppendSlice[int64](os, refInit, appendSlice)
					expectedSlice = append(expectedSlice, appendSlice...)

					// Assert that refAppend contains the new value
					resultSlice := refResult.Value()
					require.Equal(t, len(expectedSlice), len(resultSlice))

					requiredCapacity := int(fmath.NxtPowerOfTwo(int64(length + appendSize)))
					switch {
					case appendSize == 0 && capacity == 0:
						// If the capacity is 0, and we don't append anything
						// then the capacity remains 0.
						require.Equal(t, 0, cap(resultSlice))
					case capacity < requiredCapacity:
						// The capacity will be grown to the required capacity for the append
						require.Equal(t, requiredCapacity, cap(resultSlice))
					case capacity >= requiredCapacity:
						// The capacity of the slice is
						// strictly larger than needed
						// for the append. Capacity is
						// unchanged.
						require.Equal(t, capacity, cap(resultSlice))
					}
					require.Equal(t, expectedSlice, resultSlice)

					// Assert that the original reference has been invalidated
					require.Panics(t, func() { refInit.Value() })
				})
			}
		}
	}
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

		r, resultSlice := ConcatSlices[int64](os, testCase.slices...)

		assert.Equal(t, expectedSlice, resultSlice)
		assert.Equal(t, expectedSlice, r.Value())
	}
}
