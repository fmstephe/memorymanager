package intern

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInternInt_Interned(t *testing.T) {
	interner := New(64, 1024)

	intVal := int64(1234)
	internedInt := interner.GetFromInt64(int64(intVal))

	expectedStats := Stats{interned: 1}
	stats := interner.GetIntStats()

	assert.Equal(t, strconv.Itoa(int(intVal)), internedInt)
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternInt_NotInternedMaxLen(t *testing.T) {
	interner := New(3, 1024)

	intVal := int64(1234)
	internedInt := interner.GetFromInt64(int64(intVal))

	expectedStats := Stats{maxLenExceeded: 1}
	stats := interner.GetIntStats()

	assert.Equal(t, strconv.Itoa(int(intVal)), internedInt)
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternInt_NotInternedUsedInt(t *testing.T) {
	interner := New(64, 3)

	intVal := int64(1234)
	internedInt := interner.GetFromInt64(int64(intVal))

	expectedStats := Stats{usedBytesExceeded: 1}
	stats := interner.GetIntStats()

	assert.Equal(t, strconv.Itoa(int(intVal)), internedInt)
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
			internedInt := interner.GetFromInt64(int64(intVal))
			assert.Equal(t, strconv.Itoa(int(intVal)), internedInt)
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
			internedInt := interner.GetFromInt64(int64(intVal))
			assert.Equal(t, strconv.Itoa(int(intVal)), internedInt)
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
			intVal = intVal + numberOfInts
			internedInt := interner.GetFromInt64(int64(intVal))
			assert.Equal(t, strconv.Itoa(int(intVal)), internedInt)
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
			internedInt := interner.GetFromInt64(int64(intVal))
			assert.Equal(t, strconv.Itoa(int(intVal)), internedInt)
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
