// The objectstore package allows us to alloc and free objects of a specific
// type.
//
// Each Store instance can alloc/free a single type of object. This is
// controlled by the generic type of an objectstore instance e.g.
//
//	var store *objectstore.Store
//	store = objectstore.New[int]()
//
// will alloc/free only int values.
//
// Each allocated object has a corresponding Reference which acts like a
// conventional pointer to retrieve the actual object from the Store using the
// Get() method e.g.
//
//	store := objectstore.New[int]()
//	var reference objectstore.Reference
//	var i1 *int
//	reference, i1 = store.Alloc()
//	var i2 *int
//	i2 = store.Get(reference)
//	if i1 == i2 {
//	  println("This is correct, i1 and i2 will be a pointer to the same int")
//	}
//
// When you know that an allocated object will never be used again it's memory
// can be released back to the Store using Free() e.g.
//
//	store := objectstore.New[int]()
//	reference, i1 := store.Alloc()
//	println(*i1)
//	store.Free(reference)
//	// You must never user i1 or reference again
//
// A best effort has been made to panic if an object is freed twice or if a
// freed object is accessed using Get(). However, it isn't guaranteed that
// these calls will panic. For example if an object is freed the next call to
// Alloc() will reuse the freed object and future calls to Get() and Free()
// using the Reference used in the Free will operate on the re-allocated object
// and not panic. So this behaviour cannot be relied on.
//
// References can be kept and stored in arbitrary datastructures, which can
// themselves be managed by a Store e.g.
//
//	type Node struct {
//	  left  Reference[Node]
//	  right Reference[Node]
//	}
//	node := objectstore.New[Node]()
//
// The Reference type contains no pointers. This means that we can retain as
// many of them as we like with no garbage collection cost. The Store itself
// contains some pointers, and must also be referenced via a pointer to work
// properly. So storing the Store itself in the Node struct above would
// introduce pointers into the managed object type and defeat the purpose of
// using a Store.
//
// The advantage of using a Store, over just allocating objects the normal
// way, is that we get pointer-like references to objects without those
// references being visible to the garbage collector. This means that we can
// build large datastructures, like trees or linked lists, which can live in
// memory indefinitely but incur almost zero garbage collection cost.
//
// The disadvantage of using a Store, over just allocating objects the normal
// way, is that we don't get to enjoy the benefits of the Go garbage collector.
// We must be responsible for freeing unneeded objects manually. We also need
// to have a Store instance to be able to retrieve objects with a Reference, so
// using datastructures built with a Store can feel quite cumbersome. These
// disadvantages mean that effective use of the objectstore requires a careful
// design which encapsulates these details and hides the use of the objectstore
// package from the rest of the program.
//
// It is important to note that the objects managed by a Store do exist on the
// heap. They live inside a series of slices held internally by the Store. This
// means that if the objects managed by a Store contain any conventional Go
// pointers the entire Store will be filled with pointers and the garbage
// collection impact will be the same as a conventionally allocated
// datastructure. For example none of the structs below should be managed by an
// Store.
//
//	type BadStruct1 struct {
//	  stringsHavePointers string
//	}
//
//	type BadStruct2 struct {
//	  mapsHavePointers map[int]int
//	}
//
//	type BadStruct3 struct {
//	  slicesHavePointers []int
//	}
//
//	type BadStruct4 struct {
//	  pointersHavePointers *int
//	}
//
//	type BadStruct5 struct {
//	  storesHavePointers *objectstore.Store
//	}
//
// Memory Model Constraints:
//
// Store contains very limited concurrency control internally. It is only safe
// for concurrent access under limit circumstances which are described below.
//
// In a pure single threaded context the use of Store is fairly straightforward
// and should follow familiar behaviour patterns. In order to get an object
// you are must first Alloc() it, the Reference you get from Alloc() can be
// used to Get() a pointer to the object again in the future. You can Free()
// an object via its Reference, but it is only safe to do this once. Calling
// Free() multiple times on the same Reference has unpredictable behaviour
// (although we make a best-effort to panic). Calling Get() on a Reference
// after Free() has been called on that Reference has similarly unpredictable
// behaviour (we try to panic here too, but this behaviour cannot be relied
// on).
//
// Supported Concurrent Designs
//
// The design of the Store supports single-threaded construction of a
// datastructure which then becomes read-only. This allows an unlimited number
// of concurrent readers. This is a simple and robust design, but won't be
// useful for all systems because the datastructure cannot be modified after
// construction.
//
// Alternatively, it is possible to use the Store to build single-reader
// multiple writer datastructures safely. The design must avoid calls to
// Alloc/Free in readers and take care to safely publish newly allocated
// objects to the readers using some kind of happens-before barrier, channel
// send/receive mutex lock/unlock or atomic write/read etc. This allows for
// MVCC style datastructures to be developed safely. Tree style datastructures
// are ideal for this approach and it is likely that most datastructures
// developed using the Store will be tree based.
//
// 1: Independent Read Safety
//
// For a given set of live objects, previously allocated with a happens-before
// barrier between the allocator and readers, all objects can be read freely.
// Calling Get() and performing arbitrary reads of the retrieved objects from
// multiple goroutines with no other concurrency control code will work without
// data races.
//
// This guarantee continues to hold even if another goroutine is calling
// Alloc() and Free() to _independent_ objects/References concurrently with the
// reads.
//
// This seems like an unremarkable guarantee to make, but it does constrain the
// Store implementation in interesting ways. For example we cannot add a
// non-atomic read counter to Get() calls because this would be an uncontrolled
// concurrent read/write.
//
// 2: Independent Alloc Safety
//
// It is safe and possible for a writer to allocate new objects, using Alloc(),
// and then make those objects/references available to readers across a
// happens-before barrier. Preserving this guarantee requires us to ensure all
// data on the path of Get() for objects unrelated to the indepenent Alloc()
// calls are never written to during the call to Alloc().
//
// An example of an implementation restriction produced by the independent
// allocation safety rule is that we cannot re-allocate the backing slice for
// allocated objects without some form of concurrency protection. This
// protection is required because objects are stored in a slice of slices i.e.
// `objects [][]O`. When there are no free slots available in the existing
// slices we must create a new slice and append it to objects. Calls to Get()
// _must_ read from `objects` to find the slot required, so any unprotected
// read or write of `objects` will be racy.
//
// 3: Free Safety
//
// It is only safe for an object to be Freed once. It is up to the programmer
// to ensure that an object which has been Freed is never used again, and you
// must not call Get() with that object's Reference again. It is envisioned
// that a single writer will be responsible for both Alloc() and Free() calls,
// and a careful mechanism must be established to ensure Freed objects are
// never read again. The reason for mandating that the single writer be
// responsible for calling both Alloc() and Free() calls is that calls to
// Free() make the freed object available to the next call of Alloc(). Calling
// Alloc() after calling Free() from different goroutines without a
// happens-before barrier between them will always create a data-race, even
// when the Alloc() and Free() calls seem independent to the client program.

