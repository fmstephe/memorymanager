package objectstore

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fmstephe/location-system/pkg/store/fuzzutil"
)

// The single fuzzer test for objectstore
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

func NewTestRun(byteConsumer []byte) *fuzzutil.TestRun {
	objects := NewObjects()

	stepMaker := func(byteConsumer *fuzzutil.ByteConsumer) fuzzutil.Step {
		chooser := byteConsumer.ConsumeByte()
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

	return fuzzutil.NewTestRun(byteConsumer, stepMaker)
}

type Objects struct {
	store    *Store[[16]byte]
	pointers []Pointer[[16]byte]
	expected []*[16]byte
	// Indicates whether a pointer/object is still live (has not been freed)
	live []bool
}

func NewObjects() *Objects {
	return &Objects{
		store:    New[[16]byte](),
		pointers: make([]Pointer[[16]byte], 0),
		expected: make([]*[16]byte, 0),
		live:     make([]bool, 0),
	}
}

func (o *Objects) Alloc(value [16]byte) {
	fmt.Printf("Allocating %v at index %d\n", value, len(o.pointers))

	ptr, obj := o.store.Alloc()
	expected := [16]byte{}
	copy(obj[:], value[:])
	copy(expected[:], value[:])
	o.pointers = append(o.pointers, ptr)
	o.expected = append(o.expected, &expected)
	o.live = append(o.live, true)
}

func (o *Objects) Mutate(index uint32, value [16]byte) {
	if len(o.pointers) == 0 {
		// No objects to mutate
		return
	}

	// Normalise index
	index = index % uint32(len(o.pointers))

	fmt.Printf("Mutating at index %d with new value %v\n", index, value)

	if !o.live[index] {
		// object has been freed, don't mutate
		return
	}
	// Update the allocated data
	ptr := o.pointers[index]
	obj := o.store.Get(ptr)
	copy(obj[:], value[:])

	// Update the expected
	copy(o.expected[index][:], value[:])

}

func (o *Objects) Free(index uint32) {
	if len(o.pointers) == 0 {
		// No objects to mutate
		return
	}

	// Normalise the index so it points into our slice of pointers
	index = index % uint32(len(o.pointers))

	fmt.Printf("Freeing at index %d\n", index)

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
	o.store.Free(o.pointers[index])
	o.live[index] = false
}

func (o *Objects) CheckAll() {
	for idx := range o.pointers {
		o.checkObject(idx)
	}
}

func (o *Objects) checkObject(index int) {
	if len(o.pointers) == 0 {
		// No objects to mutate
		return
	}

	// Normalise the index so it points into our slice of pointers
	index = index % len(o.pointers)

	if !o.live[index] {
		// Object has already been freed
		// It would be nice to actually test the behaviour of getting a freed object
		// But, right now this behaviour is uncertain.
		// 1: If the object was freed and is still free Get panics
		// 2: If the object was freed, but has been re-allocated Get returns the re-allocated object
		// So it might panic - or it might just grab someone else's allocation
		return
	}

	ptr := o.pointers[index]
	value := o.store.Get(ptr)
	expected := o.expected[index]

	if !reflect.DeepEqual(value, expected) {
		panic(fmt.Sprintf("Unequal values found \n\t%v \n\t%v", value, expected))
	}
}

// Allocate an object
type AllocStep struct {
	objects *Objects
	value   [16]byte
}

func NewAllocStep(objects *Objects, byteConsumer *fuzzutil.ByteConsumer) *AllocStep {
	step := &AllocStep{
		objects: objects,
	}
	copy(step.value[:], byteConsumer.ConsumeBytes(len(step.value)))
	return step
}

func (s *AllocStep) DoStep() {
	s.objects.Alloc(s.value)
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
	}
	step.index = byteConsumer.ConsumeUint32()
	return step
}

func (s *FreeStep) DoStep() {
	s.objects.Free(s.index)
	s.objects.CheckAll()
}

type MutateStep struct {
	objects  *Objects
	index    uint32
	newValue [16]byte
}

func NewMutateStep(objects *Objects, byteConsumer *fuzzutil.ByteConsumer) *MutateStep {
	step := &MutateStep{
		objects: objects,
	}
	step.index = byteConsumer.ConsumeUint32()
	copy(step.newValue[:], byteConsumer.ConsumeBytes(len(step.newValue)))
	return step
}

func (s *MutateStep) DoStep() {
	s.objects.Mutate(s.index, s.newValue)
	s.objects.CheckAll()
}
