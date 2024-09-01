package intern

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestBytesInterner_Interned_EmptySlice(t *testing.T) {
	interner := NewBytesInterner(64, 1024)

	internedString := interner.Get([]byte{})
	assert.Equal(t, "", internedString)

	// a new string value has been interned
	expectedStats := Stats{Returned: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestBytesInterner_Interned(t *testing.T) {
	interner := NewBytesInterner(64, 1024)

	// A string value is returned with the same value as expectedString
	expectedString := "interned string"
	internedString := interner.Get([]byte(expectedString))
	assert.Equal(t, expectedString, internedString)

	// a new string value has been interned
	expectedStats := Stats{Interned: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string value is returned with the same value as expectedString
	internedString2 := interner.Get([]byte(expectedString))
	assert.Equal(t, expectedString, internedString2)
	// The string returned uses the same memory allocation as the first
	// value returned i.e. the string is interned as is being reused as
	// intended
	assert.Equal(t, unsafe.StringData(internedString), unsafe.StringData(internedString2))

	// An interned string value has been returned
	expectedStats = Stats{Interned: 1, Returned: 1}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestBytesInterner_NotInternedMaxLen(t *testing.T) {
	interner := NewBytesInterner(3, 1024)

	// A string is returned with the same value as expectedString
	expectedString := "interned string"
	notInternedString := interner.Get([]byte(expectedString))
	assert.Equal(t, expectedString, notInternedString)

	// The bytes passed in was too long, so maxLenExceeded should be recorded
	expectedStats := Stats{MaxLenExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as expectedString
	notInternedString2 := interner.Get([]byte(expectedString))
	assert.Equal(t, expectedString, notInternedString2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedString), unsafe.StringData(notInternedString2))

	// The bytes passed in was too long, so maxLenExceeded should be recorded
	expectedStats = Stats{MaxLenExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestBytesInterner_NotInternedUsedBytes(t *testing.T) {
	interner := NewBytesInterner(64, 3)

	// A string is returned with the same value as expectedString
	expectedString := "interned string"
	notInternedString := interner.Get([]byte(expectedString))
	assert.Equal(t, expectedString, notInternedString)

	// The bytes passed in was too long, so usedBytesExceeded should be recorded
	expectedStats := Stats{UsedBytesExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as expectedString
	notInternedString2 := interner.Get([]byte(expectedString))
	assert.Equal(t, expectedString, notInternedString2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedString), unsafe.StringData(notInternedString2))

	// The bytes passed in was too long, so usedBytesExceeded should be recorded
	expectedStats = Stats{UsedBytesExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
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
	interner := NewBytesInterner(1024, 1024)
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
			bytes := []byte(expectedString)
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
		usedBytes := interner.GetStats().UsedBytes
		bytesRemaining := 1024 - usedBytes

		filler := make([]byte, bytesRemaining-1)
		fillerStr := interner.Get(filler)
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
			bytes := []byte(expectedString)
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
			bytes := []byte(expectedString)
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

// This benchmark is intended to demonstrate that getting string values for
// []byte that have already been interned does not allocate
func BenchmarkBytesInterner(b *testing.B) {
	interner := NewBytesInterner(0, 0)

	bytes := make([][]byte, 10_000)
	for i := range bytes {
		bytes[i] = []byte(strconv.Itoa(i))
	}

	for _, bytesVal := range bytes {
		interner.Get(bytesVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	for {
		for _, bytesVal := range bytes {
			interner.Get(bytesVal)
			count++
			if count >= b.N {
				return
			}
		}
	}
}
