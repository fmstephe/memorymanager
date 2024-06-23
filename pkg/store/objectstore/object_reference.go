package objectstore

import (
	"fmt"
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

// Allocates an object of type T. The type T must not contain any pointers in
// any part of its type. If the type T is found to contain pointers this
// function will panic.
//
// The values of fields in the newly allocated object will be arbitrary. Unlike
// Go allocations objects acquired via AllocObject do _not_ have their contents
// zeroed out.
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

// Frees the allocation referenced by r. After this call returns r must never
// be used again. Any use of the object referenced by r will have
// unpredicatable behaviour.
func FreeObject[T any](s *Store, r RefObject[T]) {
	idx := indexForType[T]()
	s.free(idx, r.ref)
}

// A reference to a typed object. This reference allows us to gain access to an
// allocated object directly.
//
// It is acceptable, and enouraged, to use RefObject in fields of types which
// will be managed by a Store. This is acceptable because RefObject does not
// contain any conventional Go pointers.
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

// Returns a pointer to the raw object pointed to by this RefOject.
//
// Care must be taken not to use this object after FreeObject(...) has been
// called on this RefObject.
func (r *RefObject[T]) Value() *T {
	return (*T)((unsafe.Pointer)(r.ref.DataPtr()))
}

// Returns true if this RefObject does not point to an allocated object, false otherwise.
func (r *RefObject[T]) IsNil() bool {
	return r.ref.IsNil()
}

// Returns the stats for the allocation size of type T.
//
// It is important to note that these statistics apply to the size class
// indicated by T. The statistics allocations will capture all allocations for
// this _size_ including allocations for types other than T.
func StatsForType[T any](s *Store) pointerstore.Stats {
	stats := s.Stats()
	idx := indexForType[T]()
	return stats[idx]
}

// Returns the allocation config for the allocation size of type T
//
// It is important to note that this config apply to the size class indicated
// by T. The config apply to all allocations for this _size_ including
// allocations for types other than T.
func ConfForType[T any](s *Store) pointerstore.AllocConfig {
	configs := s.AllocConfigs()
	idx := indexForType[T]()
	return configs[idx]
}
