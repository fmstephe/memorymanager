package objectstore

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

const allocsPerGoroutine = 1000
const goroutines = 100

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store instance
// This test should be run with -race
func TestSeparateGoroutines_Race(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	barrier := sync.WaitGroup{}
	barrier.Add(1)

	complete := sync.WaitGroup{}
	for i := 0; i < goroutines; i++ {
		complete.Add(1)
		go func() {
			defer complete.Done()
			allocateAndModify(t, os, &barrier)
		}()
	}

	barrier.Done()

	complete.Wait()
}

func allocateAndModify(t *testing.T, os *Store, barrier *sync.WaitGroup) {
	barrier.Wait()
	refs := []RefObject[MutableStruct]{}
	for i := 0; i < allocsPerGoroutine; i++ {
		ref, v := AllocObject[MutableStruct](os)
		refs = append(refs, ref)
		v.Field = i
	}
	for i, ref := range refs {
		v := ref.GetValue()
		assert.Equal(t, v.Field, i)
		FreeObject(os, ref)
	}
}

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store
// In this test we allocate objects and then push the Reference to a shared channel
// Each goroutine then consumes its share of References from the channel and sums the values
// We test that the shared total is what we expect.
// This test should be run with -race
func TestAllocAndShare_Race(t *testing.T) {
	sharedChannel := make(chan RefObject[MutableStruct], goroutines*allocsPerGoroutine)

	os := NewSized(1 << 8)
	defer os.Destroy()

	barrier := sync.WaitGroup{}
	barrier.Add(1)
	total := atomic.Uint64{}

	complete := sync.WaitGroup{}
	for i := 0; i < goroutines; i++ {
		complete.Add(1)
		go func() {
			defer complete.Done()
			allocateAndModifyShared(t, os, &barrier, sharedChannel, &total)
		}()
	}

	barrier.Done()

	complete.Wait()

	expectedTotal := uint64(goroutines * ((allocsPerGoroutine - 1) * (1000) / 2))

	assert.Equal(t, total.Load(), expectedTotal)
}

func allocateAndModifyShared(
	t *testing.T,
	os *Store,
	barrier *sync.WaitGroup,
	sharedChan chan RefObject[MutableStruct],
	total *atomic.Uint64,
) {
	barrier.Wait()

	for i := 0; i < allocsPerGoroutine; i++ {
		ref, v := AllocObject[MutableStruct](os)
		v.Field = i
		sharedChan <- ref
	}

	for i := 0; i < allocsPerGoroutine; i++ {
		ref := <-sharedChan
		v := ref.GetValue()
		total.Add(uint64(v.Field))
		FreeObject(os, ref)
	}
}

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store
// In this test we allocate objects and then push the Reference to a shared channel
// Each goroutine then consumes its share of References from the channel and sums the values
// We test that the shared total is what we expect.
// This test should be run with -race
func TestAllocAndShare_Multitype_Race(t *testing.T) {
	sharedChannel := make(chan *MultitypeAllocation, goroutines*allocsPerGoroutine)

	os := NewSized(1 << 8)
	defer os.Destroy()

	barrier := sync.WaitGroup{}
	barrier.Add(1)

	complete := sync.WaitGroup{}
	for i := 0; i < goroutines; i++ {
		complete.Add(1)
		go func() {
			defer complete.Done()
			allocateAndModifySharedMultitype(t, os, &barrier, sharedChannel)
		}()
	}

	barrier.Done()
	complete.Wait()
}

func allocateAndModifySharedMultitype(
	t *testing.T,
	os *Store,
	barrier *sync.WaitGroup,
	sharedChan chan *MultitypeAllocation,
) {
	barrier.Wait()

	for i := 0; i < allocsPerGoroutine; i++ {
		allocation := allocAndWrite(os, i)
		sharedChan <- allocation
	}

	for i := 0; i < allocsPerGoroutine; i++ {
		allocation := <-sharedChan
		allocSlice := allocation.getSlice()
		writeToField(allocSlice, i)
		allocation.free(os)
	}
}
