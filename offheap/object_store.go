package offheap

import (
	"github.com/fmstephe/offheap/offheap/internal/pointerstore"
)

const defaultSlabSize = 1 << 13

type Store struct {
	sizedStores []*pointerstore.Store
}

// Returns a new *Store.
//
// This store manages allocation and freeing of any offheap allocated objects.
func New() *Store {
	return &Store{
		sizedStores: initSizeStore(defaultSlabSize),
	}
}

// Returns a new *Store.
//
// The size of each slab, contiguous chunk of memory where allocations are
// organised, is set to be at least slabSize. If slabSize is not a power of
// two, then slabSize will be rounded up to the nearest power of two and then
// used.
//
// Some users may have real need for a Store with a non-standard slab-size. But
// the motivating use of this function was to allow the creation of Stores with
// small slab sizes to allow faster tests with reduced memory usage. Most users
// will probably prefer to use the default New() above.
func NewSized(slabSize int) *Store {
	return &Store{
		sizedStores: initSizeStore(slabSize),
	}
}

func initSizeStore(slabSize int) []*pointerstore.Store {
	slabs := make([]*pointerstore.Store, maxAllocationBits())

	for i := range slabs {
		slabs[i] = pointerstore.New(pointerstore.NewAllocConfigBySize(1<<i, uint64(slabSize)))
	}

	return slabs
}

func (s *Store) alloc(idx int) pointerstore.RefPointer {
	return s.sizedStores[idx].Alloc()
}

func (s *Store) free(idx int, r pointerstore.RefPointer) {
	s.sizedStores[idx].Free(r)
}

// Releases the memory allocated by the Store back to the operating system.
// After this method is called the Store is completely unusable.
//
// There may be some use-cases for this in real systems. But the motivating use
// case for this method was allowing us to release memory of Stores created in
// unit tests (we create a lot of them). Without this method the tests,
// especially the fuzz tests, would OOM very quickly. Right now I would expect
// that most (all?) Stores will live for the entire lifecycle of the program
// they are used in, so this method probably won't be used in most cases.
func (s *Store) Destroy() error {
	for i := range s.sizedStores {
		if err := s.sizedStores[i].Destroy(); err != nil {
			return err
		}
	}

	return nil
}

// Returns the statistics across all allocation size classes for this Store.
//
// There are helper methods which allow the user to easily get the statistics
// for a single size class for object, slices and string allocations.
func (s *Store) Stats() []pointerstore.Stats {
	sizedStats := make([]pointerstore.Stats, len(s.sizedStores))
	for i := range s.sizedStores {
		sizedStats[i] = s.sizedStores[i].Stats()
	}
	return sizedStats
}

// Returns the allocation config across all allocation size classes for this
// Store.
//
// There are helper methods which allow the user to easily get the config for a
// single size class for object, slices and string allocations.
func (s *Store) AllocConfigs() []pointerstore.AllocConfig {
	sizedAllocConfigs := make([]pointerstore.AllocConfig, len(s.sizedStores))
	for i := range s.sizedStores {
		sizedAllocConfigs[i] = s.sizedStores[i].AllocConfig()
	}
	return sizedAllocConfigs
}
