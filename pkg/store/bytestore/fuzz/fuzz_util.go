package fuzz

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"reflect"

	"github.com/fmstephe/location-system/pkg/store/bytestore"
)

type TestRun struct {
	objects *Objects
	steps   []Step
}

func NewTestRun(bytes []byte) *TestRun {
	objects := NewObjects()
	tr := &TestRun{
		objects: objects,
		steps:   make([]Step, 0),
	}

	chooser := byte(0)
	step := Step(nil)
	for len(bytes) > 0 {
		consumeByte(&chooser, &bytes)
		switch chooser % 3 {
		case 0:
			step = NewAllocStep(objects, &bytes)
		case 1:
			step = NewFreeStep(objects, &bytes)
		case 2:
			step = NewMutateStep(objects, &bytes)
		}
		tr.AddStep(step)
	}
	return tr
}

func (t *TestRun) Run() {
	fmt.Printf("\nTesting Run with %d steps\n", len(t.steps))
	for _, step := range t.steps {
		step.DoStep()
		t.objects.CheckAll()
	}
}

func (t *TestRun) AddStep(step Step) {
	t.steps = append(t.steps, step)
}

type Step interface {
	DoStep()
}

type Objects struct {
	store    *bytestore.Store
	pointers []bytestore.Pointer
	expected [][]byte
	// Indicates whether a pointer/object is still live (has not been freed)
	live []bool
}

func NewObjects() *Objects {
	return &Objects{
		store:    bytestore.New(),
		pointers: make([]bytestore.Pointer, 0),
		expected: make([][]byte, 0),
		live:     make([]bool, 0),
	}
}

func (o *Objects) Alloc(value []byte) {
	fmt.Printf("Allocating at index %d\n", len(o.pointers))

	ptr, obj := o.store.Alloc(uint32(len(value)))
	expected := make([]byte, len(value))
	copy(obj[:], value[:])
	copy(expected[:], value[:])
	o.pointers = append(o.pointers, ptr)
	o.expected = append(o.expected, expected)
	o.live = append(o.live, true)
}

// Because evenly distributed random numbers tend to mostly be very large we
// artificially generate small ones here to test the full range of allocation
// sizes
func generateAllocationSizes() []uint32 {
	sizes := []uint32{}
	for i := 0; i < 1000; i++ {
		size := rand.Uint32()
		sizes = append(sizes, size)
		for ; size > 0; size = size >> 1 {
			sizes = append(sizes, size)
		}
	}
	return sizes
}

func (o *Objects) Mutate(index uint32, value []byte) {
	if len(o.pointers) == 0 {
		// No objects to mutate
		return
	}

	// Normalise index
	index = index % uint32(len(o.pointers))

	fmt.Printf("Mutating at index %d\n", index)

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
	value   []byte
}

func NewAllocStep(objects *Objects, bytes *[]byte) *AllocStep {
	// We generate the (arbitrary) size of this allocation We do this
	// carefully to try to touch a wide range of allocation size classes.
	//
	// The maximum size class generated here is 2^16, which isn't the
	// largest size class available. But seems like a reasonable tradeoff
	// for practical fuzzing.
	//
	// The aproach used here will generate a range of smaller size classes,
	// but it's unclear how even the distribution is right now.  Worth
	// checking in the future.

	// Generate a size for the slice
	size := uint16(0)
	consumeUint16(&size, bytes)

	// A random 16 bit value is biased towards allocating large size
	// classes. So we need to force the values to lower values to exercise
	// all allocation size classes.
	//
	// b will end up with a value between 0-15
	b := byte(0)
	consumeByte(&b, bytes)
	b &= 0x0F

	// Mask preserves only the b lowest bits in size
	// This forces more values down to lower size classes
	mask := uint32(1<<b) - 1
	size &= uint16(mask)

	step := &AllocStep{
		objects: objects,
		value:   make([]byte, size),
	}

	consumeBytes(step.value, bytes)
	return step
}

func (s *AllocStep) DoStep() {
	s.objects.Alloc(s.value)
}

// Free an object
type FreeStep struct {
	objects *Objects
	index   uint32
}

func NewFreeStep(objects *Objects, bytes *[]byte) *FreeStep {
	step := &FreeStep{
		objects: objects,
	}
	consumeUint32(&step.index, bytes)
	return step
}

func (s *FreeStep) DoStep() {
	s.objects.Free(s.index)
}

type MutateStep struct {
	objects  *Objects
	index    uint32
	newValue []byte
}

func NewMutateStep(objects *Objects, bytes *[]byte) *MutateStep {
	step := &MutateStep{
		objects: objects,
	}
	consumeUint32(&step.index, bytes)
	consumeBytes(step.newValue[:], bytes)
	return step
}

func (s *MutateStep) DoStep() {
	s.objects.Mutate(s.index, s.newValue)
}

func consumeBytes(dest []byte, bytes *[]byte) {
	copy(dest, *bytes)
	if len(*bytes) <= len(dest) {
		*bytes = (*bytes)[:0]
		return
	}
	*bytes = (*bytes)[len(dest):]
}

func consumeUint16(value *uint16, bytes *[]byte) {
	dest := make([]byte, 2)
	consumeBytes(dest, bytes)
	*value = binary.LittleEndian.Uint16(dest)
}

func consumeUint32(value *uint32, bytes *[]byte) {
	dest := make([]byte, 4)
	consumeBytes(dest, bytes)
	*value = binary.LittleEndian.Uint32(dest)
}

func consumeByte(value *byte, bytes *[]byte) {
	dest := make([]byte, 1)
	consumeBytes(dest, bytes)
	*value = dest[0]
}
