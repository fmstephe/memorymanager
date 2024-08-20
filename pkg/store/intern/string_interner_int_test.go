package intern

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestInternInt_Interned(t *testing.T) {
	interner := New(64, 1024)

	// A string is returned with the same value as intVal
	intVal := int64(1234)
	internedInt := interner.GetFromInt64(intVal)
	assert.Equal(t, strconv.FormatInt(intVal, 10), internedInt)

	// a new int value has been interned
	expectedStats := Stats{interned: 1}
	stats := interner.GetIntStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as intVal
	internedInt2 := interner.GetFromInt64(intVal)
	assert.Equal(t, strconv.FormatInt(intVal, 10), internedInt2)
	// The string returned uses the same memory allocation as the first
	// value returned i.e. the string is interned as is being reused as
	// intended
	assert.Equal(t, unsafe.StringData(internedInt), unsafe.StringData(internedInt2))

	// An interned string has been returned
	expectedStats = Stats{interned: 1, returned: 1}
	stats = interner.GetIntStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternInt_NotInternedMaxLen(t *testing.T) {
	interner := New(3, 1024)

	// A string is returned with the same value as intVal
	intVal := int64(1234)
	notInternedInt := interner.GetFromInt64(intVal)
	assert.Equal(t, strconv.FormatInt(intVal, 10), notInternedInt)

	// The int passed in was too long, so maxLenExceeded should be recorded
	expectedStats := Stats{maxLenExceeded: 1}
	stats := interner.GetIntStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as intVal
	notInternedInt2 := interner.GetFromInt64(intVal)
	assert.Equal(t, strconv.FormatInt(intVal, 10), notInternedInt2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedInt), unsafe.StringData(notInternedInt2))

	// The int passed in was too long, so maxLenExceeded should be recorded
	expectedStats = Stats{maxLenExceeded: 2}
	stats = interner.GetIntStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternInt_NotInternedUsedInt(t *testing.T) {
	interner := New(64, 3)

	intVal := int64(1234)
	internedInt := interner.GetFromInt64(intVal)

	expectedStats := Stats{usedBytesExceeded: 1}
	stats := interner.GetIntStats()

	assert.Equal(t, strconv.FormatInt(intVal, 10), internedInt)
	assert.Equal(t, expectedStats, stats.Total)
}

// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new ints, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned int values.
func TestInternInt_Complex(t *testing.T) {
	interner := New(1024, 1024)
	numberOfInts := 100

	// When we intern all these ints, each one is unique and is interned
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal)
			internedInt := interner.GetFromInt64(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			interned: numberOfInts,
		}
		stats := interner.GetIntStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these ints again, they are already
	// interned and their interned values are returned to us
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal)
			internedInt := interner.GetFromInt64(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			interned: numberOfInts,
			returned: numberOfInts,
		}
		stats := interner.GetIntStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Fill up the rest of the bytes so they are all used up
	{
		usedBytes := interner.GetIntStats().UsedBytes
		bytesRemaining := 1024 - usedBytes

		filler := make([]byte, bytesRemaining-1)
		fillerStr := interner.GetFromBytes(filler)
		assert.Equal(t, string(filler), fillerStr)

		expectedStats := Stats{
			interned: numberOfInts,
			returned: numberOfInts,
		}
		stats := interner.GetIntStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern new floats there aren't enough bytes left
	// to intern any of them
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal + numberOfInts)
			internedInt := interner.GetFromInt64(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			interned:          numberOfInts,
			returned:          numberOfInts,
			usedBytesExceeded: numberOfInts,
		}
		stats := interner.GetIntStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Finally we attempt to intern the ints again, they are already
	// interned and their interned values are returned to us
	{
		for intVal := range numberOfInts {
			intVal64 := int64(intVal)
			internedInt := interner.GetFromInt64(intVal64)
			assert.Equal(t, strconv.FormatInt(intVal64, 10), internedInt)
		}

		expectedStats := Stats{
			interned:          numberOfInts,
			returned:          numberOfInts * 2,
			usedBytesExceeded: numberOfInts,
		}
		stats := interner.GetIntStats()
		assert.Equal(t, expectedStats, stats.Total)
	}
}
