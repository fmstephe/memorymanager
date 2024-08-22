package intern

import (
	"math"
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestInternFloat_Interned_NaN(t *testing.T) {
	interner := New(64, 1024)

	// A string is returned with the same value as floatVal
	floatVal := math.NaN()
	internedFloat := interner.GetFromFloat64(floatVal)
	assert.Equal(t, "NaN", internedFloat)

	// a new float value has been interned
	expectedStats := Stats{returned: 1}
	stats := interner.GetFloatStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternFloat_Interned(t *testing.T) {
	interner := New(64, 1024)

	// A string is returned with the same value as floatVal
	floatVal := float64(12.34)
	internedFloat := interner.GetFromFloat64(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)

	// a new float value has been interned
	expectedStats := Stats{interned: 1}
	stats := interner.GetFloatStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as floatVal
	internedFloat2 := interner.GetFromFloat64(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat2)
	// The string returned uses the same memory allocation as the first
	// value returned i.e. the string is interned as is being reused as
	// intended
	assert.Equal(t, unsafe.StringData(internedFloat), unsafe.StringData(internedFloat2))

	// An interned string has been returned
	expectedStats = Stats{interned: 1, returned: 1}
	stats = interner.GetFloatStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternFloat_NotInternedMaxLen(t *testing.T) {
	interner := New(3, 1024)

	// A string is returned with the same value as floatVal
	floatVal := float64(12.34)
	notInternedFloat := interner.GetFromFloat64(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat)

	// The float passed in was too long, so maxLenExceeded should be recorded
	expectedStats := Stats{maxLenExceeded: 1}
	stats := interner.GetFloatStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as floatVal
	notInternedFloat2 := interner.GetFromFloat64(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedFloat), unsafe.StringData(notInternedFloat2))

	// The float passed in was too long, so maxLenExceeded should be recorded
	expectedStats = Stats{maxLenExceeded: 2}
	stats = interner.GetFloatStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternFloat_NotInternedUsedBytes(t *testing.T) {
	interner := New(64, 3)

	// A string is returned with the same value as floatVal
	floatVal := float64(12.34)
	notInternedFloat := interner.GetFromFloat64(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat)

	// The float passed in was too long, so usedBytesExceeded should be recorded
	expectedStats := Stats{usedBytesExceeded: 1}
	stats := interner.GetFloatStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as floatVal
	notInternedFloat2 := interner.GetFromFloat64(floatVal)
	assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), notInternedFloat2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedFloat), unsafe.StringData(notInternedFloat2))

	// The float passed in was too long, so usedBytesExceeded should be recorded
	expectedStats = Stats{usedBytesExceeded: 2}
	stats = interner.GetFloatStats()
	assert.Equal(t, expectedStats, stats.Total)
}

// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new floats, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned float values.
func TestInternFloat_Complex(t *testing.T) {
	interner := New(1024, 1024)
	numberOfFloats := 100

	// When we intern all these floats each one is unique and is interned
	{
		for intVal := range numberOfFloats {
			floatVal := float64(intVal) + 0.1234
			internedFloat := interner.GetFromFloat64(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			interned: numberOfFloats,
		}
		stats := interner.GetFloatStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these floats again, they are already
	// interned and their interned values are returned to us
	{
		for intVal := range numberOfFloats {
			floatVal := float64(intVal) + 0.1234
			internedFloat := interner.GetFromFloat64(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			interned: numberOfFloats,
			returned: numberOfFloats,
		}
		stats := interner.GetFloatStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Fill up the rest of the bytes so they are all used up
	{
		usedBytes := interner.GetFloatStats().UsedBytes
		bytesRemaining := 1024 - usedBytes

		filler := make([]byte, bytesRemaining-1)
		fillerStr := interner.GetFromBytes(filler)
		assert.Equal(t, string(filler), fillerStr)

		expectedStats := Stats{
			interned: numberOfFloats,
			returned: numberOfFloats,
		}
		stats := interner.GetFloatStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern new floats there aren't enough bytes left
	// to intern any of them
	{
		for intVal := range numberOfFloats {
			floatVal := float64(intVal+numberOfFloats) + 0.1234
			internedFloat := interner.GetFromFloat64(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			interned:          numberOfFloats,
			returned:          numberOfFloats,
			usedBytesExceeded: numberOfFloats,
		}
		stats := interner.GetFloatStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Finally we attempt to intern the floats again, they are already
	// interned and their interned values are returned to us
	{
		for intVal := range numberOfFloats {
			floatVal := float64(intVal) + 0.1234
			internedFloat := interner.GetFromFloat64(floatVal)
			assert.Equal(t, strconv.FormatFloat(floatVal, 'f', -1, 64), internedFloat)
		}

		expectedStats := Stats{
			interned:          numberOfFloats,
			returned:          numberOfFloats * 2,
			usedBytesExceeded: numberOfFloats,
		}
		stats := interner.GetFloatStats()
		assert.Equal(t, expectedStats, stats.Total)
	}
}
