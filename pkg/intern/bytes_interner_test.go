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

func TestBytesInterner_Interned_EmptySlice(t *testing.T) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 64, MaxBytes: 1024})

	internedString := interner.Get([]byte{})
	assert.Equal(t, "", internedString)

	// a new string value has been interned
	expectedStats := internbase.Stats{Returned: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestBytesInterner_Interned(t *testing.T) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 64, MaxBytes: 1024})
	internedBytes := "interned string"
	bytes := []byte(internedBytes)

	DoTestGenericInterner_Interned(t, interner, bytes, internedBytes)
}

func TestBytesInterner_NotInternedMaxLen(t *testing.T) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 3, MaxBytes: 1024})
	internedBytes := "interned string"
	bytes := []byte(internedBytes)

	DoTestGenericInterner_NotInternedMaxLen(t, interner, bytes, internedBytes)
}

func TestBytesInterner_NotInternedMaxBytes(t *testing.T) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 64, MaxBytes: 3})
	internedBytes := "interned string"
	bytes := []byte(internedBytes)

	DoTestGenericInterner_NotInternedMaxBytes(t, interner, bytes, internedBytes)
}

func TestBytesInterner_NotInternedHashCollision(t *testing.T) {
	// Right now I don't know of any xxhash collisions When we find a
	// colliding pair of a manageable sized strings we can complete this
	// test
}

// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new strings, then
// returning those strings, then running out of usedBytes but continuing to
// return previously interned string values.
func TestBytesInterner_Complex(t *testing.T) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 1024, MaxBytes: 1024})
	strings := []string{
		"Heavens!",
		"what",
		"a",
		"virulent",
		"attack!”",
		"replied",
		"the",
		"prince,",
		"not",
		"in",
		"least",
		"disconcerted",
		"by",
		"this",
		"reception.",
		"He",
		"just",
		"entered,",
		"wearing",
		"an",
		"embroidered",
		"court",
		"uniform,",
		"knee",
		"breeches,",
		"and",
		"shoes,",
		"stars",
		"on",
		"his",
		"breast",
		"serene",
		"expression",
		"flat",
		"face.",
		"spoke",
		"refined",
		"French",
		"which",
		"our",
		"grandfathers",
		"only",
		"thought,",
		"with",
		"gentle,",
		"patronizing",
		"intonation",
		"natural",
		"to",
		"man",
		"of",
		"importance",
		"who",
		"had",
		"grown",
		"old",
		"at",
		"went",
		"up",
		"Anna",
		"Pávlovna,",
		"kissed",
		"her",
		"hand,",
		"presenting",
		"bald,",
		"scented,",
		"shining",
		"head,",
		"complacently",
		"seated",
		"himself",
		"sofa",
	}

	// When we intern all these strings, via bytes, each one is unique and
	// is interned
	{
		for _, expectedString := range strings {
			bytes := []byte(expectedString)
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := internbase.Stats{
			Interned: len(strings),
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these strings again, they are already
	// interned and their interned values are returned to us
	{
		for _, expectedString := range strings {
			bytes := []byte(expectedString)
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := internbase.Stats{
			Interned: len(strings),
			Returned: len(strings),
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Fill up the rest of the bytes so they are all used up
	{
		usedBytes := interner.GetStats().UsedBytes
		bytesRemaining := 1024 - usedBytes

		filler := make([]byte, bytesRemaining-1)
		fillerStr := interner.Get(filler)
		assert.Equal(t, string(filler), fillerStr)

		expectedStats := internbase.Stats{
			Interned: len(strings) + 1,
			Returned: len(strings),
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern new strings there aren't enough bytes left
	// to intern any of them
	{
		for _, expectedString := range strings {
			expectedString = expectedString + "_unique"
			bytes := []byte(expectedString)
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := internbase.Stats{
			Interned:          len(strings) + 1,
			Returned:          len(strings),
			UsedBytesExceeded: len(strings),
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Finally we attempt to intern the strings again, they are already
	// interned and their interned values are returned to us
	{
		for _, expectedString := range strings {
			bytes := []byte(expectedString)
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := internbase.Stats{
			Interned:          len(strings) + 1,
			Returned:          len(strings) * 2,
			UsedBytesExceeded: len(strings),
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}
}

// Assert that getting a string, where the value has already been interned,
// does not allocate
func TestBytesInterner_NoAllocations(t *testing.T) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 0, MaxBytes: 0})

	byteVals := make([][]byte, 10_000)
	for i := range byteVals {
		byteVals[i] = []byte(strconv.Itoa(i))
	}

	DoTestGenericInterner_NoAllocations(t, interner, byteVals)
}
