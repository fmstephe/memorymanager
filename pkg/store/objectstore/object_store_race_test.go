package objectstore

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store instance
// This test should be run with -race
func TestSeparateGoroutines_Race(t *testing.T) {
	os := New()

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

func allocateAndModify(t *testing.T, os *Store, barrier *sync.WaitGroup) {
	barrier.Wait()
	refs := []Reference[MutableStruct]{}
	for i := 0; i < 10_000; i++ {
		ref, v := Alloc[MutableStruct](os)
		refs = append(refs, ref)
		v.Field = i
	}
	for i, ref := range refs {
		v := ref.GetValue()
		assert.Equal(t, v.Field, i)
		Free(os, ref)
	}
}

// Demonstrate that multiple goroutines can alloc/get/free on a shared Store
// In this test we allocate objects and then push the Reference to a shared channel
// Each goroutine then consumes its share of References from the channel and sums the values
// We test that the shared total is what we expect.
// This test should be run with -race
func TestAllocAndShare_Race(t *testing.T) {
	sharedChannel := make(chan Reference[MutableStruct], 100*10_000)

	os := New()

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
	os *Store,
	barrier *sync.WaitGroup,
	sharedChan chan Reference[MutableStruct],
	total *atomic.Uint64,
) {
	barrier.Wait()

	for i := 0; i < 10_000; i++ {
		ref, v := Alloc[MutableStruct](os)
		v.Field = i
		sharedChan <- ref
	}

	for i := 0; i < 10_000; i++ {
		ref := <-sharedChan
		v := ref.GetValue()
		total.Add(uint64(v.Field))
		Free(os, ref)
	}
}
