package offheap

import "testing"

type benchStruct struct {
	field int
}

func BenchmarkAllocWriteRead(b *testing.B) {
	os := New()

	refs := make([]RefObject[benchStruct], 0, b.N)

	for range b.N {
		ref := os.AllocObject[benchStruct]()
		refs = append(refs, ref)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := range b.N {
		refs[i].Value().field = i
	}

	sum := 0

	for i := range b.N {
		sum += refs[i].Value().field
	}

	println(sum)
}
