package objectstore

import (
	"fmt"
	"math/bits"
	"reflect"
	"unsafe"
)

var maxAllocSize = maxAllocationSize()

// The maximum number of bits allowable for an allocation, given the CPU
// architecture we are running on
func maxAllocationBits() int {
	return maxAllocationBitsInternal(wordSizeBits())
}

// This function exists to allow easy unit testing of this functions behaviour
//
// An important feature of these values is that for each architecture an 'int'
// value is able to capture any legal allocation size
func maxAllocationBitsInternal(wordSizeBits int) int {
	switch wordSizeBits {
	case 32:
		// We allow allocations up to 31 bits in size
		//
		// The upper limit is chosen because slices store capacity and
		// length as an int value. So 31 bits is the largest value for
		// the size of a slice on a 32 bit machine. We chose to limit
		// _all_ allocations to 31 bits as a simplification.
		return 31
	case 64:
		// We allow allocations up to 48 bits in size
		//
		// The upper limit is chosen because (at time of writing) X86 systems
		// only use 48 bits in 64 bit addresses so this size limit feels like
		// an inclusive and generous upper limit
		return 48
	default:
		panic(fmt.Errorf("Unsupported architecture word size %d", wordSizeBits))
	}
}

// The maximum allocation size, given the CPU architecture we are running on
func maxAllocationSize() int {
	return maxAllocationSizeInternal(wordSizeBits())
}

// This function exists to allow easy unit testing of this functions behaviour
func maxAllocationSizeInternal(wordSizeBits int) int {
	return 1 << (maxAllocationBitsInternal(wordSizeBits) - 1)
}

// Indicate the number of bits in a word for the CPU architecture we are
// running on
func wordSizeBits() int {
	return int(unsafe.Sizeof(uintptr(0)) * 8)
}

func indexForType[T any]() int {
	size := sizeForType[T]()
	return indexForSize(size)
}

func indexForSlice[T any](capacity int) int {
	sliceSize := sizeForSlice[T](capacity)
	return indexForSize(sliceSize)
}

func sizeForSlice[T any](capacity int) int {
	tSize := sizeForType[T]()
	return residentObjectSize(tSize * capacity)
}

func sizeForType[T any]() int {
	t := reflect.TypeFor[T]()
	return residentObjectSize(int(t.Size()))
}

func indexForSize(size int) int {
	if size == 0 {
		return 0
	}
	return bits.Len(uint(size) - 1)
}

// NB: It is very important to note that this function deals with the capacity
// reported by cap(slice).  This is not the same as the actual memory size of
// the slice or allocation, as it does not include the size of slice's type.
//
// The reason we have a distinct function to calculate the capacity is that
// requested-capacity of 0 is preserved. Whereas an allocation size 0 currently
// occupies 1 byte of memory.
func capacityForSlice(requestedCapacity int) int {
	return nextPowerOfTwo(requestedCapacity)
}

// Returns the smallest power of two >= val
// With the exception that 0 sized objects are size 1 in memory
func residentObjectSize(requestedSize int) int {
	if requestedSize < 0 {
		panic(fmt.Errorf("Allocation size (%d) negative. Very likely uintptr size of type overflowed int", requestedSize))
	}
	if requestedSize == 0 {
		return 1
	}
	residentSize := nextPowerOfTwo(requestedSize)
	if residentSize > maxAllocSize {
		panic(fmt.Errorf("Allocation size (%d, resident %d) too large. Can't exceed %d", requestedSize, residentSize, maxAllocSize))
	}
	if residentSize < 0 {
		panic(fmt.Errorf("Allocation size (%d, resident %d) too large. Has overflowed int", requestedSize, residentSize))
	}
	return residentSize
}

func nextPowerOfTwo(val int) int {
	if isPowerOfTwo(val) {
		return val
	}
	return 1 << bits.Len(uint(val))
}

// Returns true if val is a power of two, otherwise returns false.  NB: This
// function considers 0 to be a power of two. This strictly wrong, but it's an
// acceptable behaviour for us.
func isPowerOfTwo(val int) bool {
	return val&(val-1) == 0
}
