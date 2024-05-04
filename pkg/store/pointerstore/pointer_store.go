package pointerstore

import (
	"sync"
	"sync/atomic"
)

type Stats struct {
	Allocs    int
	Frees     int
	RawAllocs int
	Live      int
	Reused    int
	Slabs     int
}

type Store struct {
	// Immutable fields
	allocConf AllocationConfig

	// Accounting fields
	allocs atomic.Uint64
	frees  atomic.Uint64
	reused atomic.Uint64

	// allIdx provides unique allocation locations for each new allocation
	allocIdx atomic.Uint64

	// freeRWLock protects rootFree
	freeLock sync.Mutex
	rootFree Reference

	// objectsLock protects objects
	// Allocating to an existing slab with a free slot only needs a read lock
	// Adding a new slab to objects requires a write lock
	objectsLock sync.RWMutex
	metadata    [][]uintptr
	objects     [][]uintptr
}

func New(allocConf AllocationConfig) *Store {
	objects := [][]uintptr{}
	return &Store{
		allocConf: allocConf,
		allocIdx:  atomic.Uint64{},
		objects:   objects,
	}
}

func (s *Store) Alloc() Reference {
	s.allocs.Add(1)

	if r, ok := s.allocFromFree(); ok {
		s.reused.Add(1)
		return r
	}

	// allocFromFree failed, fall back to allocating from new slot
	return s.allocFromOffset()
}

func (s *Store) Free(r Reference) {
	s.freeLock.Lock()
	defer s.freeLock.Unlock()

	r.Free(s.rootFree)
	s.rootFree = r

	s.frees.Add(1)
}

func (s *Store) Destroy() error {
	s.objectsLock.Lock()
	defer s.objectsLock.Unlock()

	for _, slab := range s.objects {
		if err := MunmapSlab(slab[0], s.allocConf); err != nil {
			// This is pretty unrecoverable - so we just give up.
			// Maybe we should _try_ to unmap the remaining slabs.
			// I expect that the only useful response to this error
			// is to exit your application, or in the current
			// use-case stop fuzzing.
			return err
		}
	}

	return nil
}

func (s *Store) GetStats() Stats {
	allocs := s.allocs.Load()
	frees := s.frees.Load()
	reused := s.reused.Load()

	// make sure the size of s.objects doesn't change
	s.objectsLock.RLock()
	slabs := len(s.objects)
	s.objectsLock.RUnlock()

	return Stats{
		Allocs:    int(allocs),
		Frees:     int(frees),
		RawAllocs: int(allocs - reused),
		Live:      int(allocs - frees),
		Reused:    int(reused),
		Slabs:     slabs,
	}
}

func (s *Store) GetAllocationConfig() AllocationConfig {
	return s.allocConf
}

func (s *Store) allocFromFree() (Reference, bool) {
	s.freeLock.Lock()
	defer s.freeLock.Unlock()

	// No free objects available - allocFromFree failed
	if s.rootFree.IsNil() {
		return Reference{}, false
	}

	// Get pointer to the next available freed slot
	alloc := s.rootFree
	s.rootFree = alloc.AllocFromFree()

	return alloc, true
}

func (s *Store) allocFromOffset() Reference {
	allocIdx := s.acquireAllocIdx()
	// TODO do some power of 2 work here, to eliminate all this division
	slabIdx := allocIdx / s.allocConf.ObjectsPerSlab
	offsetIdx := allocIdx % s.allocConf.ObjectsPerSlab

	// Take read lock to access s.objects
	s.objectsLock.RLock()
	if slabIdx >= uint64(len(s.objects)) {
		// Release read lock
		s.objectsLock.RUnlock()
		s.growObjects(int(slabIdx + 1))
		// Reacquire read lock
		s.objectsLock.RLock()
	}
	obj := s.objects[slabIdx][offsetIdx]
	meta := s.metadata[slabIdx][offsetIdx]
	// Release read lock
	s.objectsLock.RUnlock()

	ref := NewReference(obj, meta)
	return ref
}

func (s *Store) acquireAllocIdx() uint64 {
	for {
		allocIdx := s.allocIdx.Load()
		if s.allocIdx.CompareAndSwap(allocIdx, allocIdx+1) {
			// Success
			return allocIdx
		}
	}
}

func (s *Store) growObjects(targetLen int) {
	// Acquire write lock to grow the objects slice
	s.objectsLock.Lock()
	for len(s.objects) < targetLen {
		// Create a new slab
		objects, metadata := MmapSlab(s.allocConf)
		s.objects = append(s.objects, objects)
		s.metadata = append(s.metadata, metadata)
	}

	// Release write lock
	s.objectsLock.Unlock()
}
