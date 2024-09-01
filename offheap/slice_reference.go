package offheap

import (
	"fmt"
	"unsafe"

	"github.com/fmstephe/offheap/offheap/internal/pointerstore"
)

// Allocates a new slice with the desired length and capacity. The capacity of
// the actual slice may not be the same as requestedCapacity, but it will never
// be smaller than requestedCapacity.
//
// The contents of the slice will be arbitrary. Unlike Go slices acquired via
// AllocSlice do _not_ have their contents zeroed out.
func AllocSlice[T any](s *Store, length, requestedCapacity int) RefSlice[T] {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	// Round the requested capacity up to a power of 2
	actualCapacity := capacityForSlice(requestedCapacity)

	idx := indexForSlice[T](actualCapacity)

	pRef := s.alloc(idx)
	sRef := newRefSlice[T](length, actualCapacity, pRef)
	return sRef
}

// Allocates a new slice which contains the elements of slices concatenated together
func ConcatSlices[T any](s *Store, slices ...[]T) RefSlice[T] {
	totalLength := 0
	// TODO check this value for overflow
	for _, slice := range slices {
		totalLength += len(slice)
	}

	r := AllocSlice[T](s, totalLength, totalLength)
	newSlice := r.Value()

	newSlice = newSlice[:0]
	for _, slice := range slices {
		newSlice = append(newSlice, slice...)
	}

	return r
}

// Returns a new RefSlice pointing to a slice whose size and contents is the
// same as append(into.Value(), value).
//
// After this function returns into is no longer a valid RefSlice, and will
// behave as if Free(...) was called on it.  Internally there is an
// optimisation which _may_ reuse the existing allocation slot if possible. But
// externally this function behaves as if a new allocation is made and the old
// one freed.
func Append[T any](s *Store, into RefSlice[T], value T) RefSlice[T] {
	pRef, newCapacity := resizeAndInvalidate[T](s, into.ref, into.capacity, into.length, 1)

	// We have the capacity available, append the element
	newRef := newRefSlice[T](into.length, newCapacity, pRef)
	newRef.length++
	slice := newRef.Value()
	slice[len(slice)-1] = value

	return newRef
}

// Returns a new RefSlice pointing to a slice whose size and contents is the
// same as append(into.Value(), fromSlice...).
//
// After this function returns into is no longer a valid RefSlice, and will
// behave as if Free(...) was called on it.  Internally there is an
// optimisation which _may_ reuse the existing allocation slot if possible. But
// externally this function behaves as if a new allocation is made and the old
// one freed.
func AppendSlice[T any](s *Store, into RefSlice[T], fromSlice []T) RefSlice[T] {
	pRef, newCapacity := resizeAndInvalidate[T](s, into.ref, into.capacity, into.length, len(fromSlice))

	// We have the capacity available, append the slice
	newRef := newRefSlice[T](into.length, newCapacity, pRef)
	newRef.length += len(fromSlice)
	intoSlice := newRef.Value()
	copy(intoSlice[into.length:], fromSlice)

	return newRef
}

// Frees the allocation referenced by r. After this call returns r must never
// be used again. Any use of the slice referenced by r will have unpredicatable
// behaviour.
func FreeSlice[T any](s *Store, r RefSlice[T]) {
	idx := indexForSlice[T](r.capacity)
	s.free(idx, r.ref)
}

// A reference to a slice. This reference allows us to gain access to an
// allocated slice directly.
//
// It is acceptable, and enouraged, to use RefSlice in fields of types which
// will be managed by a Store. This is acceptable because RefSlice does not
// contain any conventional Go pointers, unlike native slices.
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

// Returns the raw slice pointed to by this RefSlice.
//
// Care must be taken not to use this slice after FreeSlice(...) has been
// called on this RefSlice.
func (r *RefSlice[T]) Value() []T {
	slice := unsafe.Slice((*T)(unsafe.Pointer(r.ref.DataPtr())), r.capacity)
	return slice[:r.length]
}

// Returns true if this RefSlice does not point to an allocated slice, false otherwise.
func (r *RefSlice[T]) IsNil() bool {
	return r.ref.IsNil()
}

// Returns the stats for the allocation size of a []T with capacity.
//
// It is important to note that these statistics apply to the size class
// indicated here. The statistics allocations will capture all allocations for
// this _size_ including allocations for non-slice types.
func StatsForSlice[T any](s *Store, capacity int) pointerstore.Stats {
	stats := s.Stats()
	idx := indexForSlice[T](capacity)
	return stats[idx]
}

// Returns the allocation config for the allocation size of a []T with
// capacity.
//
// It is important to note that this config apply to the size class indicated
// here. The config apply to all allocations for this _size_ including
// allocations for non-slice types.
func ConfForSlice[T any](s *Store, capacity int) pointerstore.AllocConfig {
	configs := s.AllocConfigs()
	idx := indexForSlice[T](capacity)
	return configs[idx]
}

func resizeAndInvalidate[T any](s *Store, oldRef pointerstore.RefPointer, oldCapacity, oldLength, extra int) (newRef pointerstore.RefPointer, newCapacity int) {
	// Calculate the new length
	newLength := oldLength + extra
	// TODO test this overflow case We first need to introduce a new option
	// where the pointerstore tracks allocations with no memory backing the
	// allocation data. This would allow us to create test allocations
	// which are very large without having to actually have that much
	// memory available for the test
	if newLength < oldLength {
		panic(fmt.Errorf("resize (oldLength %d extra %d) has overflowed int", oldLength, extra))
	}

	newCapacity = capacityForSlice(newLength)

	// Check if the current allocation slot has enough space for the new
	// capacity. If it does, then we just re-alloc the current reference
	if newCapacity <= oldCapacity {
		return oldRef.Realloc(), oldCapacity
	}

	newIdx := indexForSlice[T](newCapacity)
	newRef = s.alloc(newIdx)

	// Copy the content of the old allocation into the new
	oldCapacitySize := sizeForSlice[T](oldCapacity)
	oldValue := oldRef.Bytes(oldCapacitySize)
	newValue := newRef.Bytes(oldCapacitySize)
	copy(newValue, oldValue)

	oldIdx := indexForSlice[T](oldCapacity)
	s.free(oldIdx, oldRef)

	return newRef, newCapacity
}
