package objectstore

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

// Allocates a new slice with the desired length and capacity
func AllocSlice[T any](s *Store, length, requestedCapacity int) (RefSlice[T], []T) {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	// Round the requested capacity up to a power of 2
	actualCapacity := sliceCapacity(requestedCapacity)

	idx := indexForSlice[T](actualCapacity)
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", sizeForType[T]()))
	}

	pRef := s.alloc(idx)
	sRef := newRefSlice[T](length, actualCapacity, pRef)
	return sRef, sRef.Value()
}

// Allocates a new slice which contains the elements of slices concatenated together
func ConcatSlices[T any](s *Store, slices ...[]T) (RefSlice[T], []T) {
	totalLength := 0
	for _, slice := range slices {
		totalLength += len(slice)
	}

	r, newSlice := AllocSlice[T](s, totalLength, totalLength)

	newSlice = newSlice[:0]
	for _, slice := range slices {
		newSlice = append(newSlice, slice...)
	}

	return r, newSlice
}

// Append the value onto the end of the slice 'into'.  The reference 'into' is
// no longer valid after this function returns.  The returned reference should
// be used instead.
func Append[T any](s *Store, into RefSlice[T], value T) RefSlice[T] {
	pRef, newCapacity := resizeAndInvalidate[T](s, into.ref, into.capacity, into.length+1)

	// We have the capacity available, append the element
	newRef := newRefSlice[T](into.length, newCapacity, pRef)
	slice := newRef.Value()
	slice = append(slice, value)
	newRef.length++

	return newRef
}

// Append all of the values in 'fromSlice' to the end of the slice 'into'.  The
// reference 'into' is no longer valid after this function returns.  The
// returned reference should be used instead.
func AppendSlice[T any](s *Store, into RefSlice[T], fromSlice []T) RefSlice[T] {
	pRef, newCapacity := resizeAndInvalidate[T](s, into.ref, into.capacity, into.length+len(fromSlice))

	// We have the capacity available, append the slice
	newRef := newRefSlice[T](into.length, newCapacity, pRef)
	intoSlice := newRef.Value()
	intoSlice = append(intoSlice, fromSlice...)
	newRef.length += len(fromSlice)

	return newRef
}

func FreeSlice[T any](s *Store, r RefSlice[T]) {
	idx := indexForSlice[T](r.capacity)
	s.free(idx, r.ref)
}

// A reference to a typed slice
// length is len() of the slice
// capacity is cap() of the slice
type RefSlice[T any] struct {
	length   int
	capacity int
	ref      pointerstore.RefPointer
}

func newRefSlice[T any](length, capacity int, ref pointerstore.RefPointer) RefSlice[T] {
	if ref.IsNil() {
		panic("cannot create new RefSlice with nil pointerstore.RefSlice")
	}

	return RefSlice[T]{
		length:   length,
		capacity: capacity,
		ref:      ref,
	}
}

func (r *RefSlice[T]) Value() (slice []T) {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Data = r.ref.DataPtr()
	sliceHeader.Len = r.length
	sliceHeader.Cap = r.capacity
	return slice
}

func (r *RefSlice[T]) IsNil() bool {
	return r.ref.IsNil()
}

func (r *RefSlice[T]) realloc() RefSlice[T] {
	// Copy this reference
	newRef := *r
	// Reallocate it's pointer reference
	newRef.ref = newRef.ref.Realloc()
	return newRef
}

func resizeAndInvalidate[T any](s *Store, oldRef pointerstore.RefPointer, oldCapacity, newLength int) (newRef pointerstore.RefPointer, newCapacity int) {
	newCapacity = sliceCapacity(newLength)

	// Check if the current allocation slot has enough space for the new
	// capacity. If it does, then we just re-alloc the current reference
	if newCapacity <= oldCapacity {
		return oldRef.Realloc(), oldCapacity
	}

	newIdx := indexForSlice[T](newCapacity)
	newRef = s.alloc(newIdx)

	// Copy the content of the old allocation into the new
	oldCapacitySize := int(sizeForSlice[T](oldCapacity))
	oldValue := oldRef.Bytes(oldCapacitySize)
	newValue := newRef.Bytes(oldCapacitySize)
	copy(newValue, oldValue)

	oldIdx := indexForSlice[T](oldCapacity)
	s.free(oldIdx, oldRef)

	return newRef, newCapacity
}
