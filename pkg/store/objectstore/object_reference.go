package objectstore

import (
	"fmt"
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

func AllocObject[T any](s *Store) (RefObject[T], *T) {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	idx := indexForType[T]()
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", sizeForType[T]()))
	}

	pRef := s.alloc(idx)
	oRef := newRefObject[T](pRef)
	return oRef, oRef.GetValue()
}

func FreeObject[T any](s *Store, r RefObject[T]) {
	idx := indexForType[T]()
	s.free(idx, r.ref)
}

// A reference to a typed object
type RefObject[T any] struct {
	ref pointerstore.Reference
}

func newRefObject[T any](ref pointerstore.Reference) RefObject[T] {
	if ref.IsNil() {
		panic("cannot create new Reference with nil pointerstore.Reference")
	}

	return RefObject[T]{
		ref: ref,
	}
}

func (r *RefObject[T]) GetValue() *T {
	return (*T)((unsafe.Pointer)(r.ref.GetDataPtr()))
}

func (r *RefObject[T]) getGen() uint8 {
	return r.ref.GetGen()
}

func (r *RefObject[T]) IsNil() bool {
	return r.ref.IsNil()
}
