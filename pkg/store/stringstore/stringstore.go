package stringstore

import (
	"github.com/fmstephe/location-system/pkg/store/bytestore"
	"github.com/fmstephe/unsafeutil"
)

type Stats bytestore.Stats

type Pointer bytestore.Pointer

type Store struct {
	store *bytestore.Store
}

func New() *Store {
	return &Store{
		store: bytestore.New(),
	}
}

func (s *Store) Alloc(str string) (Pointer, string) {
	size := len(str)
	p, bytes := s.store.Alloc(uint32(size))
	copy(bytes, str)
	return Pointer(p), unsafeutil.BytesToString(bytes)
}

func (s *Store) Get(p Pointer) string {
	bytes := s.store.Get(bytestore.Pointer(p))
	return unsafeutil.BytesToString(bytes)
}

func (s *Store) Free(p Pointer) {
	s.store.Free(bytestore.Pointer(p))
}

func (s *Store) GetStats() Stats {
	return Stats(s.store.GetStats())
}
