package objectstore

import (
	"fmt"
	"reflect"

	"github.com/fmstephe/location-system/pkg/store/pointerstore"
)

type Store[O any] struct {
	store *pointerstore.Store
}

func New[O any](objectsPerSlab uint64) *Store[O] {
	if err := containsNoPointers[O](); err != nil {
		panic(fmt.Errorf("cannot instantiate Store with generic type containing pointers %w", err))
	}

	t := reflect.TypeFor[O]()
	objectSize := uint64(t.Size())
	pStore := pointerstore.New(objectSize, objectsPerSlab)
	return &Store[O]{
		store: pStore,
	}
}

func (s *Store[O]) Alloc() (Reference[O], *O) {
	pRef := s.store.Alloc()
	oRef := newReference[O](pRef)
	return oRef, oRef.GetValue()
}

func (s *Store[O]) Free(r Reference[O]) {
	s.store.Free(r.ref)
}

func (s *Store[O]) GetStats() pointerstore.Stats {
	return s.store.GetStats()
}
