// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap

import (
	"fmt"
	"testing"

	"github.com/fmstephe/memorymanager/testpkg/testutil"
	"github.com/stretchr/testify/assert"
)

// Test that when we allocate a string, the correct value is stored and
// retrieved.
func Test_String_AllocateAndGet_Simple(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, ss.Destroy())
	}()

	// Create string
	value := "test string"

	// Allocate it
	refString := ss.AllocStringFromString(value)
	valueOutString := refString.Value()

	// Assert that we can get the correct string from the Reference
	assert.Equal(t, value, valueOutString)
	assert.Equal(t, value, refString.Value())

	// Allocate it
	refBytes := ss.AllocStringFromBytes([]byte(value))
	valueOutBytes := refBytes.Value()

	// Assert that we can get the correct string from the Reference
	assert.Equal(t, value, valueOutBytes)
	assert.Equal(t, value, refBytes.Value())
}

// Test that when we allocate a string, the correct value is stored and
// retrieved.  This is a more complex version of the test above, testing a wide
// range of string sizes
func Test_String_AllocateAndGet(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, ss.Destroy())
	}()

	rsm := testutil.NewRandomStringMaker()

	for _, length := range testSizeRanges {
		t.Run(fmt.Sprintf("Allocate and get %d", length), func(t *testing.T) {
			// Generate a string of the desired size
			value := rsm.MakeSizedString(length)

			// Allocate it using the string
			refString := ss.AllocStringFromString(value)
			valueOutString := refString.Value()

			// Assert that we can get the correct string from the Reference
			assert.Equal(t, value, valueOutString)
			assert.Equal(t, value, refString.Value())

			// Allocate it using bytes
			refBytes := ss.AllocStringFromBytes([]byte(value))
			valueOutBytes := refBytes.Value()

			// Assert that we can get the correct string from the Reference
			assert.Equal(t, value, valueOutBytes)
			assert.Equal(t, value, refBytes.Value())
		})
	}
}

// Demonstrate that we can create a string then free it. If we call Value()
// on the freed RefStr call will panic
func Test_String_NewFreeGet_Panic(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, ss.Destroy())
	}()

	// Allocate and free a string value
	value := "test string"

	ref := ss.AllocStringFromString(value)
	ss.FreeString(ref)

	// Assert that calling Value() now panics
	assert.Panics(t, func() { ref.Value() })
}

// Demonstrate that we can create a string then free it twice. The second Free
// call will panic.
func Test_String_NewFreeFree_Panic(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, ss.Destroy())
	}()

	// Allocate and free a string value
	value := "test string"
	ref := ss.AllocStringFromString(value)
	ss.FreeString(ref)

	// Assert that calling FreeStr() now panics
	assert.Panics(t, func() { ss.FreeString(ref) })
}

// Demonstrate that when we double free a re-allocated string we still panic.
func Test_String_NewFreeAllocFree_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	value := "test string"
	r := os.AllocStringFromString(value)
	os.FreeString(r)
	// This will re-allocate the just-freed string
	os.AllocStringFromString(value)

	assert.Panics(t, func() { os.FreeString(r) })
}

// Demonstrate that when we call Value() on a re-allocated RefStr we still panic
func Test_String_NewFreeAllocGet_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	value := "test string"
	r := os.AllocStringFromString(value)
	os.FreeString(r)
	// This will re-allocate the just-freed string
	os.AllocStringFromString(value)

	assert.Panics(t, func() { r.Value() })
}

// These tests are a bit fragile, as we have to _carefully_ only allocate
// objects of each size class only once. Because we track the number of slabs
// allocated as well as raw/reused allocations asserting the correct metrics
// quickly becomes difficult when we exercise the same size class multiple
// times.
func Test_String_SizedStats(t *testing.T) {
	os := New()
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	rsm := testutil.NewRandomStringMaker()

	for _, length := range []int{
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
			expectedStats := os.StatsForString(length)

			value := rsm.MakeSizedString(length)

			r1 := os.AllocStringFromString(value)
			r2 := os.AllocStringFromString(value)
			os.FreeString(r1)
			r3 := os.AllocStringFromString(value)
			os.FreeString(r2)
			os.FreeString(r3)

			expectedStats.Allocs = 3
			expectedStats.Frees = 3
			expectedStats.RawAllocs = 2
			expectedStats.Reused = 1

			conf := os.ConfForString(length)

			if conf.ObjectsPerSlab > 1 {
				// Only expect one slab to be allocated for smaller objects
				expectedStats.Slabs = 1
			} else {
				// Larger objects will require a slab per allocation
				expectedStats.Slabs = 2
			}

			actualStats := os.StatsForString(length)

			assert.Equal(t, expectedStats, actualStats, "Bad stats for %d sized string", length)
		})
	}
}

func Test_String_AppendString(t *testing.T) {
	os := New()
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	rsm := testutil.NewRandomStringMaker()

	for _, firstLength := range testSizeRanges {
		for _, secondLength := range testSizeRanges {
			t.Run(fmt.Sprintf("AppendString first %d second %d", firstLength, secondLength), func(t *testing.T) {
				firstStr := rsm.MakeSizedString(firstLength)
				secondStr := rsm.MakeSizedString(secondLength)

				expectedString := firstStr + secondStr

				firstRef := os.AllocStringFromString(firstStr)
				resultRef := os.AppendString(firstRef, secondStr)

				assert.Equal(t, expectedString, resultRef.Value())
				assert.Panics(t, func() { firstRef.Value() })
			})
		}
	}
}

func Test_String_ConcatStrings(t *testing.T) {
	os := New()
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	for _, testCase := range []struct {
		strs []string
	}{
		// Empty cases
		{nil},
		{[]string{}},
		// Single string cases
		{[]string{
			"1",
		}},
		{[]string{
			"1, 2",
		}},
		{[]string{
			"1, 2, 3",
		}},
		{[]string{
			"1, 2, 3, 4",
		}},
		// Multi slice cases
		{[]string{
			"1, 2",
			"1",
		}},
		{[]string{
			"1",
			"1, 2, 3",
			"1, 2",
		}},
		{[]string{
			"1",
			"1, 2, 3",
			"1, 2",
			"1, 2, 3, 4",
		}},
		{[]string{
			"1",
			"1, 2, 3",
			"1, 2",
			"1, 2, 3, 4, 5",
			"1, 2, 3, 4",
		}},
	} {
		expectedString := ""
		for _, str := range testCase.strs {
			expectedString += str
		}

		r := os.ConcatStrings(testCase.strs...)
		resultString := r.Value()

		assert.Equal(t, expectedString, resultString)
		assert.Equal(t, expectedString, r.Value())
	}
}
