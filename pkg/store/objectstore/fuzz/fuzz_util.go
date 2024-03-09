package fuzz

import (
	"encoding/binary"
	"fmt"

	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

type Step interface {
	ConsumeByters([]byte) []byte
	ProduceBytes() []byte
	DoStep()
}

type Objects struct {
	store    *objectstore.Store[[16]byte]
	pointers []objectstore.Pointer[[16]byte]
	expected [][16]byte
	// Indicates whether a pointer/object is still live (has not been freed)
	live []bool
}

func (o *Objects) Alloc(value [16]byte) int {
	ptr, obj := o.store.Alloc()
	expected := [16]byte{}
	copy(obj, value)
	copy(expected, value)
	o.pointers = append(o.pointers, ptr)
	o.expected = append(o.expected, expected)
	o.live = append(o.live, true)
	return len(o.pointers) - 1
}

func (o *Objects) Mutate(index uint32, value [16]byte) {
	// Update the allocated data
	ptr := o.pointers[index]
	obj := o.store.Get(ptr)
	copy(obj, value)

	// Update the expected
	copy(o.expected, value)
}

func (o *Objects) Free(index uint32) {
	// Normalise the index so it points into our slice of pointers
	index = index % uint32(len(o.pointers))

	// Handle the expected or unexpected panics
	defer func(wasLive bool) {
		if r := recover(); wasLive == (r != nil) {
			// Either the object was live and we panicked or the
			// object was not live, and we did not panic, either
			// way This is an illegal execution, propogate the
			// panic.
			panic(fmt.Errorf("live (%v) was freed and panicked with %v", wasLive, r))
		}
	}(o.live[index])

	// Free the object at index
	o.store.Free(o.pointers[index])
	o.live[index] = false
}

// Allocate an object
type AllocStep struct {
	objects *Objects
	value   [16]byte
	consumed []byte
}

func (s *AllocStep) ConsumeBytes(bytes []byte) (remaining []byte) {
	s.consumed, remaining = consumeUpTo(bytes, 16)
	copy(s.value, s.consumed)
	return bytes
}

func (s *AllocStep) ProduceBytes() []byte {
	return consumed
}

func (s *AllocStep) DoStep() {
	s.objects.Alloc(s.value)
}

// Free an object
type FreeStep struct {
	objects *Objects
	index   uint32
	consumed []byte
}

func (s *FreeStep) ConsumeBytes(bytes []byte) (remaining []byte) {
	s.index, s.consumed, remaining = consumeUpTo(bytes, 4)
	return remaining
}

func (s *FreeStep) ProduceBytes() []byte {
	return consumed
}

func (s *FreeStep) DoStep() {
	s.objects.Free(s.index)
}

type MutateStep struct {
	objects *Objects
	index uint32
	newValue [16]int
	consumed []byte
}

func (s *MutateStep) ConsumeBytes(bytes []byte) (remaining []byte) {
	s.consumed, remaining = consumeUpTo(bytes, 16)
	copy(s.value, s.consumed)
	return remaining
}

func (s *MutateStep) ProduceBytes() []byte {
	return s.consumed
}

func (s *MutateStep) DoStep() {
	s.objects.Mutate(s.index, s.newValue)
}

func consumeUpTo(bytes []byte, limit int) (consumed, remaining []byte) {
	if len(bytes) <= limit {
		consumed = make([]byte, len(bytes))
		copy(s.consumed, bytes)
		return consumed, bytes[:0]
	}
	consumed = make([]byte, limit)
	copy(s.consumed, bytes)
	return consumed, bytes[limit:]
}

func consumeUint32(bytes []byte) (value uint32, consumed, remaining []byte) {
	consumed, remaining = consumeUpTo(bytes, 4)

	// Build a 4 byte version of consumed to create the index
	value := make([]byte, 4)
	copy(value, consumed)
	value = binary.LittleEndian.Uint32(value)

	return value, consumed, remaining
}
