package objectstore

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that when we allocate a string, the correct value is stored and
// retrieved.
func Test_String_AllocateAndGet_Simple(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Create string
	value := "test string"

	// Allocate it
	refString, valueOutString := AllocStringFromString(ss, value)

	// Assert that we can get the correct string from the Reference
	assert.Equal(t, value, valueOutString)
	assert.Equal(t, value, refString.Value())

	// Allocate it
	refBytes, valueOutBytes := AllocStringFromBytes(ss, []byte(value))

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
		t.Run(fmt.Sprintf("Allocate and get %d", length), func(t *testing.T) {
			// Generate a string of the desired size
			value := makeSizedString(length)

			// Allocate it using the string
			refString, valueOutString := AllocStringFromString(ss, value)

			// Assert that we can get the correct string from the Reference
			assert.Equal(t, value, valueOutString)
			assert.Equal(t, value, refString.Value())

			// Allocate it using bytes
			refBytes, valueOutBytes := AllocStringFromBytes(ss, []byte(value))

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
		ss.Destroy()
	}()

	// Allocate and free a string value
	value := "test string"
	ref, _ := AllocStringFromString(ss, value)
	FreeStr(ss, ref)

	// Assert that calling Value() now panics
	assert.Panics(t, func() { ref.Value() })
}

// Demonstrate that we can create a string then free it twice. The second Free
// call will panic.
func Test_String_NewFreeFree_Panic(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Allocate and free a string value
	value := "test string"
	ref, _ := AllocStringFromString(ss, value)
	FreeStr(ss, ref)

	// Assert that calling FreeStr() now panics
	assert.Panics(t, func() { FreeStr(ss, ref) })
}

// Demonstrate that when we double free a re-allocated string we still panic.
func Test_String_NewFreeAllocFree_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	value := "test string"
	r, _ := AllocStringFromString(os, value)
	FreeStr(os, r)
	// This will re-allocate the just-freed string
	AllocStringFromString(os, value)

	assert.Panics(t, func() { FreeStr(os, r) })
}

// Demonstrate that when we call Value() on a re-allocated RefStr we still panic
func Test_String_NewFreeAllocGet_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	value := "test string"
	r, _ := AllocStringFromString(os, value)
	FreeStr(os, r)
	// This will re-allocate the just-freed string
	AllocStringFromString(os, value)

	assert.Panics(t, func() { r.Value() })
}

// These tests are a bit fragile, as we have to _carefully_ only allocate
// objects of each size class only once. Because we track the number of slabs
// allocated as well as raw/reused allocations asserting the correct metrics
// quickly becomes difficult when we exercise the same size class multiple
// times.
func Test_String_SizedStats(t *testing.T) {
	os := New()
	defer os.Destroy()

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
			expectedStats := StatsForString(os, length)

			value := makeSizedString(length)

			r1, _ := AllocStringFromString(os, value)
			r2, _ := AllocStringFromString(os, value)
			FreeStr(os, r1)
			r3, _ := AllocStringFromString(os, value)
			FreeStr(os, r2)
			FreeStr(os, r3)

			expectedStats.Allocs = 3
			expectedStats.Frees = 3
			expectedStats.RawAllocs = 2
			expectedStats.Reused = 1

			conf := ConfForString(os, length)

			if conf.ObjectsPerSlab > 1 {
				// Only expect one slab to be allocated for smaller objects
				expectedStats.Slabs = 1
			} else {
				// Larger objects will require a slab per allocation
				expectedStats.Slabs = 2
			}

			actualStats := StatsForString(os, length)

			assert.Equal(t, expectedStats, actualStats, "Bad stats for %d sized string", length)
		})
	}
}

func Test_String_ConcatStrings(t *testing.T) {
	os := New()
	defer os.Destroy()

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

		r, resultString := ConcatStrings(os, testCase.strs...)

		assert.Equal(t, expectedString, resultString)
		assert.Equal(t, expectedString, r.Value())
	}
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var strRand = rand.New(rand.NewSource(1))

func makeSizedString(length int) string {
	builder := strings.Builder{}
	builder.Grow(length)
	for range length {
		builder.WriteByte(letters[strRand.Intn(len(letters))])
	}
	return builder.String()
}
