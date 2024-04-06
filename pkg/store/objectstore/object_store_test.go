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
// We ensure that we allocate so many objects that we will need more than one slab
// to store all objects.
func Test_Object_NewModifyGet(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)

	// Create all the objects and modify field
	refs := make([]Reference[MutableStruct], objectsPerSlab*3)
	for i := range refs {
		r, s := os.Alloc()
		s.Field = i
		refs[i] = r
	}

	stats := os.GetStats()
	assert.Equal(t, len(refs), stats.Allocs)
	assert.Equal(t, len(refs), stats.Live)
	assert.Equal(t, 0, stats.Frees)

	// Assert that all of the modifications are visible
	for i, r := range refs {
		s := r.GetValue()
		assert.Equal(t, i, s.Field)
	}
}

// Demonstrate that we can create an object, then get that object and modify it
// we can then get that object again and will see the modification
// We ensure that we allocate so many objects that we will need more than one slab
// to store all objects.
func Test_Object_GetModifyGet(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)

	// Create all the objects
	refs := make([]Reference[MutableStruct], objectsPerSlab*3)
	for i := range refs {
		r, _ := os.Alloc()
		refs[i] = r
	}

	stats := os.GetStats()
	assert.Equal(t, len(refs), stats.Allocs)
	assert.Equal(t, len(refs), stats.Live)
	assert.Equal(t, 0, stats.Frees)

	// Get each object and modify field
	for i, r := range refs {
		s := r.GetValue()
		s.Field = i * 2
	}

	// Assert that all of the modifications are visible
	for i, r := range refs {
		s := r.GetValue()
		assert.Equal(t, i*2, s.Field)
	}
}

// Demonstrate that we can create an object, then free it. If we try to Get()
// the freed object ObjectStore panics
func Test_Object_NewFreeGet_Panic(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)
	r, _ := os.Alloc()
	os.Free(r)

	assert.Panics(t, func() { r.GetValue() })
}

// Demonstrate that we can create an object, then free it. If we try to Free()
// the freed object ObjectStore panics
func Test_Object_NewFreeFree_Panic(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)
	r, _ := os.Alloc()
	os.Free(r)

	assert.Panics(t, func() { os.Free(r) })
}

// Demonstrate that when we double free a re-allocated object we panic
func Test_Object_NewFreeAllocFree_Panic(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)
	r, _ := os.Alloc()
	os.Free(r)
	// This will re-allocate the just-freed object
	_, _ = os.Alloc()

	assert.Panics(t, func() { os.Free(r) })
}

// Demonstrate that the gen check on Free suffers from the ABA problem.
// This means if we re-allocate the same slot repeatedly the gen field will
// eventually overflow and old values will be repeated.
func Test_Object_NewFree256ReallocFree_NoPanic(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)
	r, _ := os.Alloc()
	oldGen := r.ref.GetGen()
	os.Free(r)

	// Keep allocating and free the slot until the gen overflows back to
	// the oldGen value
	temp, _ := os.Alloc()
	for temp.ref.GetGen() != oldGen {
		// This will re-allocate the just-freed object
		os.Free(temp)
		temp, _ = os.Alloc()
	}

	assert.NotPanics(t, func() { os.Free(r) })
}

// Demonstrate that when we get a re-allocated object we panic
func Test_Object_NewFreeAllocGet_Panic(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)
	r, _ := os.Alloc()
	os.Free(r)
	// This will re-allocate the just-freed object
	_, _ = os.Alloc()

	assert.Panics(t, func() { r.GetValue() })
}

// Demonstrate that the gen check on Free suffers from the ABA problem.
// This means if we re-allocate the same slot repeatedly the gen field will
// eventually overflow and old values will be repeated.
func Test_Object_NewFree256ReallocGet_NoPanic(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)
	r, _ := os.Alloc()
	oldGen := r.ref.GetGen()
	os.Free(r)

	// Keep allocating and free the slot until the gen overflows back to
	// the oldGen value
	temp, _ := os.Alloc()
	for temp.ref.GetGen() != oldGen {
		// This will re-allocate the just-freed object
		os.Free(temp)
		temp, _ = os.Alloc()
	}

	assert.NotPanics(t, func() { r.GetValue() })
}

