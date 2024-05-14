package objectstore

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

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

// The reference 'into' is no longer valid after this function returns.
// The returned reference should be used instead.
func Append[T any](s *Store, into RefSlice[T], value T) (newRef RefSlice[T]) {
	// Ensure we have enough capacity to append into
	//
	// TODO there is a missing optimisation in the code below. If the
	// capacity of the slice is smaller than the allocation slot then the
	// capacity can be enlarged without allocating a new slot for the
	// slice. We'll do this in the future.
	if into.capacity > into.length {
		// There is room to append value _without_ allocating new
		// memory
		newRef = into
	} else {
		// We need to reallocate and copy to slice to make room for the
		// additional value
		nextCapacity := int(fmath.NxtPowerOfTwo(int64(into.capacity + 1)))

		// Check nextCapacity for overflow.  On a 64 bit machine we run
		// out of memory long before running out of bits, on 32 bit
		// machine we would basically be about to allocate half of the
		// available memory available.
		if nextCapacity < into.capacity {
			// We just overflowed capacity, try to set it to the
			// largest value available
			nextCapacity = math.MaxInt
			if nextCapacity == into.capacity {
				// The previous capacity was the largest
				// possible
				panic(fmt.Errorf("Cannot grow slice beyond %d", math.MaxInt))
			}
		}

		newRef, _ = AllocSlice[T](s, into.length, nextCapacity)
		copy(newRef.Value(), into.Value())
		FreeSlice[T](s, into)
	}

	// We have the capacity available, append the element
	slice := newRef.Value()
	slice = append(slice, value)
	newRef.length++
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
