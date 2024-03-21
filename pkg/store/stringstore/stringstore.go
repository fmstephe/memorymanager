package stringstore

import (
	"github.com/fmstephe/flib/funsafe"
	"github.com/fmstephe/location-system/pkg/store/bytestore"
)

type Stats bytestore.Stats

type Reference bytestore.Reference

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
	return Reference(ref), funsafe.BytesToString(bytes)
}

func (s *Store) Get(ref Reference) string {
	bytes := s.store.Get(bytestore.Reference(ref))
	return funsafe.BytesToString(bytes)
}

func (s *Store) Free(ref Reference) {
	s.store.Free(bytestore.Reference(ref))
}

func (s *Store) GetStats() Stats {
	return Stats(s.store.GetStats())
}
