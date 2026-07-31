// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/fmstephe/fuzzhelper"
)

// The single fuzzer test for offheap
func FuzzObjectStore(f *testing.F) {
	// Generate a set of test cases to get fuzzing off to a good start
	r := rand.New(rand.NewSource(1))
	testCases := [][]byte{
		[]byte{},
		randomBytes(r, 1),
		randomBytes(r, 10),
		randomBytes(r, 50),
		randomBytes(r, 100),
		randomBytes(r, 500),
		randomBytes(r, 1000),
		randomBytes(r, 5000),
		randomBytes(r, 10000),
		randomBytes(r, 50000),
	}
	for _, tc := range testCases {
		f.Add(tc)
	}

	// Run the fuzz tests
	f.Fuzz(func(t *testing.T, bytes []byte) {
		objects := NewObjects()
		defer objects.Cleanup()

		testSteps := fuzzhelper.MakeSliceOf[TestStep]([]TestStep{&AllocStep{}, &FreeStep{}, &MutateStep{}}, bytes)
		for _, testStep := range testSteps {
			testStep.DoStep(objects)
		}
	})
}

func randomBytes(r *rand.Rand, size int) []byte {
	bytes := make([]byte, size)
	r.Read(bytes)
	return bytes
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
	writeToField(allocSlice, value)
	expected := generateField(len(allocSlice), value)
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
	writeToField(allocSlice, value)

	// Update the expected
	writeToField(o.expected[index][:], value)
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

type TestStep interface {
	DoStep(objects *Objects)
}

// Allocate an object
type AllocStep struct {
	AllocFuncSelector uint
	Value             byte
}

func (s *AllocStep) DoStep(objects *Objects) {
	allocFunc := multitypeAllocFunc(s.AllocFuncSelector)
	objects.Alloc(allocFunc, s.Value)
	objects.CheckAll()
}

// Free an object
type FreeStep struct {
	Index uint32
}

func (s *FreeStep) DoStep(objects *Objects) {
	objects.Free(s.Index)
	objects.CheckAll()
}

// Mutate an object
type MutateStep struct {
	Index    uint32
	NewValue byte
}

func (s *MutateStep) DoStep(objects *Objects) {
	objects.Mutate(s.Index, s.NewValue)
	objects.CheckAll()
}
