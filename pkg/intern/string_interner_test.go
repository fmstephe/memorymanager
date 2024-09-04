package intern

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringInterner_Interned_EmptySlice(t *testing.T) {
	interner := NewStringInterner(Config{MaxLen: 64, MaxBytes: 1024})

	internedString := interner.Get("")
	assert.Equal(t, "", internedString)

	// a new string value has been interned
	expectedStats := Stats{Returned: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestStringInterner_Interned(t *testing.T) {
	interner := NewStringInterner(Config{MaxLen: 64, MaxBytes: 1024})
	str := "interned string"

	DoTestGenericInterner_Interned(t, interner, str, str)
}

func TestStringInterner_NotInternedMaxLen(t *testing.T) {
	interner := NewStringInterner(Config{MaxLen: 3, MaxBytes: 1024})
	str := "interned string"

	DoTestGenericInterner_NotInternedMaxLen(t, interner, str, str)
}

func TestStringInterner_NotInternedMaxBytes(t *testing.T) {
	interner := NewStringInterner(Config{MaxLen: 64, MaxBytes: 3})
	str := "interned string"

	DoTestGenericInterner_NotInternedMaxBytes(t, interner, str, str)
}

func TestStringInterner_NotInternedHashCollision(t *testing.T) {
	// Right now I don't know of any xxhash collisions When we find a
	// colliding pair of a manageable sized strings we can complete this
	// test
}

// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new strings, then
// returning those strings, then running out of usedString but continuing to
// return previously interned string values.
func TestStringInterner_Complex(t *testing.T) {
	interner := NewStringInterner(Config{MaxLen: 1024, MaxBytes: 1024})
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
			bytes := expectedString
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
			Interned: len(strings),
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these strings again, they are already
	// interned and their interned values are returned to us
	{
		for _, expectedString := range strings {
			bytes := expectedString
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
			Interned: len(strings),
			Returned: len(strings),
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Fill up the rest of the bytes so they are all used up
	{
		usedString := interner.GetStats().UsedBytes
		bytesRemaining := 1024 - usedString

		filler := make([]byte, bytesRemaining-1)
		fillerStr := interner.Get(string(filler))
		assert.Equal(t, string(filler), fillerStr)

		expectedStats := Stats{
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
			bytes := expectedString
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
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
			bytes := expectedString
			internedString := interner.Get(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
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
func TestStringInterner_NoAllocations(t *testing.T) {
	interner := NewStringInterner(Config{MaxLen: 0, MaxBytes: 0})

	strings := make([]string, 10_000)
	for i := range strings {
		strings[i] = strconv.Itoa(i)
	}

	DoTestGenericInterner_NoAllocations(t, interner, strings)
}
