package intern

import (
	"testing"
)

func BenchmarkFloat64Interner_NoneInterned(b *testing.B) {
	interner := NewFloat64Interner(Config{MaxLen: 0, MaxBytes: 0}, 'f', -1, 64)

	floats := make([]float64, b.N)
	for i := range floats {
		floats[i] = float64(i) + 0.123
	}

	b.ReportAllocs()
	b.ResetTimer()
	for _, floatVal := range floats {
		interner.Get(floatVal)
	}
}

func BenchmarkFloat64Interner_AllInterned(b *testing.B) {
	interner := NewFloat64Interner(Config{MaxLen: 0, MaxBytes: 0}, 'f', -1, 64)

	floats := make([]float64, b.N)
	for i := range floats {
		floats[i] = float64(i) + 0.123
	}

	for _, floatVal := range floats {
		interner.Get(floatVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for _, floatVal := range floats {
		interner.Get(floatVal)
	}
}

// Benchmark getting already interned values, but limit the size of the set of interned values.
//
// This simulates the behaviour when the interner is used on a smallish fixed set of common values.
func BenchmarkFloat64Interner_AllInterned10K(b *testing.B) {
	interner := NewFloat64Interner(Config{MaxLen: 0, MaxBytes: 0}, 'f', -1, 64)

	floats := make([]float64, 10_000)
	for i := range floats {
		floats[i] = float64(i) + 0.123
	}

	for _, floatVal := range floats {
		interner.Get(floatVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	for {
		for _, floatVal := range floats {
			interner.Get(floatVal)
			count++
			if count >= b.N {
				return
			}
		}
	}
}
