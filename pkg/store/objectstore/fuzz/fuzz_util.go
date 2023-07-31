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
	objects  []*[16]byte
	expected [][16]byte
	// Indicates whether a pointer/object is still live (has not been freed)
	live []bool
}

func (o *Objects) Add(value []byte) int {
	ptr, obj := o.store.Alloc()
	expected := [16]byte{}
	copy(obj[:], value)
	copy(expected[:], value)
	o.pointers = append(o.pointers, ptr)
	o.objects = append(o.objects, obj)
	o.expected = append(o.expected, expected)
	o.live = append(o.live, true)
	return len(o.pointers) - 1
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
	value   []byte
}

func (s *AllocStep) ConsumeBytes(bytes []byte) []byte {
	s.value = make([]byte, 16)
	if len(bytes) <= 16 {
		copy(s.value, bytes)
		bytes = bytes[:0]
	} else {
		copy(s.value, bytes)
		bytes = bytes[16:]
	}
	return bytes
}

func (s *AllocStep) ProduceBytes() []byte {
	return append([]byte{}, s.value...)
}

func (s *AllocStep) DoStep() {
	s.objects.Add(s.value)
}

// Free an object
type FreeStep struct {
	objects *Objects
	index   uint32
}

func (s *FreeStep) ConsumeBytes(bytes []byte) []byte {
	value := make([]byte, 4)
	if len(bytes) < 4 {
		copy(value, bytes)
		bytes = bytes[:0]
	} else {
		copy(value, bytes[:4])
		bytes = bytes[4:]
	}
	s.index = binary.LittleEndian.Uint32(value)
	return bytes
}

func (s *FreeStep) ProduceBytes() []byte {
	return binary.LittleEndian.AppendUint32([]byte{}, s.index)
}

func (s *FreeStep) DoStep() {
	s.objects.Free(s.index)
}