package objectstore

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const objectSlabSize = 1024

type Stats struct {
	Allocs    int
	Frees     int
	RawAllocs int
	Live      int
	Reused    int
	Slabs     int
}

type Store[O any] struct {
	// Immutable fields
	slabSize uint64

	// Accounting fields
	allocs atomic.Uint64
	frees  atomic.Uint64
	reused atomic.Uint64

	// allIdx provides unique allocation locations for each new allocation
	allocIdx atomic.Uint64

	// freeRWLock protects rootFree
	freeLock sync.Mutex
	rootFree Reference[O]

	// objectsLock protects objects
	// Allocating to an existing slab with a free slot only needs a read lock
	// Adding a new slab to objects requires a write lock
	objectsLock sync.RWMutex
	objects     []*[objectSlabSize]object[O]
}

// If the object has a non-nil nextFree pointer then the object is currently
// free. Object's which have never been allocated are implicitly free, but have
// a nil nextFree
type object[O any] struct {
	nextFree Reference[O]
	value    O
}

func New[O any]() *Store[O] {
	slabSize := uint64(objectSlabSize)
	objects := []*[objectSlabSize]object[O]{}
	return &Store[O]{
		slabSize: slabSize,
		allocIdx: atomic.Uint64{},
		objects:  objects,
	}
}

func (s *Store[O]) Alloc() (Reference[O], *O) {
	s.allocs.Add(1)

	r, o := s.allocFromFree()
	if o != nil {
		s.reused.Add(1)
		return r, o
	}

	// allocFromFree failed, fall back to allocating from new slot
	return s.allocFromOffset()
}

func (s *Store[O]) Free(r Reference[O]) {
	o := r.getObject()

	if !o.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed object %v", r))
	}

	s.freeLock.Lock()
	defer s.freeLock.Unlock()

	if s.rootFree.IsNil() {
		o.nextFree = r
	} else {
		o.nextFree = s.rootFree
	}

	s.rootFree = r

	s.frees.Add(1)
}

func (s *Store[O]) GetStats() Stats {
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

func (s *Store[O]) allocFromFree() (Reference[O], *O) {
	s.freeLock.Lock()
	defer s.freeLock.Unlock()

	// No free objects available - allocFromFree failed
	if s.rootFree.IsNil() {
		return Reference[O]{}, nil
	}

	// Get pointer to the next available freed slot
	alloc := s.rootFree

	// Grab the meta-data for the slot and nil out the, now
	// allocated, slot's nextFree pointer
	freeObject := alloc.getObject()
	nextFree := freeObject.nextFree
	freeObject.nextFree = Reference[O]{}

	// If the nextFree pointer points to the just allocated slot, then
	// there are no more freed slots available
	s.rootFree = nextFree
	if nextFree == alloc {
		s.rootFree = Reference[O]{}
	}

	return alloc, &freeObject.value
}

func (s *Store[O]) allocFromOffset() (Reference[O], *O) {
	allocIdx := s.acquireAllocIdx()
	// TODO do some power of 2 work here, to eliminate all this division
	slabIdx := allocIdx / s.slabSize
	offsetIdx := allocIdx % s.slabSize

	// Take read lock to access s.objects
	s.objectsLock.RLock()
	if slabIdx >= uint64(len(s.objects)) {
		// Release read lock
		s.objectsLock.RUnlock()
		s.growObjects(int(slabIdx + 1))
		// Reacquire read lock
		s.objectsLock.RLock()
	}
	obj := &(s.objects[slabIdx][offsetIdx])
	// Release read lock
	s.objectsLock.RUnlock()

	ref := newReference[O](obj)
	return ref, &(obj.value)
}

func (s *Store[O]) acquireAllocIdx() uint64 {
	for {
		allocIdx := s.allocIdx.Load()
		if s.allocIdx.CompareAndSwap(allocIdx, allocIdx+1) {
			// Success
			return allocIdx
		}
	}
}

func (s *Store[O]) growObjects(targetLen int) {
	newSlab := mmapSlab[O]()

	// Acquire write lock to grow the objects slice
	s.objectsLock.Lock()

	s.objects = append(s.objects, newSlab)

	for len(s.objects) < targetLen {
		// Create a new slab
		newSlab := mmapSlab[O]()
		s.objects = append(s.objects, newSlab)
	}

	// Release write lock
	s.objectsLock.Unlock()
}
