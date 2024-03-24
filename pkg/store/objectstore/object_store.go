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
// conventional pointer to retrieve the allocated object Reference.GetValue()
// e.g.
//
//	store := objectstore.New[int]()
//	var ref objectstore.Reference
//	var i1 *int
//	ref, i1 = store.Alloc()
//	var i2 *int
//	i2 = ref.GetValue()
//	if i1 == i2 {
//	  println("This is correct, i1 and i2 will be a pointer to the same int")
//	}
//
// When you know that an allocated object will never be used again it's memory
// can be released back to the Store using Free() e.g.
//
//	store := objectstore.New[int]()
//	ref, i1 := store.Alloc()
//	println(*i1)
//	store.Free(ref)
//	// You must never user i1 or ref again
//
// A best effort has been made to panic if an object is freed twice or if a
// freed object is accessed using Reference.GetValue(). However, it isn't
// guaranteed that these calls will panic. For example if an object is freed
// the next call to Alloc() will reuse the freed object and future calls to
// Reference.GetValue() and Free() using the Reference used in the Free will
// operate on the re-allocated object and not panic. So this behaviour cannot
// be relied on.
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
// The Reference type contains no pointers which are recognised by the garbage
// collector. This means that we can retain as many of them as we like with no
// garbage collection cost. The Store itself contains some pointers, and must
// also be referenced via a pointer to work properly. So storing the Store
// itself in the Node struct above would introduce pointers into the managed
// object type and defeat the purpose of using a Store.
//
// The advantage of using a Store, over just allocating objects the normal
// way, is that we get pointer-like references to objects without those
// references being visible to the garbage collector. This means that we can
// build large datastructures, like trees or linked lists, which can live in
// memory indefinitely but incur almost zero garbage collection cost.
//
// The disadvantage of using a Store, over just allocating objects the normal
// way, is that we don't get to enjoy the benefits of the Go garbage collector.
// We must be responsible for freeing unneeded objects manually.  The
// disadvantages mean that effective use of the objectstore requires a careful
// design which encapsulates these details and hides the use of the objectstore
// package from the rest of the program.
//
// It is important to note that the objects managed by a Store do not exist on
// the managed Go heap. They live in a series of manually mapped memory regions
// which are managed internally by the Store. This means that the amount of
// memory used by the Store has no impact on the frequency of garbage
// collection runs.
//
// If we attempt to manage an object which itself contains conventional Go
// pointers we invalidate the purpose of the Store because the garbage
// collector would have to mark all of the objects in the store providing no
// performance benefit. Alternatively it's possible the garbage collector may
// simply fail to observe the pointers inside the Store's objects and may free
// data pointed too by these objects. For example none of the structs below
// should be managed by an Store.
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
// Store has a moderate degree of concurrency safety, but users must still be
// careful how they access and modify data allocated by a Store instance.
//
// Concurrency Guarantees
//
// 1: Independent Read Safety
//
// For a given set of live objects, previously allocated with a happens-before
// barrier between the allocator and readers, all objects can be read freely.
// Calling Reference.GetValue() and performing arbitrary reads of the retrieved
// objects from multiple goroutines with no other concurrency control code will
// work without data races.
//
// This guarantee continues to hold even if other goroutines are calling
// Alloc() and Free() to _independent_ objects/References concurrently with
// these reads.
//
// This seems like an unremarkable guarantee to make, but it does constrain the
// Store implementation in interesting ways. For example we cannot add a
// non-atomic read counter to Reference.GetValue() calls because this would be
// an uncontrolled concurrent read/write.
//
// 2: Independent Alloc Safety
//
// It is safe and possible for a writer to allocate new objects, using Alloc(),
// and then make those objects/references available to readers across a
// happens-before barrier. Preserving this guarantee requires us to ensure all
// data on the path of Reference.GetValue() for objects unrelated to the
// indepenent Alloc() calls are never written to during the call to Alloc().
//
// 3: Free Safety
//
// It is only safe for an object to be Freed once. It is up to the programmer
// to ensure that an object which has been Freed is never used again, and you
// must not call Reference.GetValue() with that object's Reference again.
//
// If we call Free() on the same Reference (a Reference pointing to the same
// allocation) concurrently from two or more goroutines this will be a data
// race. The behaviour is unpredictable in this case. This is also a bug, but
// potentially one with stranger behaviour than just calling Free() twice from
// a single goroutine.
//
// It is always a bug to call Reference.GetValue() on a Reference which has
// previously had Free() called on it. But, if we call Reference.GetValue() and
// Free() on that Reference from two or more goroutines this will be a data
// race. Similar to the Free() case above the behaviour will be a less
// predictable bug than the single threaded case.
//
// 4: Safe Data Publication
//
// It is safe to create objects using Alloc() and then share those objects with
// other goroutines. We must establish the usual happens-before relationships
// when sharing objects/References with other goroutines. For example it is
// safe to Alloc() new objects and publish References to those objects on a
// channel.
//
// 5: Safe Objecet Reads And Writes
//
// It is safe to call Reference.GetValue() on the same Reference from multiple
// goroutines and it is safe to read the contents of the object returned. It is
// not safe for multiple goroutines to freely write to the object, nor to have
// multiple goroutines freely perform a mixture of read/write operations on the
// object. You can however perform arbitrary reads and writes to a shared
// object. This is a pretty long-winded way to say that allocated objects
// (accessed via Reference.GetValue()) work like normal go data.
//

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
