package intern

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestInternBytes_Interned(t *testing.T) {
	interner := New(64, 1024)

	// A string value is returned with the same value as expectedString
	expectedString := "interned string"
	internedString := interner.GetFromBytes([]byte(expectedString))
	assert.Equal(t, expectedString, internedString)

	// a new string value has been interned
	expectedStats := Stats{interned: 1}
	stats := interner.GetBytesStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string value is returned with the same value as expectedString
	internedString2 := interner.GetFromBytes([]byte(expectedString))
	assert.Equal(t, expectedString, internedString2)
	// The string returned uses the same memory allocation as the first
	// value returned i.e. the string is interned as is being reused as
	// intended
	assert.Equal(t, unsafe.StringData(internedString), unsafe.StringData(internedString2))

	// An interned string value has been returned
	expectedStats = Stats{interned: 1, returned: 1}
	stats = interner.GetBytesStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternBytes_NotInternedMaxLen(t *testing.T) {
	interner := New(3, 1024)

	// A string is returned with the same value as expectedString
	expectedString := "interned string"
	notInternedString := interner.GetFromBytes([]byte(expectedString))
	assert.Equal(t, expectedString, notInternedString)

	// The bytes passed in was too long, so maxLenExceeded should be recorded
	expectedStats := Stats{maxLenExceeded: 1}
	stats := interner.GetBytesStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as expectedString
	notInternedString2 := interner.GetFromBytes([]byte(expectedString))
	assert.Equal(t, expectedString, notInternedString2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedString), unsafe.StringData(notInternedString2))

	// The bytes passed in was too long, so maxLenExceeded should be recorded
	expectedStats = Stats{maxLenExceeded: 2}
	stats = interner.GetBytesStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternBytes_NotInternedUsedBytes(t *testing.T) {
	interner := New(64, 3)

	expectedString := "interned string"
	internedString := interner.GetFromBytes([]byte(expectedString))

	expectedStats := Stats{usedBytesExceeded: 1}
	stats := interner.GetBytesStats()

	assert.Equal(t, expectedString, internedString)
	assert.Equal(t, expectedStats, stats.Total)
}

func TestInternBytes_NotInternedHashCollision(t *testing.T) {
	// Right now I don't know of any xxhash collisions When we find a
	// colliding pair of a manageable sized strings we can complete this
	// test
}

// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new strings, then
// returning those strings, then running out of usedBytes but continuing to
// return previously interned string values.
func TestInternBytes_Complex(t *testing.T) {
	interner := New(1024, 1024)
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
			internedString := interner.GetFromBytes(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
			interned: len(strings),
		}
		stats := interner.GetBytesStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these strings again, they are already
	// interned and their interned values are returned to us
	{
		for _, expectedString := range strings {
			bytes := []byte(expectedString)
			internedString := interner.GetFromBytes(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
			interned: len(strings),
			returned: len(strings),
		}
		stats := interner.GetBytesStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Fill up the rest of the bytes so they are all used up
	{
		usedBytes := interner.GetBytesStats().UsedBytes
		bytesRemaining := 1024 - usedBytes

		filler := make([]byte, bytesRemaining-1)
		fillerStr := interner.GetFromBytes(filler)
		assert.Equal(t, string(filler), fillerStr)

		expectedStats := Stats{
			interned: len(strings) + 1,
			returned: len(strings),
		}
		stats := interner.GetBytesStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern new strings there aren't enough bytes left
	// to intern any of them
	{
		for _, expectedString := range strings {
			expectedString = expectedString + "_unique"
			bytes := []byte(expectedString)
			internedString := interner.GetFromBytes(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
			interned:          len(strings) + 1,
			returned:          len(strings),
			usedBytesExceeded: len(strings),
		}
		stats := interner.GetBytesStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Finally we attempt to intern the strings again, they are already
	// interned and their interned values are returned to us
	{
		for _, expectedString := range strings {
			bytes := []byte(expectedString)
			internedString := interner.GetFromBytes(bytes)
			assert.Equal(t, expectedString, internedString)
		}

		expectedStats := Stats{
			interned:          len(strings) + 1,
			returned:          len(strings) * 2,
			usedBytesExceeded: len(strings),
		}
		stats := interner.GetBytesStats()
		assert.Equal(t, expectedStats, stats.Total)
	}
}
