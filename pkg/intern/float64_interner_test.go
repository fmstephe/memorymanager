// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"strconv"
	"testing"

	"github.com/fmstephe/memorymanager/pkg/intern/internbase"
	"github.com/stretchr/testify/assert"
)

func TestFloat64Interner_Interned(t *testing.T) {
	interner := NewFloat64Interner(internbase.Config{MaxLen: 64, MaxBytes: 1024}, 'f', -1, 64)
	floatVal := float64(12.34)
	internedFloat := strconv.FormatFloat(floatVal, 'f', -1, 64)

	DoTestGenericInterner_Interned(t, interner, floatVal, internedFloat)
}

func TestFloat64Interner_NotInternedMaxLen(t *testing.T) {
	interner := NewFloat64Interner(internbase.Config{MaxLen: 3, MaxBytes: 1024}, 'f', -1, 64)
	floatVal := float64(12.34)
	internedFloat := strconv.FormatFloat(floatVal, 'f', -1, 64)

	DoTestGenericInterner_NotInternedMaxLen(t, interner, floatVal, internedFloat)
}

func TestFloat64Interner_NotInternedMaxBytes(t *testing.T) {
	interner := NewFloat64Interner(internbase.Config{MaxLen: 64, MaxBytes: 3}, 'f', -1, 64)
	floatVal := float64(12.34)
	internedFloat := strconv.FormatFloat(floatVal, 'f', -1, 64)

	DoTestGenericInterner_NotInternedMaxBytes(t, interner, floatVal, internedFloat)
}

/*
// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new floats, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned float values.
func TestFloat64Interner_Complex(t *testing.T) {
	interner := NewFloat64Interner(internbase.Config{MaxLen: 1024, MaxBytes: 1024}, 'f', -1, 64)
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
	interner := NewFloat64Interner(internbase.Config{MaxLen: 0, MaxBytes: 0}, 'f', -1, 64)

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
