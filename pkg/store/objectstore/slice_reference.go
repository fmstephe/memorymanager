package objectstore

import (
	"fmt"
	"math/bits"
	"reflect"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

// Allocates a new slice with the desired length and capacity
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
	capacitySize := into.capacitySize()
	lengthSize := into.lengthSize()
	appendSize := sizeForType[T]()

	newSize := capacityForLength(lengthSize+appendSize, capacitySize)

	pRef := resizeAndInvalidate(s, into.ref, capacitySize, newSize)

	// We have the capacity available, append the element
	// TODO sizeForType() is a power of two - devision can be eliminated
	newRef := newRefSlice[T](into.length, int(newSize/sizeForType[T]()), pRef)
	slice := newRef.Value()
	slice = append(slice, value)
	newRef.length++

	return newRef
}

// Append all of the values in 'fromSlice' to the end of the slice 'into'.  The
// reference 'into' is no longer valid after this function returns.  The
// returned reference should be used instead.
func AppendSlice[T any](s *Store, into RefSlice[T], fromSlice []T) RefSlice[T] {
	capacitySize := into.capacitySize()
	lengthSize := into.lengthSize()
	appendSize := uint64(len(fromSlice)) * sizeForType[T]()

	newCapacitySize := capacityForLength(lengthSize+appendSize, capacitySize)

	pRef := resizeAndInvalidate(s, into.ref, capacitySize, newCapacitySize)

	// We have the capacity available, append the slice
	// TODO sizeForType() is a power of two - devision can be eliminated
	newRef := newRefSlice[T](into.length, int(newCapacitySize/sizeForType[T]()), pRef)
	intoSlice := newRef.Value()
	intoSlice = append(intoSlice, fromSlice...)
	newRef.length += len(fromSlice)

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

func (r *RefSlice[T]) capacitySize() uint64 {
	return sizeForType[T]() * uint64(r.capacity)
}

func (r *RefSlice[T]) lengthSize() uint64 {
	return sizeForType[T]() * uint64(r.length)
}

func capacityForLength(lengthSize, capacitySize uint64) uint64 {
	// Get the power of two value that holds lengthSize
	// NB: If lengthSize is 0, capForLength will also be 0
	capForLength := uint64(1 << bits.Len64(lengthSize-1))

	// If the original capacity is larger than the capacity needed for the
	// length We keep the original capacity. We don't want to shrink a
	// slice just because it doesn't _yet_ have enough data in it.
	return max(capForLength, capacitySize)
}
