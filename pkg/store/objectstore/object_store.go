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
)

const objectChunkSize = 1024

type Stats struct {
	Allocs    int
	Frees     int
	RawAllocs int
	Live      int
	Reused    int
	Chunks    int
}

type Store[O any] struct {
	// Immutable fields
	chunkSize uint64

	// Accounting fields
	allocs int
	frees  int
	reused int

	// Data fields
	allocIdx uint64
	rootFree Reference[O]
	meta     [][]meta[O]
	objects  [][]O
}

// If the meta for an object has a non-nil nextFree pointer then the
// object is currently free. Object's which have never been allocated are
// implicitly free, but have a nil nextFree point in their meta.
type meta[O any] struct {
	nextFree Reference[O]
}

func New[O any]() *Store[O] {
	chunkSize := uint64(objectChunkSize)
	// Initialise the first chunk
	meta := [][]meta[O]{make([]meta[O], chunkSize)}
	objects := [][]O{make([]O, chunkSize)}
	return &Store[O]{
		chunkSize: chunkSize,
		allocIdx:  0,
		meta:      meta,
		objects:   objects,
	}
}

func (s *Store[O]) Alloc() (Reference[O], *O) {
	s.allocs++

	if s.rootFree.IsNil() {
		return s.allocFromOffset()
	}

	s.reused++
	return s.allocFromFree()
}

func (s *Store[O]) Get(r Reference[O]) *O {
	m := s.getMeta(r)
	if !m.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Get freed object %v", r))
	}
	return s.getObject(r)
}

func (s *Store[O]) Free(r Reference[O]) {
	meta := s.getMeta(r)

	if !meta.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to Free freed object %v", r))
	}

	s.frees++

	if s.rootFree.IsNil() {
		meta.nextFree = r
	} else {
		meta.nextFree = s.rootFree
	}

	s.rootFree = r
}

func (s *Store[O]) GetStats() Stats {
	return Stats{
		Allocs:    s.allocs,
		Frees:     s.frees,
		RawAllocs: s.allocs - s.reused,
		Live:      s.allocs - s.frees,
		Reused:    s.reused,
		Chunks:    len(s.objects),
	}
}

func (s *Store[O]) allocFromFree() (Reference[O], *O) {
	// Get pointer to the next available freed slot
	alloc := s.rootFree

	// Grab the meta-data for the slot and nil out the, now
	// allocated, slot's nextFree pointer
	freeMeta := s.getMeta(alloc)
	nextFree := freeMeta.nextFree
	freeMeta.nextFree = Reference[O]{}

	// If the nextFree pointer points to the just allocated slot, then
	// there are no more freed slots available
	s.rootFree = nextFree
	if nextFree == alloc {
		s.rootFree = Reference[O]{}
	}

	return alloc, s.getObject(alloc)
}

func (s *Store[O]) allocFromOffset() (Reference[O], *O) {
	allocIdx := s.allocIdx
	s.allocIdx++
	ref := newReference[O](allocIdx)
	chunkIdx, offsetIdx := ref.chunkAndOffset(s.chunkSize)
	if chunkIdx >= uint64(len(s.objects)) {
		// Create a new chunk
		s.meta = append(s.meta, make([]meta[O], s.chunkSize))
		s.objects = append(s.objects, make([]O, s.chunkSize))
	}
	return ref, &s.objects[chunkIdx][offsetIdx]
}

func (s *Store[O]) getObject(r Reference[O]) *O {
	chunkIdx, offsetIdx := r.chunkAndOffset(s.chunkSize)
	return &s.objects[chunkIdx][offsetIdx]
}

func (s *Store[O]) getMeta(r Reference[O]) *meta[O] {
	chunkIdx, offsetIdx := r.chunkAndOffset(s.chunkSize)
	return &s.meta[chunkIdx][offsetIdx]
}
