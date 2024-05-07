package objectstore

import (
	"fmt"
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/pointerstore"
)

func Alloc[T any](s *Store) (Reference[T], *T) {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	idx := indexForType[T]()
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", sizeForType[T]()))
	}

	pRef := s.alloc(idx)
	oRef := newReference[T](pRef)
	return oRef, oRef.GetValue()
}

func Free[T any](s *Store, r Reference[T]) {
	idx := indexForType[T]()
	s.free(idx, r.ref)
}

// The address field holds a pointer to an object[O], but also sneaks a
// generation value in the top 8 bits of the address field.
//
// The generation must be masked out to get a usable pointer value. The object
// pointed to must have the same generation value in order to access/free that
// object.
type Reference[O any] struct {
	ref pointerstore.Reference
}

func newReference[O any](ref pointerstore.Reference) Reference[O] {
	if ref.IsNil() {
		panic("cannot create new Reference with nil pointerstore.Reference")
	}

	return Reference[O]{
		ref: ref,
	}
}

func (r *Reference[O]) GetValue() *O {
	return (*O)((unsafe.Pointer)(r.ref.GetDataPtr()))
}

func (r *Reference[O]) getGen() uint8 {
	return r.ref.GetGen()
}

func (r *Reference[O]) IsNil() bool {
	return r.ref.IsNil()
}
