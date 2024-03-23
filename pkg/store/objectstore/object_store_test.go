package objectstore

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MutableStruct struct {
	Field int
}

// Demonstrate that we can create an object, modify that object and when we get
// that object from the store we can see the modifications
// We ensure that we allocate so many objects that we will need more than one chunk
// to store all objects.
func Test_Object_NewModifyGet(t *testing.T) {
	os := New[MutableStruct]()

	// Create all the objects and modify field
	pointers := make([]Reference[MutableStruct], objectChunkSize*3)
	for i := range pointers {
		p, s := os.Alloc()
		s.Field = i
		pointers[i] = p
	}

	stats := os.GetStats()
	assert.Equal(t, len(pointers), stats.Allocs)
	assert.Equal(t, len(pointers), stats.Live)
	assert.Equal(t, 0, stats.Frees)

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		s := p.GetValue()
		assert.Equal(t, i, s.Field)
	}
}

// Demonstrate that we can create an object, then get that object and modify it
// we can then get that object again and will see the modification
// We ensure that we allocate so many objects that we will need more than one chunk
// to store all objects.
func Test_Object_GetModifyGet(t *testing.T) {
	os := New[MutableStruct]()

	// Create all the objects
	pointers := make([]Reference[MutableStruct], objectChunkSize*3)
	for i := range pointers {
		p, _ := os.Alloc()
		pointers[i] = p
	}

	stats := os.GetStats()
	assert.Equal(t, len(pointers), stats.Allocs)
	assert.Equal(t, len(pointers), stats.Live)
	assert.Equal(t, 0, stats.Frees)

	// Get each object and modify field
	for i, p := range pointers {
		s := p.GetValue()
		s.Field = i * 2
	}

	// Assert that all of the modifications are visible
	for i, p := range pointers {
		s := p.GetValue()
		assert.Equal(t, i*2, s.Field)
	}
}

// Demonstrate that we can create an object, then free it. If we try to Get()
// the freed object ObjectStore panics
func Test_Object_NewFreeGet_Panic(t *testing.T) {
	os := New[MutableStruct]()
	p, _ := os.Alloc()
	os.Free(p)

	assert.Panics(t, func() { p.GetValue() })
}

// Demonstrate that we can create an object, then free it. If we try to Free()
// the freed object ObjectStore panics
func Test_Object_NewFreeFree_Panic(t *testing.T) {
	os := New[MutableStruct]()
	p, _ := os.Alloc()
	os.Free(p)

	assert.Panics(t, func() { os.Free(p) })
}

// Demonstrate that if we create a large number of objects, then free them,
// then allocate that same number again, we re-use the freed objects
func Test_Object_NewFreeNew_ReusesOldObjects(t *testing.T) {
	os := New[MutableStruct]()

	objectAllocations := objectChunkSize * 3

	// Create a large number of objects
	pointers := make([]Reference[MutableStruct], objectChunkSize*3)
	for i := range pointers {
		p, _ := os.Alloc()
		pointers[i] = p
	}

	stats := os.GetStats()
	// We have allocate one batch of objects
	assert.Equal(t, objectAllocations, stats.Allocs)
	// They are all live
	assert.Equal(t, objectAllocations, stats.Live)
	// Nothing has been freed
	assert.Equal(t, 0, stats.Frees)
	// Internally 3 chunks have been created
	assert.Equal(t, 3, stats.Chunks)

	// Free all of those objects
	for _, p := range pointers {
		os.Free(p)
	}

	stats = os.GetStats()
	// We have allocate one batch of objects
	assert.Equal(t, objectAllocations, stats.Allocs)
	// None are live
	assert.Equal(t, 0, stats.Live)
	// We have freed one batch of objects
	assert.Equal(t, objectAllocations, stats.Frees)
	// Internally 3 chunks have been created
	assert.Equal(t, 3, stats.Chunks)

	// Allocate the same number of objects again
	for range pointers {
		os.Alloc()
	}

	stats = os.GetStats()
	// We have allocated 2 batches of objects
	assert.Equal(t, 2*objectAllocations, stats.Allocs)
	// We have freed one batch
	assert.Equal(t, objectAllocations, stats.Live)
	// One batch is live
	assert.Equal(t, objectAllocations, stats.Frees)
	// Internally 3 chunks have been created
	assert.Equal(t, 3, stats.Chunks)
}

// This small test is in response to a bug found in the free implementation.
// The bug was that there is a loop in the `nextFree` of the last freed slot in
// the OjbectStore.  This is because a freed slot must always have a non-nil
// `nextFree` pointer in its meta-data.  However because we weren't checking
// for this exact case the last freed slot would re-add itself to the root
// `nextFree` pointer in the ObjectStore.  This means that in this case calls
// to `Alloc()` would allocate the same slot over and over, meaning multiple
// independently allocated pointers would all point to the same slot.
func TestFreeThenAllocTwice(t *testing.T) {
	os := New[MutableStruct]()

	// Allocate an object
	p1, o1 := os.Alloc()
	o1.Field = 1
	// Free it
	os.Free(p1)

	// Allocate another - this should reuse o1
	_, o2 := os.Alloc()
	o2.Field = 2

	// Allocate a third, this should be a non-recycled allocation
	_, o3 := os.Alloc()
	o3.Field = 3

	// Assert that our allocations are independent of each other
	assert.Equal(t, 2, o2.Field)
	assert.Equal(t, 3, o3.Field)
}

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store instance
// This test should be run with -race
func TestSeparateGoroutines(t *testing.T) {
	os := New[MutableStruct]()

	barrier := sync.WaitGroup{}
	barrier.Add(1)

	complete := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		complete.Add(1)
		go func() {
			defer complete.Done()
			allocateAndModify(t, os, &barrier)
		}()
	}

	barrier.Done()

	complete.Wait()
}

func allocateAndModify(t *testing.T, os *Store[MutableStruct], barrier *sync.WaitGroup) {
	barrier.Wait()
	refs := []Reference[MutableStruct]{}
	for i := 0; i < 10_000; i++ {
		ref, v := os.Alloc()
		refs = append(refs, ref)
		v.Field = i
	}
	for i, ref := range refs {
		v := ref.GetValue()
		assert.Equal(t, v.Field, i)
		os.Free(ref)
	}
}

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store
// In this test we allocate objects and then push the Reference to a shared channel
// Each goroutine then consumes its share of References from the channel and sums the values
// We test that the shared total is what we expect.
// This test should be run with -race
func TestAllocAndShare(t *testing.T) {
	sharedChannel := make(chan Reference[MutableStruct], 100*10_000)

	os := New[MutableStruct]()

	barrier := sync.WaitGroup{}
	barrier.Add(1)
	total := atomic.Uint64{}

	complete := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		complete.Add(1)
		go func() {
			defer complete.Done()
			allocateAndModifyShared(t, os, &barrier, sharedChannel, &total)
		}()
	}

	barrier.Done()

	complete.Wait()

	expectedTotal := uint64(100 * ((10_000 - 1) * (10_000) / 2))

	assert.Equal(t, total.Load(), expectedTotal)
}

func allocateAndModifyShared(
	t *testing.T,
	os *Store[MutableStruct],
	barrier *sync.WaitGroup,
	sharedChan chan Reference[MutableStruct],
	total *atomic.Uint64,
) {
	barrier.Wait()

	for i := 0; i < 10_000; i++ {
		ref, v := os.Alloc()
		v.Field = i
		sharedChan <- ref
	}

	for i := 0; i < 10_000; i++ {
		ref := <-sharedChan
		v := ref.GetValue()
		total.Add(uint64(v.Field))
		os.Free(ref)
	}
}
