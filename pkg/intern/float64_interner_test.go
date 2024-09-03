package intern

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestFloat64Interner_Interned(t *testing.T) {
	interner := NewFloat64Interner(Config{MaxLen: 64, MaxBytes: 1024}, 'f', -1, 64)

	// A string is returned with the same value as floatVal
	floatVal := float64(12.34)
	internedFloat := interner.Get(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)

	// a new float value has been interned
	expectedStats := Stats{Interned: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as floatVal
	internedFloat2 := interner.Get(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat2)
	// The string returned uses the same memory allocation as the first
	// value returned i.e. the string is interned as is being reused as
	// intended
	assert.Equal(t, unsafe.StringData(internedFloat), unsafe.StringData(internedFloat2))

	// An interned string has been returned
	expectedStats = Stats{Interned: 1, Returned: 1}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestFloat64Interner_NotInternedMaxLen(t *testing.T) {
	interner := NewFloat64Interner(Config{MaxLen: 3, MaxBytes: 1024}, 'f', -1, 64)

	// A string is returned with the same value as floatVal
	floatVal := float64(12.34)
	notInternedFloat := interner.Get(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat)

	// The float passed in was too long, so maxLenExceeded should be recorded
	expectedStats := Stats{MaxLenExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as floatVal
	notInternedFloat2 := interner.Get(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedFloat), unsafe.StringData(notInternedFloat2))

	// The float passed in was too long, so maxLenExceeded should be recorded
	expectedStats = Stats{MaxLenExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestFloat64Interner_NotInternedUsedInt(t *testing.T) {
	interner := NewFloat64Interner(Config{MaxLen: 64, MaxBytes: 3}, 'f', -1, 64)

	// A string is returned with the same value as floatVal
	floatVal := float64(12.34)
	notInternedFloat := interner.Get(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat)

	// The float passed in was too long, so usedBytesExceeded should be recorded
	expectedStats := Stats{UsedBytesExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as floatVal
	notInternedFloat2 := interner.Get(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedFloat), unsafe.StringData(notInternedFloat2))

	// The float passed in was too long, so usedBytesExceeded should be recorded
	expectedStats = Stats{UsedBytesExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

/*
// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new floats, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned float values.
func TestFloat64Interner_Complex(t *testing.T) {
	interner := NewFloat64Interner(Config{MaxLen: 1024, MaxBytes: 1024}, 'f', -1, 64)
	numberOfFloats := 100

	// When we intern all these floats, each one is unique and is interned
	{
		for i := range numberOfFloats {
			floatVal := float64(i) + 0.123
			internedFloat := interner.Get(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			Interned: numberOfFloats,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these floats again, they are already
	// interned and their interned values are returned to us
	{
		for i := range numberOfFloats {
			floatVal := float64(i) + 0.123
			internedFloat := interner.Get(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			Interned: numberOfFloats,
			Returned: numberOfFloats,
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
		for i := range numberOfFloats {
			floatVal := float64(i+numberOfFloats) + 0.123
			internedFloat := interner.Get(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			Interned:          numberOfFloats,
			Returned:          numberOfFloats,
			UsedBytesExceeded: numberOfFloats,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Finally we attempt to intern the floats again, they are already
	// interned and their interned values are returned to us
	{
		for i := range numberOfFloats {
			floatVal := float64(i) + 0.123
			internedFloat := interner.Get(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			Interned:          numberOfFloats,
			Returned:          numberOfFloats * 2,
			UsedBytesExceeded: numberOfFloats,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}
}
*/

// Assert that getting a string, where the value has already been interned,
// does not allocate
func TestFloat64Interner_NoAllocations(t *testing.T) {
	interner := NewFloat64Interner(Config{MaxLen: 0, MaxBytes: 0}, 'f', -1, 64)

	floats := make([]float64, 10_000)
	for i := range floats {
		floats[i] = float64(i) + 0.123
	}

	for _, floatVal := range floats {
		interner.Get(floatVal)
	}

	avgAllocs := testing.AllocsPerRun(100, func() {
		for _, floatVal := range floats {
			interner.Get(floatVal)
		}
	})
	// getting strings for floats which have already been interned does not
	// allocate
	assert.Equal(t, 0.0, avgAllocs)
}
