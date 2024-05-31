package objectstore

import (
	"fmt"
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

func AllocObject[T any](s *Store) RefObject[T] {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	idx := indexForType[T]()

	pRef := s.alloc(idx)
	oRef := newRefObject[T](pRef)
	return oRef
}

func FreeObject[T any](s *Store, r RefObject[T]) {
	idx := indexForType[T]()
	s.free(idx, r.ref)
}

// A reference to a typed object
type RefObject[T any] struct {
	ref pointerstore.RefPointer
}

func newRefObject[T any](ref pointerstore.RefPointer) RefObject[T] {
	if ref.IsNil() {
		panic("cannot create new Reference with nil pointerstore.Reference")
	}

	return RefObject[T]{
		ref: ref,
	}
}

func (r *RefObject[T]) Value() *T {
	return (*T)((unsafe.Pointer)(r.ref.DataPtr()))
}

func (r *RefObject[T]) IsNil() bool {
	return r.ref.IsNil()
}
