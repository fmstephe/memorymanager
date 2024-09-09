package offheap

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fmstephe/memorymanager/testpkg/fuzzutil"
)

// The single fuzzer test for offheap
func FuzzObjectStore(f *testing.F) {
	testCases := fuzzutil.MakeRandomTestCases()
	for _, tc := range testCases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, bytes []byte) {
		tr := NewTestRun(bytes)
		tr.Run()
	})
}

func NewTestRun(bytes []byte) *fuzzutil.TestRun {
	objects := NewObjects()

	stepMaker := func(byteConsumer *fuzzutil.ByteConsumer) fuzzutil.Step {
		chooser := byteConsumer.Byte()
		switch chooser % 3 {
		case 0:
			return NewAllocStep(objects, byteConsumer)
		case 1:
			return NewFreeStep(objects, byteConsumer)
		case 2:
			return NewMutateStep(objects, byteConsumer)
		}
		panic("Unreachable")
	}

	cleanup := func() {
		objects.Cleanup()
	}

	return fuzzutil.NewTestRun(bytes, stepMaker, cleanup)
}

type Objects struct {
	store       *Store
	allocations []*MultitypeAllocation
	expected    [][]byte
	// Indicates whether a reference/object is still live (has not been freed)
	live []bool
}

func NewObjects() *Objects {
	return &Objects{
		store:       New(),
		allocations: make([]*MultitypeAllocation, 0),
		expected:    make([][]byte, 0),
		live:        make([]bool, 0),
	}
}

func (o *Objects) Alloc(allocFunc func(*Store) *MultitypeAllocation, value byte) {
	//fmt.Printf("Allocating %v at index %d\n", value, len(o.allocations))

	allocation := allocFunc(o.store)
	allocSlice := allocation.getSlice()
	writeToField(allocSlice, int(value))
	expected := generateField(len(allocSlice), int(value))
	o.allocations = append(o.allocations, allocation)
	o.expected = append(o.expected, expected)
	o.live = append(o.live, true)
}

func (o *Objects) Mutate(index uint32, value byte) {
	if len(o.allocations) == 0 {
		// No objects to mutate
		return
	}

	// Normalise index
	index = index % uint32(len(o.allocations))

	//fmt.Printf("Mutating at index %d with new value %v\n", index, value)

	if !o.live[index] {
		// object has been freed, don't mutate
		return
	}
	// Update the allocated data
	allocation := o.allocations[index]
	allocSlice := allocation.getSlice()
	writeToField(allocSlice, int(value))

	// Update the expected
	writeToField(o.expected[index][:], int(value))
}

func (o *Objects) Free(index uint32) {
	if len(o.allocations) == 0 {
		// No objects to mutate
		return
	}

	// Normalise the index so it points into our slice of allocations
	index = index % uint32(len(o.allocations))

	//fmt.Printf("Freeing at index %d\n", index)

	if !o.live[index] {
		// Object has already been freed
		// It would be nice to actually test the behaviour of freeing a freed object
		// But, right now this behaviour is uncertain.
		// 1: If the object was freed and is still free this method call panics
		// 2: If the object was freed, but has been re-allocated this method call frees the re-allocated object
		// So it might panic - or it might break someone else's allocation
		return
	}

	// Free the object at index
	o.allocations[index].free(o.store)
	o.live[index] = false
}

func (o *Objects) CheckAll() {
	for idx := range o.allocations {
		o.checkObject(idx)
	}
}

func (o *Objects) checkObject(index int) {
	if len(o.allocations) == 0 {
		// No objects to mutate
		return
	}

	// Normalise the index so it points into our slice of allocations
	index = index % len(o.allocations)

	if !o.live[index] {
		// Object has already been freed
		// It would be nice to actually test the behaviour of getting a freed object
		// But, right now this behaviour is uncertain.
		// 1: If the object was freed and is still free Get panics
		// 2: If the object was freed, but has been re-allocated Get returns the re-allocated object
		// So it might panic - or it might just grab someone else's allocation
		return
	}

	allocation := o.allocations[index]
	allocSlice := allocation.getSlice()
	expected := o.expected[index]

	if !reflect.DeepEqual(allocSlice, expected) {
		panic(fmt.Sprintf("Unequal values found \n\t%v \n\t%v", allocSlice, expected))
	}
}

func (o *Objects) Cleanup() {
	if err := o.store.Destroy(); err != nil {
		panic(err)
	}
}

// Allocate an object
type AllocStep struct {
	objects   *Objects
	allocFunc func(*Store) *MultitypeAllocation
	value     byte
}

func NewAllocStep(objects *Objects, byteConsumer *fuzzutil.ByteConsumer) *AllocStep {
	step := &AllocStep{
		objects:   objects,
		allocFunc: multitypeAllocFunc(int(byteConsumer.Uint32())),
		value:     byteConsumer.Byte(),
	}
	return step
}

func (s *AllocStep) DoStep() {
	s.objects.Alloc(s.allocFunc, s.value)
	s.objects.CheckAll()
}

// Free an object
type FreeStep struct {
	objects *Objects
	index   uint32
}

func NewFreeStep(objects *Objects, byteConsumer *fuzzutil.ByteConsumer) *FreeStep {
	step := &FreeStep{
		objects: objects,
		index:   byteConsumer.Uint32(),
	}
	return step
}

func (s *FreeStep) DoStep() {
	s.objects.Free(s.index)
	s.objects.CheckAll()
}

type MutateStep struct {
	objects  *Objects
	index    uint32
	newValue byte
}

func NewMutateStep(objects *Objects, byteConsumer *fuzzutil.ByteConsumer) *MutateStep {
	step := &MutateStep{
		objects:  objects,
		index:    byteConsumer.Uint32(),
		newValue: byteConsumer.Byte(),
	}
	return step
}

func (s *MutateStep) DoStep() {
	s.objects.Mutate(s.index, s.newValue)
	s.objects.CheckAll()
}
