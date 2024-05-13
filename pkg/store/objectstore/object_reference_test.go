package objectstore

import (
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
	os := NewSized(1 << 8)
	defer os.Destroy()

	allocConf := ConfForType[MutableStruct](os)

	// Create all the objects and modify field
	refs := make([]RefObject[MutableStruct], allocConf.ObjectsPerSlab*3)
	for i := range refs {
		r, s := AllocObject[MutableStruct](os)
		s.Field = i
		refs[i] = r
	}

	stats := StatsForType[MutableStruct](os)

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
	os := NewSized(1 << 8)
	defer os.Destroy()

	allocConf := ConfForType[MutableStruct](os)

	// Create all the objects
	refs := make([]RefObject[MutableStruct], allocConf.ObjectsPerSlab*3)
	for i := range refs {
		r, _ := AllocObject[MutableStruct](os)
		refs[i] = r
	}

	stats := StatsForType[MutableStruct](os)

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
	os := NewSized(1 << 8)
	defer os.Destroy()

	r, _ := AllocObject[MutableStruct](os)
	FreeObject(os, r)

	assert.Panics(t, func() { r.GetValue() })
}

// Demonstrate that we can create an object, then free it. If we try to Free()
// the freed object ObjectStore panics
func Test_Object_NewFreeFree_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r, _ := AllocObject[MutableStruct](os)
	FreeObject(os, r)

	assert.Panics(t, func() { FreeObject(os, r) })
}

// Demonstrate that when we double free a re-allocated object we panic
func Test_Object_NewFreeAllocFree_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r, _ := AllocObject[MutableStruct](os)
	FreeObject(os, r)
	// This will re-allocate the just-freed object
	_, _ = AllocObject[MutableStruct](os)

	assert.Panics(t, func() { FreeObject(os, r) })
}

// Demonstrate that the gen check on Free suffers from the ABA problem.
// This means if we re-allocate the same slot repeatedly the gen field will
// eventually overflow and old values will be repeated.
func Test_Object_NewFree256ReallocFree_NoPanic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r, _ := AllocObject[MutableStruct](os)
	oldGen := r.ref.Gen()
	FreeObject(os, r)

	// Keep allocating and free the slot until the gen overflows back to
	// the oldGen value
	temp, _ := AllocObject[MutableStruct](os)
	for temp.ref.Gen() != oldGen {
		// This will re-allocate the just-freed object
		FreeObject(os, temp)
		temp, _ = AllocObject[MutableStruct](os)
	}

	assert.NotPanics(t, func() { FreeObject(os, r) })
}

// Demonstrate that when we get a re-allocated object we panic
func Test_Object_NewFreeAllocGet_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	r, _ := AllocObject[MutableStruct](os)
	FreeObject(os, r)
	// This will re-allocate the just-freed object
	_, _ = AllocObject[MutableStruct](os)

	assert.Panics(t, func() { r.GetValue() })
}

// Demonstrate that the gen check on Free suffers from the ABA problem.
// This means if we re-allocate the same slot repeatedly the gen field will
// eventually overflow and old values will be repeated.
func Test_Object_NewFree256ReallocGet_NoPanic(t *testing.T) {
	os := New()
	defer os.Destroy()

	r, _ := AllocObject[MutableStruct](os)
	oldGen := r.ref.Gen()
	FreeObject(os, r)

	// Keep allocating and free the slot until the gen overflows back to
	// the oldGen value
	temp, _ := AllocObject[MutableStruct](os)
	for temp.ref.Gen() != oldGen {
		// This will re-allocate the just-freed object
		FreeObject(os, temp)
		temp, _ = AllocObject[MutableStruct](os)
	}

	assert.NotPanics(t, func() { r.GetValue() })
}

