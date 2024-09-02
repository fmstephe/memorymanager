package intern

import (
	"testing"
)

func BenchmarkInt64Interner_NoneInterned(b *testing.B) {
	interner := NewInt64Interner(Config{MaxLen: 0, MaxBytes: 0}, 10)

	ints := make([]int64, b.N)
	for i := range ints {
		ints[i] = int64(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for _, intVal := range ints {
		interner.Get(intVal)
	}
}

func BenchmarkInt64Interner_AllInterned(b *testing.B) {
	interner := NewInt64Interner(Config{MaxLen: 0, MaxBytes: 0}, 10)

	ints := make([]int64, b.N)
	for i := range ints {
		ints[i] = int64(i)
	}

	for _, intVal := range ints {
		interner.Get(intVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for _, intVal := range ints {
		interner.Get(intVal)
	}
}

// Benchmark getting already interned values, but limit the size of the set of interned values.
//
// This simulates the behaviour when the interner is used on a smallish fixed set of common values.
func BenchmarkInt64Interner_AllInterned10K(b *testing.B) {
	interner := NewInt64Interner(Config{MaxLen: 0, MaxBytes: 0}, 10)

	ints := make([]int64, 10_000)
	for i := range ints {
		ints[i] = int64(i)
	}

	for _, intVal := range ints {
		interner.Get(intVal)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	for {
		for _, intVal := range ints {
			interner.Get(intVal)
			count++
			if count >= b.N {
				return
			}
		}
	}
}
