package objectstore

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

// Allocates a new empty slice with the desired length and capacity
func AllocSlice[T any](s *Store, length, capacity int) (RefSlice[T], []T) {
	// TODO this is not fast - we _need_ to cache this type data
	if err := containsNoPointers[T](); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	idx := indexForSlice[T](capacity)
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", sizeForType[T]()))
	}

	pRef := s.alloc(idx)
	sRef := newRefSlice[T](length, capacity, pRef)
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
	newRef := ensureCapacity(s, into, 1)

	// We have the capacity available, append the element
	slice := newRef.Value()
	slice = append(slice, value)
	newRef.length++

	return newRef
}

// Append all of the values in 'fromSlice' to the end of the slice 'into'.  The
// reference 'into' is no longer valid after this function returns.  The
// returned reference should be used instead.
func AppendSlice[T any](s *Store, into RefSlice[T], fromSlice []T) RefSlice[T] {
	newRef := ensureCapacity(s, into, len(fromSlice))

	// We have the capacity available, append the element
	intoSlice := newRef.Value()
	intoSlice = append(intoSlice, fromSlice...)
	newRef.length += len(fromSlice)

	return newRef
}

func ensureCapacity[T any](s *Store, r RefSlice[T], increase int) RefSlice[T] {
	newLength := r.length + increase

	// Ensure we have enough capacity to append into
	if r.capacity >= newLength {
		// There is room to append value _without_ allocating new
		// memory
		return r.realloc()
	}

	// Determine the new capacity for this slice
	// Right now we just grow the slice by powers of two
	// NB: Go doesn't do this - so maybe this is too aggressive
	newCapacity := int(fmath.NxtPowerOfTwo(int64(newLength)))

	// Check if the current allocation slot has enough space for the new
	// capacity. If it does, then we just re-alloc the current reference
	// and grow the slice in-place
	if indexForSlice[T](r.capacity) == indexForSlice[T](newCapacity) {
		newRef := r.realloc()
		newRef.capacity = newCapacity
		return newRef
	}

	// We need to reallocate and copy to slice to make room for the
	// additional element(s)
	//
	// Check nextCapacity for overflow.  On a 64 bit machine we run
	// out of memory long before running out of bits, on 32 bit
	// machine we would basically be about to allocate half of the
	// available memory available.
	if newCapacity < r.capacity {
		// We just overflowed capacity, try to set it to the
		// largest value available
		newCapacity = math.MaxInt
		if newCapacity == r.capacity {
			// The previous capacity was the largest
			// possible
			panic(fmt.Errorf("Cannot grow slice beyond %d", math.MaxInt))
		}
	}

	newRef, newSlice := AllocSlice[T](s, r.length, newCapacity)
	copy(newSlice, r.Value())
	FreeSlice[T](s, r)
	return newRef
}

func FreeSlice[T any](s *Store, r RefSlice[T]) {
	idx := indexForSlice[T](r.capacity)
	s.free(idx, r.ref)
}

func indexForSlice[T any](capacity int) int {
	typeSize := uint64(fmath.NxtPowerOfTwo(int64(sizeForType[T]())))
	sliceSize := uint64(capacity) * typeSize
	return indexForSize(sliceSize)
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
