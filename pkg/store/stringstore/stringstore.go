package stringstore

import (
	"github.com/fmstephe/location-system/pkg/store/bytestore"
	"github.com/fmstephe/unsafeutil"
)

type Stats bytestore.Stats

type Reference bytestore.Pointer

type Store struct {
	store *bytestore.Store
}

func New() *Store {
	return &Store{
		store: bytestore.New(),
	}
}

func (s *Store) Alloc(str string) (Reference, string) {
	size := len(str)
	ref, bytes := s.store.Alloc(uint32(size))
	copy(bytes, str)
	return Reference(ref), unsafeutil.BytesToString(bytes)
}

func (s *Store) Get(ref Reference) string {
	bytes := s.store.Get(bytestore.Pointer(ref))
	return unsafeutil.BytesToString(bytes)
}

func (s *Store) Free(ref Reference) {
	s.store.Free(bytestore.Pointer(ref))
}

func (s *Store) GetStats() Stats {
	return Stats(s.store.GetStats())
}
