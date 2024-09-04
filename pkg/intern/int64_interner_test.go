package intern

import (
	"strconv"
	"testing"
)

func TestInt64Interner_Interned(t *testing.T) {
	interner := NewInt64Interner(Config{MaxLen: 64, MaxBytes: 1024}, 10)
	intVal := int64(1234)
	internedInt := strconv.FormatInt(intVal, 10)

	DoTestGenericInterner_Interned(t, interner, intVal, internedInt)
}

func TestInt64Interner_NotInternedMaxLen(t *testing.T) {
	interner := NewInt64Interner(Config{MaxLen: 3, MaxBytes: 1024}, 10)
	intVal := int64(1234)
	internedInt := strconv.FormatInt(intVal, 10)

	DoTestGenericInterner_NotInternedMaxLen(t, interner, intVal, internedInt)
}

func TestInt64Interner_NotInternedMaxBytes(t *testing.T) {
	interner := NewInt64Interner(Config{MaxLen: 64, MaxBytes: 3}, 10)
	intVal := int64(1234)
	internedInt := strconv.FormatInt(intVal, 10)

	DoTestGenericInterner_NotInternedMaxBytes(t, interner, intVal, internedInt)
}

/*
// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new ints, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned int values.
func TestInt64Interner_Complex(t *testing.T) {
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
func TestInt64Interner_NoAllocations(t *testing.T) {
	interner := NewInt64Interner(Config{MaxLen: 0, MaxBytes: 0}, 10)

	ints := make([]int64, 10_000)
	for i := range ints {
		ints[i] = int64(i)
	}

	DoTestGenericInterner_NoAllocations(t, interner, ints)
}
