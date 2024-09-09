// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package intern

import (
	"testing"
	"time"
)

func TestTimeInterner_Interned(t *testing.T) {
	interner := NewTimeInterner(Config{MaxLen: 64, MaxBytes: 1024}, time.RFC1123)
	timestamp := time.Now()
	internedTimestamp := timestamp.Format(time.RFC1123)

	DoTestGenericInterner_Interned(t, interner, timestamp, internedTimestamp)
}

func TestTimeInterner_NotInternedMaxLen(t *testing.T) {
	interner := NewTimeInterner(Config{MaxLen: 3, MaxBytes: 1024}, time.RFC1123)
	timestamp := time.Now()
	internedTimestamp := timestamp.Format(time.RFC1123)

	DoTestGenericInterner_NotInternedMaxLen(t, interner, timestamp, internedTimestamp)
}

func TestTimeInterner_NotInternedMaxBytes(t *testing.T) {
	interner := NewTimeInterner(Config{MaxLen: 64, MaxBytes: 3}, time.RFC1123)
	timestamp := time.Now()
	internedTimestamp := timestamp.Format(time.RFC1123)

	DoTestGenericInterner_NotInternedMaxBytes(t, interner, timestamp, internedTimestamp)
}

/*
// This test demonstrates that the interner can handle passing through a
// variety of states successfully. Specifically interning new timestamps, then
// returning those as strings, then running out of usedBytes but continuing to
// return previously interned timestamp values.
func TestTimeInterner_Complex(t *testing.T) {
	interner := NewTimeInterner(Config{MaxLen: 1024, MaxBytes: 1024 * 1024}, time.RFC1123)
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

	// When we attempt to intern new timestamps there aren't enough bytes left
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
*/

// Assert that getting a string, where the value has already been interned,
// does not allocate
func TestTimeInterner_NoAllocations(t *testing.T) {
	interner := NewTimeInterner(Config{MaxLen: 0, MaxBytes: 0}, time.RFC1123)

	timestamps := make([]time.Time, 10_000)
	now := time.Now()
	for i := range timestamps {
		timestamps[i] = now.Add(time.Nanosecond)
	}

	DoTestGenericInterner_NoAllocations(t, interner, timestamps)
}