// Demonstrate that if we create a large number of objects, then free them,
// then allocate that same number again, we re-use the freed objects
func Test_Object_NewFreeNew_ReusesOldObjects(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)

	objectAllocations := int(objectsPerSlab * 3)

	// Create a large number of objects
	refs := make([]Reference[MutableStruct], objectsPerSlab*3)
	for i := range refs {
		r, _ := os.Alloc()
		refs[i] = r
	}

	stats := os.GetStats()
	// We have allocate one batch of objects
	assert.Equal(t, objectAllocations, stats.Allocs)
	// They are all live
	assert.Equal(t, objectAllocations, stats.Live)
	// Nothing has been freed
	assert.Equal(t, 0, stats.Frees)
	// Internally 3 slabs have been created
	assert.Equal(t, 3, stats.Slabs)

	// Free all of those objects
	for _, r := range refs {
		os.Free(r)
	}

	stats = os.GetStats()
	// We have allocate one batch of objects
	assert.Equal(t, objectAllocations, stats.Allocs)
	// None are live
	assert.Equal(t, 0, stats.Live)
	// We have freed one batch of objects
	assert.Equal(t, objectAllocations, stats.Frees)
	// Internally 3 slabs have been created
	assert.Equal(t, 3, stats.Slabs)

	// Allocate the same number of objects again
	for range refs {
		os.Alloc()
	}

	stats = os.GetStats()
	// We have allocated 2 batches of objects
	assert.Equal(t, 2*objectAllocations, stats.Allocs)
	// We have freed one batch
	assert.Equal(t, objectAllocations, stats.Live)
	// One batch is live
	assert.Equal(t, objectAllocations, stats.Frees)
	// Internally 3 slabs have been created
	assert.Equal(t, 3, stats.Slabs)
}

// This small test is in response to a bug found in the free implementation.
// The bug was that there is a loop in the `nextFree` of the last freed slot in
// the ObjectStore.  This is because a freed slot must always have a non-nil
// `nextFree` reference in its meta-data.  However because we weren't checking
// for this exact case the last freed slot would re-add itself to the root
// `nextFree` reference in the ObjectStore.  This means that in this case calls
// to `Alloc()` would allocate the same slot over and over, meaning multiple
// independently allocated references would all point to the same slot.
func TestFreeThenAllocTwice(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)

	// Allocate an object
	r1, o1 := os.Alloc()
	o1.Field = 1
	// This is an original object - gen is 0
	assert.Equal(t, byte(0), r1.ref.GetGen())
	// Free it
	os.Free(r1)

	// Allocate another - this should reuse o1
	r2, o2 := os.Alloc()
	o2.Field = 2
	// This object is re-allocated - gen is 1
	assert.Equal(t, byte(1), r2.ref.GetGen())

	// Allocate a third, this should be a non-recycled allocation
	r3, o3 := os.Alloc()
	// This is an original object - gen is 0
	assert.Equal(t, byte(0), r3.ref.GetGen())
	o3.Field = 3

	// Assert that the references point to distinct memory locations
	assert.NotEqual(t, r2.GetValue(), r3.GetValue())

	// Assert that our allocations are independent of each other
	assert.Equal(t, 2, o2.Field)
	assert.Equal(t, 3, o3.Field)
}

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store instance
// This test should be run with -race
func TestSeparateGoroutines(t *testing.T) {
	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)

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

	objectsPerSlab := uint64(32)
	os := New[MutableStruct](objectsPerSlab)

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

func Test_New_CheckGenericTypeForPointers(t *testing.T) {
	// If generic type contains pointers, New will panic
	assert.Panics(t, func() { New[*int](32) })

	// If generic type does not contain pointers, New will not panic
	assert.NotPanics(t, func() { New[int](32) })
}
