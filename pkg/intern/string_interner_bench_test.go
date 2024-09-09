// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package intern

import (
	"strconv"
	"testing"
)

func BenchmarkStringInterner_NoneInterned(b *testing.B) {
	interner := NewStringInterner(Config{MaxLen: 0, MaxBytes: 0})

	strings := make([]string, b.N)
	for i := range strings {
		strings[i] = strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for _, stringVal := range strings {
		interner.Get(stringVal)
	}
}

func BenchmarkStringInterner_AllInterned(b *testing.B) {
	interner := NewStringInterner(Config{MaxLen: 0, MaxBytes: 0})

	strings := make([]string, b.N)
	for i := range strings {
		strings[i] = strconv.Itoa(i)
	}

	for _, stringVal := range strings {
		interner.Get(stringVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for _, stringVal := range strings {
		interner.Get(stringVal)
	}
}

// Benchmark getting already interned values, but limit the size of the set of interned values.
//
// This simulates the behaviour when the interner is used on a smallish fixed set of common values.
func BenchmarkStringInterner_AllInterned10K(b *testing.B) {
	interner := NewStringInterner(Config{MaxLen: 0, MaxBytes: 0})

	strings := make([]string, 10_000)
	for i := range strings {
		strings[i] = strconv.Itoa(i)
	}

	for _, stringVal := range strings {
		interner.Get(stringVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	for {
		for _, stringVal := range strings {
			interner.Get(stringVal)
			count++
			if count >= b.N {
				return
			}
		}
	}
}
