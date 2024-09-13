// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"strconv"
	"testing"

	"github.com/fmstephe/memorymanager/pkg/intern/internbase"
)

func BenchmarkBytesInterner_NoneInterned(b *testing.B) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 0, MaxBytes: 0})

	bytes := make([][]byte, b.N)
	for i := range bytes {
		bytes[i] = []byte(strconv.Itoa(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for _, bytesVal := range bytes {
		interner.Get(bytesVal)
	}
}

func BenchmarkBytesInterner_AllInterned(b *testing.B) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 0, MaxBytes: 0})

	bytes := make([][]byte, b.N)
	for i := range bytes {
		bytes[i] = []byte(strconv.Itoa(i))
	}

	for _, bytesVal := range bytes {
		interner.Get(bytesVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for _, bytesVal := range bytes {
		interner.Get(bytesVal)
	}
}

// Benchmark getting already interned values, but limit the size of the set of interned values.
//
// This simulates the behaviour when the interner is used on a smallish fixed set of common values.
func BenchmarkBytesInterner_AllInterned10K(b *testing.B) {
	interner := NewBytesInterner(internbase.Config{MaxLen: 0, MaxBytes: 0})

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