// Demonstrate that if we create a large number of objects, then free them,
// then allocate that same number again, we re-use the freed objects
func Test_Object_NewFreeNew_ReusesOldObjects(t *testing.T) {
	os := New()
	defer os.Destroy()

	allocConf := ConfForType[MutableStruct](os)

	objectAllocations := int(allocConf.ObjectsPerSlab * 3)

	// Create a large number of objects
	refs := make([]RefObject[MutableStruct], allocConf.ObjectsPerSlab*3)

	for i := range refs {
		r, _ := AllocObject[MutableStruct](os)
		refs[i] = r
	}

	stats := StatsForType[MutableStruct](os)

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
		FreeObject(os, r)
	}

	stats = StatsForType[MutableStruct](os)

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
		AllocObject[MutableStruct](os)
	}

	stats = StatsForType[MutableStruct](os)

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
func Test_Object_FreeThenAllocTwice(t *testing.T) {
	os := New()
	defer os.Destroy()

	// Allocate an object
	r1, o1 := AllocObject[MutableStruct](os)
	o1.Field = 1
	// This is an original object - gen is 0
	assert.Equal(t, byte(0), r1.ref.Gen())
	// Free it
	FreeObject(os, r1)

	// Allocate another - this should reuse o1
	r2, o2 := AllocObject[MutableStruct](os)
	o2.Field = 2
	// This object is re-allocated - gen is 1
	assert.Equal(t, byte(1), r2.ref.Gen())

	// Allocate a third, this should be a non-recycled allocation
	r3, o3 := AllocObject[MutableStruct](os)
	// This is an original object - gen is 0
	assert.Equal(t, byte(0), r3.ref.Gen())
	o3.Field = 3

	// Assert that the references point to distinct memory locations
	assert.NotEqual(t, r2.GetValue(), r3.GetValue())

	// Assert that our allocations are independent of each other
	assert.Equal(t, 2, o2.Field)
	assert.Equal(t, 3, o3.Field)
}

func Test_Object_CheckGenericTypeForPointersInAlloc(t *testing.T) {
	os := New()
	defer os.Destroy()

	// If generic type contains pointers, Alloc will panic
	assert.Panics(t, func() { AllocObject[*int](os) })

	// If generic type does not contain pointers, Alloc will not panic
	assert.NotPanics(t, func() { AllocObject[int](os) })
}

func Test_Object_CannotAllocateVeryBigStruct(t *testing.T) {
	// In principle this code should panic - but Go won't even compile a
	// type this large.  At the time when this test was written the Store's
	// allocation limit is larger than what can be allocated by Go natively
	// on the development system.
	/*
		os := New()
		assert.Panics(t, func() { Alloc[[2 << 49]byte](os) })
	*/
}

// This is a very odd looking test. It is a response to an intermittent failure
// case with zero sized types.  The problem occurs because when we get the data
// portion of the Reference we add the size of the meta-data to the pointer and
// treat that as pointing to the data-portion of the allocation. In the case of
// zero sized types this data-portion does not exist. In most cases the
// data-portion will now point to the beginning of the next allocation's
// meta-data, but for the last allocation in a slab it will point just past the
// allocated memory region. If we get lucky the address just after the
// allocated region is another allocated region, and everything works. If we
// are unlucky the address just after the allocated region is invalid memory
// and we segfault.
//
// Because we cannot rely on being lucky, we now allocate a single byte to
// ensure we can safely point to valid memory even though the zero sized type
// won't use it to read/write.
//
// This test should alert us if this problem ever reappears.
//
// NB: In the future meta-data will likely be moved to a separate allocation
// space and some details described above will become out of date. The test
// will still be useful though. Zero sized types are a likely source of
// edge-case bugs for all eternity.
func Test_Object_ZeroSizedType_FullSlab(t *testing.T) {
	os := New()
	defer os.Destroy()

	allocConf := ConfForType[SizedArrayZero](os)

	lenTotal := 0

	for range allocConf.ObjectsPerSlab * 24 {
		_, val := AllocObject[SizedArrayZero](os)
		lenTotal += len(val.Field[:])
	}

	assert.Equal(t, 0, lenTotal)
}
