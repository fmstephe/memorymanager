// The objectstore package allows us to alloc, free and retrieve Go objects.
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
//	var ref objectstore.Reference[int]
//	var i1 *int
//	ref, i1 = store.Alloc()
//	var i2 *int
//	i2 = ref.GetValue()
//	if i1 == i2 {
//	  println("This is correct, i1 and i2 are pointers to the same int")
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
// the next call to Alloc() may reuse that freed object and future calls to
// Reference.GetValue() and Free() will operate on the re-allocated object and
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
// objectstore requires a careful design which encapsulates these details and
// hides the use of the objectstore package from the rest of the program.
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
// and call Reference.GetValue() on those References.
//
// 3: Independent Read Safety
//
// For a given set of live objects, previously allocated with a happens-before
// barrier between the allocator and readers, all objects can be read freely.
// Calling Reference.GetValue() and performing arbitrary reads of the retrieved
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
// Reference.GetValue() this is a data race. This will have unpredictable
// behaviour, and is never safe.
package objectstore

import (
	"fmt"
	"math/bits"
	"reflect"

	"github.com/fmstephe/location-system/pkg/store/pointerstore"
)

const slabSize = 1 << 13

type Store struct {
	sizedStores []*pointerstore.Store
}

func New() *Store {
	return &Store{
		sizedStores: initSizeStore(),
	}
}

func initSizeStore() []*pointerstore.Store {
	// We allow allocations up to 48 bits in size
	//
	// The upper limit is chosen because (at time of writing) X86 systems
	// only use 48 bits in 64 bit addresses so this size limit feels like
	// an inclusive and generous upper limit
	slabs := make([]*pointerstore.Store, 48)

	for i := range slabs {
		slabs[i] = pointerstore.New(pointerstore.NewAllocationConfigBySize(1<<i, slabSize))
	}

	return slabs
}

func Alloc[T any](s *Store) (Reference[T], *T) {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	idx := indexForType[T]()
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", sizeForType[T]()))
	}

	pRef := s.alloc(idx)
	oRef := newReference[T](pRef)
	return oRef, oRef.GetValue()
}

func (s *Store) alloc(idx int) pointerstore.Reference {
	return s.sizedStores[idx].Alloc()
}

func Free[T any](s *Store, r Reference[T]) {
	idx := indexForType[T]()
	s.free(idx, r.ref)
}

func (s *Store) free(idx int, r pointerstore.Reference) {
	s.sizedStores[idx].Free(r)
}

func (s *Store) Destroy() error {
	for i := range s.sizedStores {
		if err := s.sizedStores[i].Destroy(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetStats() []pointerstore.Stats {
	sizedStats := make([]pointerstore.Stats, len(s.sizedStores))
	for i := range s.sizedStores {
		sizedStats[i] = s.sizedStores[i].GetStats()
	}
	return sizedStats
}

func (s *Store) GetAllocationConfigs() []pointerstore.AllocationConfig {
	sizedAllocConfigs := make([]pointerstore.AllocationConfig, len(s.sizedStores))
	for i := range s.sizedStores {
		sizedAllocConfigs[i] = s.sizedStores[i].GetAllocationConfig()
	}
	return sizedAllocConfigs
}

func indexForType[T any]() int {
	size := sizeForType[T]()

	if size == 0 {
		return 0
	}

	return bits.Len64(uint64(size) - 1)
}

func sizeForType[T any]() uint64 {
	t := reflect.TypeFor[T]()
	return uint64(t.Size())
}
