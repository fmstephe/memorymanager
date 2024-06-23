// The offheap package allows us to alloc, free and retrieve Go objects.
//
// Each Store instance can alloc/free a single type of object. This is
// controlled by the generic type of an offheap instance e.g.
//
//	var store *offheap.Store
//	store = offheap.New[int]()
//
// will alloc/free only int values.
//
// Each allocated object has a corresponding Reference which acts like a
// conventional pointer to retrieve the allocated object Reference.Value()
// e.g.
//
//	store := offheap.New[int]()
//	var ref offheap.Reference[int]
//	var i1 *int
//	ref, i1 = store.Alloc()
//	var i2 *int
//	i2 = ref.Value()
//	if i1 == i2 {
//	  println("This is correct, i1 and i2 are pointers to the same int")
//	}
//
// When you know that an allocated object will never be used again it's memory
// can be released back to the Store using Free() e.g.
//
//	store := offheap.New[int]()
//	ref, i1 := store.Alloc()
//	println(*i1)
//	store.Free(ref)
//	// You must never user i1 or ref again
//
// A best effort has been made to panic if an object is freed twice or if a
// freed object is accessed using Reference.Value(). However, it isn't
// guaranteed that these calls will panic. For example if an object is freed
// the next call to Alloc() may reuse that freed object and future calls to
// Reference.Value() and Free() will operate on the re-allocated object and
// not panic. So this behaviour cannot be relied on.
//
// References can be kept and stored in arbitrary datastructures, which can
// themselves be managed by a Store e.g.
//
//	type Node struct {
//	  left  Reference[Node]
//	  right Reference[Node]
//	}
//	nodeStore := New[Node]()
//
// The Reference type contains no pointers which are recognised by the garbage
// collector. This means that we can retain as many of them as we like with no
// garbage collection cost.
//
// The advantage of using a Store, over just allocating objects the normal
// way, is that we get pointer-like references to objects without those
// references being visible to the garbage collector. This means that we can
// build large datastructures, like trees or linked lists, which can live in
// memory indefinitely but incur almost zero garbage collection cost.
//
// The disadvantage of using a Store is that we don't get to enjoy the benefits
// of the Go garbage collector.  We must be responsible for freeing unneeded
// objects manually.  The disadvantages mean that effective use of the
// offheap requires a careful design which encapsulates these details and
// hides the use of the offheap package from the rest of the program.
//
// It is important to note that the objects managed by a Store do not exist on
// the managed Go heap. They live in a series of manually mapped memory regions
// which are managed separately by the Store. This means that the amount of
// memory used by the Store has no impact on the frequency of garbage
// collection runs.
//
// If we attempt to manage an object which itself contains conventional Go
// pointers we invalidate the purpose of the Store because the garbage
// collector would have to mark all of the objects in the store providing no
// performance benefit. Alternatively it's possible the garbage collector may
// simply fail to observe the pointers inside the Store's objects and may free
// data pointed too by these objects. For example none of the types below
// should be managed by a Store.
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
//	  storesHavePointers *Store
//	}
//
// Trying to create an instance of Store with a generic type which contains
// pointers will panic.
//
// Memory Model Constraints:
//
// A Store has a moderate degree of concurrency safety, but users must still be
// careful how they access and modify data allocated by a Store instance. A
// shorter version of the guarantees described below would be to say that
// allocating and retrieving objects managed by a Store has the same guarantees
// and limitations that conventionally allocated Go objects have.
//
// # Concurrency Guarantees
//
// 1: Independent Alloc/Free Safety
//
// It is safe for multiple goroutines using a shared Store instance to call
// Alloc() and Free() generating independent sets of objects/References. They
// can safely read the objects they have allocated without any additional
// concurrency protection.
//
// 2: Safe Data Publication
//
// It is safe to create objects using Alloc() and then share those objects with
// other goroutines. We must establish the usual happens-before relationships
// when sharing objects/References with other goroutines.
//
// For example it is safe to Alloc() new objects and publish References to
// those objects on a channel and have other goroutines read from that channel
// and call Reference.Value() on those References.
//
// 3: Independent Read Safety
//
// For a given set of live objects, previously allocated with a happens-before
// barrier between the allocator and readers, all objects can be read freely.
// Calling Reference.Value() and performing arbitrary reads of the retrieved
// objects from multiple goroutines with no other concurrency control code will
// work without data races.
//
// 4: Safe Object Reads And Writes
//
// It is not safe for multiple goroutines to freely write to the object, nor to
// have multiple goroutines freely perform a mixture of read/write operations
// on the object. You can however perform concurrent reads and writes to a
// shared object if you use appropriate concurrency controls such as sync.Mutex
// etc.
//
// 5: Free Safety
//
// If we call Free() on the same Reference (a Reference pointing to the same
// allocation) concurrently from two or more goroutines this will be a data
// race. The behaviour is unpredictable in this case. This is also a bug, but
// potentially one with stranger behaviour than just calling Free() twice from
// a single goroutine.
//
// If we call Free() on a Reference while another goroutine is calling
// Reference.Value() this is a data race. This will have unpredictable
// behaviour, and is never safe.
package offheap

import (
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
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
