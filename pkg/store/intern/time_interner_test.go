package intern

import (
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestTimeInterner_Interned(t *testing.T) {
	interner := NewTimeInterner(64, 1024, time.RFC1123)

	// A string is returned with the same value as timestamp
	timestamp := time.Now()
	internedTime := interner.Get(timestamp)
	assert.Equal(t, timestamp.Format(time.RFC1123), internedTime)

	// a new int value has been interned
	expectedStats := Stats{Interned: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as timestamp
	internedTime2 := interner.Get(timestamp)
	assert.Equal(t, timestamp.Format(time.RFC1123), internedTime2)
	// The string returned uses the same memory allocation as the first
	// value returned i.e. the string is interned as is being reused as
	// intended
	assert.Equal(t, unsafe.StringData(internedTime), unsafe.StringData(internedTime2))

	// An interned string has been returned
	expectedStats = Stats{Interned: 1, Returned: 1}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestTimeInterner_NotInternedMaxLen(t *testing.T) {
	interner := NewTimeInterner(3, 1024, time.RFC1123)

	// A string is returned with the same value as timestamp
	timestamp := time.Now()
	notInternedInt := interner.Get(timestamp)
	assert.Equal(t, timestamp.Format(time.RFC1123), notInternedInt)

	// The int passed in was too long, so maxLenExceeded should be recorded
	expectedStats := Stats{MaxLenExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as timestamp
	notInternedInt2 := interner.Get(timestamp)
	assert.Equal(t, timestamp.Format(time.RFC1123), notInternedInt2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedInt), unsafe.StringData(notInternedInt2))

	// The int passed in was too long, so maxLenExceeded should be recorded
	expectedStats = Stats{MaxLenExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

func TestTimeInterner_NotInternedUsedInt(t *testing.T) {
	interner := NewTimeInterner(64, 3, time.RFC1123)

	// A string is returned with the same value as timestamp
	timestamp := time.Now()
	notInternedInt := interner.Get(timestamp)
	assert.Equal(t, timestamp.Format(time.RFC1123), notInternedInt)

	// The int passed in was too long, so usedBytesExceeded should be recorded
	expectedStats := Stats{UsedBytesExceeded: 1}
	stats := interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)

	// A string is returned with the same value as timestamp
	notInternedInt2 := interner.Get(timestamp)
	assert.Equal(t, timestamp.Format(time.RFC1123), notInternedInt2)
	// The string returned uses a different memory allocation from the
	// first value returned i.e. the strings were not interned, and a new
	// string is being allocated each time
	assert.NotSame(t, unsafe.StringData(notInternedInt), unsafe.StringData(notInternedInt2))

	// The int passed in was too long, so usedBytesExceeded should be recorded
	expectedStats = Stats{UsedBytesExceeded: 2}
	stats = interner.GetStats()
	assert.Equal(t, expectedStats, stats.Total)
}

// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new timestamps, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned timestamp values.
func TestTimeInterner_Complex(t *testing.T) {
	interner := NewTimeInterner(1024, 1024*1024, time.RFC1123)
	numberOfTimestamps := 100

	timestamps := make([]time.Time, 0, numberOfTimestamps)
	now := time.Now()
	for i := range numberOfTimestamps {
		timestamps = append(timestamps, now.Add(time.Millisecond*time.Duration(i)))
	}

	// When we intern all these ints, each one is unique and is interned
	{
		for _, timestamp := range timestamps {
			internedTime := interner.Get(timestamp)
			assert.Equal(t, timestamp.Format(time.RFC1123), internedTime)
		}

		expectedStats := Stats{
			Interned: numberOfTimestamps,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// When we attempt to intern these ints again, they are already
	// interned and their interned values are returned to us
	{
		for _, timestamp := range timestamps {
			internedTime := interner.Get(timestamp)
			assert.Equal(t, timestamp.Format(time.RFC1123), internedTime)
		}

		expectedStats := Stats{
			Interned: numberOfTimestamps,
			Returned: numberOfTimestamps,
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
		for _, timestamp := range timestamps {
			// push the timestamp forward 1 day to make them new,
			// uninterned, timestamps
			timestamp = timestamp.Add(24 * time.Hour)
			internedTime := interner.Get(timestamp)
			assert.Equal(t, timestamp.Format(time.RFC1123), internedTime)
		}

		expectedStats := Stats{
			Interned:          numberOfTimestamps,
			Returned:          numberOfTimestamps,
			UsedBytesExceeded: numberOfTimestamps,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}

	// Finally we attempt to intern the ints again, they are already
	// interned and their interned values are returned to us
	{
		for _, timestamp := range timestamps {
			internedTime := interner.Get(timestamp)
			assert.Equal(t, timestamp.Format(time.RFC1123), internedTime)
		}

		expectedStats := Stats{
			Interned:          numberOfTimestamps,
			Returned:          numberOfTimestamps * 2,
			UsedBytesExceeded: numberOfTimestamps,
		}
		stats := interner.GetStats()
		assert.Equal(t, expectedStats, stats.Total)
	}
}

// This benchmark is intended to demonstrate that getting string values for
// time.Time values that have already been interned does not allocate
func BenchmarkTimeInterner(b *testing.B) {
	interner := NewTimeInterner(0, 0, time.RFC1123)

	now := time.Now()
	timestamps := make([]time.Time, 10_000)
	for i := range timestamps {
		timestamps[i] = now.Add(time.Duration(i))
	}

	for _, timestamp := range timestamps {
		interner.Get(timestamp)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	for {
		for _, timestamp := range timestamps {
			interner.Get(timestamp)
			count++
			if count >= b.N {
				return
			}
		}
	}
}
