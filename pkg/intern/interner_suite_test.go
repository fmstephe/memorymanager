// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package intern

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func DoTestGenericInterner_Interned[T any](t *testing.T, interner Interner[T], val T, strVal string) {
	t.Helper()

	// A string is returned with the same value as intVal
	internedVal := interner.Get(val)
	assert.Equal(t, strVal, internedVal)

	// a new int value has been interned
	expectedStats := Stats{Interned: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as intVal
	internedVal2 := interner.Get(val)
	assert.Equal(t, strVal, internedVal2)
	// The string returned uses the same memory allocation as the first
	// value returned i.e. the string is interned as is being reused as
	// intended
	assert.Same(t, unsafe.StringData(internedVal), unsafe.StringData(internedVal2))

	// An interned string has been returned
	expectedStats = Stats{Interned: 1, Returned: 1}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func DoTestGenericInterner_NotInternedMaxLen[T any](t *testing.T, interner Interner[T], val T, strVal string) {
	t.Helper()

	// A string is returned with the same value as intVal
	notInternedInt := interner.Get(val)
	assert.Equal(t, strVal, notInternedInt)

	// The int passed in was too long, so maxLenExceeded should be recorded
	expectedStats := Stats{MaxLenExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as intVal
	notInternedInt2 := interner.Get(val)
	assert.Equal(t, strVal, notInternedInt2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedInt), unsafe.StringData(notInternedInt2))

	// The int passed in was too long, so maxLenExceeded should be recorded
	expectedStats = Stats{MaxLenExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func DoTestGenericInterner_NotInternedMaxBytes[T any](t *testing.T, interner Interner[T], val T, strVal string) {
	t.Helper()

	// A string is returned with the same value as intVal
	notInternedInt := interner.Get(val)
	assert.Equal(t, strVal, notInternedInt)

	// The int passed in was too long, so usedBytesExceeded should be recorded
	expectedStats := Stats{UsedBytesExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as intVal
	notInternedInt2 := interner.Get(val)
	assert.Equal(t, strVal, notInternedInt2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedInt), unsafe.StringData(notInternedInt2))

	// The int passed in was too long, so usedBytesExceeded should be recorded
	expectedStats = Stats{UsedBytesExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

/*
// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new ints, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned int values.
func TestGenericInterner_Complex(t *testing.T) {
	interner := NewInt64Interner(Config{MaxLen: 1024, MaxBytes: 1024}, 10)
	numberOfInts := 100

	// When we intern all these ints, each one is unique and is interned
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal)
			internedInt := interner.Get(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			Interned: numberOfInts,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these ints again, they are already
	// interned and their interned values are returned to us
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal)
			internedInt := interner.Get(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			Interned: numberOfInts,
			Returned: numberOfInts,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	{
		// This hack pushes up the recorded used-bytes to the limit.
		// This means no other strings can be interned from here.
		interner.interner.controller.usedBytes.Store(1024 * 1024)
	}

	// When we attempt to intern new floats there aren't enough bytes left
	// to intern any of them
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal + numberOfInts)
			internedInt := interner.Get(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			Interned:          numberOfInts,
			Returned:          numberOfInts,
			UsedBytesExceeded: numberOfInts,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Finally we attempt to intern the ints again, they are already
	// interned and their interned values are returned to us
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal)
			internedInt := interner.Get(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			Interned:          numberOfInts,
			Returned:          numberOfInts * 2,
			UsedBytesExceeded: numberOfInts,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}
}
*/

// Assert that getting a string, where the value has already been interned,
// does not allocate
func DoTestGenericInterner_NoAllocations[T any](t *testing.T, interner Interner[T], vals []T) {
	t.Helper()

	for _, val := range vals {
		interner.Get(val)
	}

	avgAllocs := testing.AllocsPerRun(100, func() {
		for _, val := range vals {
			interner.Get(val)
		}
	})
	// getting strings for ints which have already been interned does not
	// allocate
	assert.Equal(t, 0.0, avgAllocs)
}
