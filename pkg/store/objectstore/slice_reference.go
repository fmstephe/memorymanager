package objectstore

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

func AllocSlice[T any](s *Store, length, capacity int) (RefSlice[T], []T) {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	idx := indexForSlice[T](capacity)
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", sizeForType[T]()))
	}

	pRef := s.alloc(idx)
	sRef := newRefSlice[T](length, capacity, pRef)
	return sRef, sRef.Value()
}

func FreeSlice[T any](s *Store, r RefSlice[T]) {
	idx := indexForSlice[T](r.capacity)
	s.free(idx, r.ref)
}

func indexForSlice[T any](capacity int) int {
	typeSize := uint64(fmath.NxtPowerOfTwo(int64(sizeForType[T]())))
	sliceSize := uint64(capacity) * typeSize
	return indexForSize(sliceSize)
}

// A reference to a typed slice
// length is len() of the slice
// capacity is cap() of the slice
type RefSlice[T any] struct {
	length   int
	capacity int
	ref      pointerstore.Reference
}

func newRefSlice[T any](length, capacity int, ref pointerstore.Reference) RefSlice[T] {
	if ref.IsNil() {
		panic("cannot create new RefSlice with nil pointerstore.RefSlice")
	}

	return RefSlice[T]{
		length:   length,
		capacity: capacity,
		ref:      ref,
	}
}

func (r *RefSlice[T]) Value() (slice []T) {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Data = r.ref.DataPtr()
	sliceHeader.Len = r.length
	sliceHeader.Cap = r.capacity
	return slice
}

func (r *RefSlice[T]) IsNil() bool {
	return r.ref.IsNil()
}
