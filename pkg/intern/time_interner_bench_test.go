// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"testing"
	"time"
)

func BenchmarkTimeInterner_NoneInterned(b *testing.B) {
	interner := NewTimeInterner(Config{MaxLen: 0, MaxBytes: 0}, time.RFC1123)

	timestamps := make([]time.Time, b.N)
	now := time.Now()
	for i := range timestamps {
		timestamps[i] = now.Add(time.Nanosecond)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for _, timestamp := range timestamps {
		interner.Get(timestamp)
	}
}

func BenchmarkTimeInterner_AllInterned(b *testing.B) {
	interner := NewTimeInterner(Config{MaxLen: 0, MaxBytes: 0}, time.RFC1123)

	timestamps := make([]time.Time, b.N)
	now := time.Now()
	for i := range timestamps {
		timestamps[i] = now.Add(time.Nanosecond)
	}

	for _, timestamp := range timestamps {
		interner.Get(timestamp)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for _, timestamp := range timestamps {
		interner.Get(timestamp)
	}
}

// Benchmark getting already interned values, but limit the size of the set of interned values.
//
// This simulates the behaviour when the interner is used on a smallish fixed set of common values.
func BenchmarkTimeInterner_AllInterned10K(b *testing.B) {
	interner := NewTimeInterner(Config{MaxLen: 0, MaxBytes: 0}, time.RFC1123)

	timestamps := make([]time.Time, 10_000)
	now := time.Now()
	for i := range timestamps {
		timestamps[i] = now.Add(time.Nanosecond)
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
