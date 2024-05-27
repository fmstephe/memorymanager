package objectstore

import (
	"fmt"
	"math/bits"
	"reflect"
	"unsafe"
)

// The maximum number of bits allowable for an allocation, given the CPU
// architecture we are running on
func maxAllocationBits() int {
	return maxAllocationBitsInternal(wordSizeBits())
}

// This function exists to allow easy unit testing of this functions behaviour
//
// An important feature of these values is that for each architecture an 'int'
// value is able to capture any legal allocation size
func maxAllocationBitsInternal(wordSizeBits uintptr) int {
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
func maxAllocationSizeInternal(wordSizeBits uintptr) int {
	return 1 << (maxAllocationBitsInternal(wordSizeBits) - 1)
}

// Indicate the number of bits in a word for the CPU architecture we are
// running on
func wordSizeBits() uintptr {
	return unsafe.Sizeof(uintptr(0)) * 8
}

func indexForType[T any]() int {
	size := sizeForType[T]()
	return indexForSize(size)
}

func sizeForType[T any]() uint64 {
	t := reflect.TypeFor[T]()
	return residentObjectSize(uint64(t.Size()))
}

func sizeForSlice[T any](capacity int) uint64 {
	tSize := sizeForType[T]()
	return nextPowerOfTwo(tSize * uint64(capacity))
}

func capacityForSlice(requestedCapacity int) int {
	return int(nextPowerOfTwo(uint64(requestedCapacity)))
}

func indexForSlice[T any](capacity int) int {
	sliceSize := sizeForSlice[T](capacity)
	return indexForSize(sliceSize)
}

func indexForSize(size uint64) int {
	if size == 0 {
		return 0
	}
	return bits.Len64(uint64(size) - 1)
}

// Returns the smallest power of two >= val
// With the exception that 0 sized objects are size 1 in memory
func residentObjectSize(val uint64) uint64 {
	if val == 0 {
		return 1
	}
	return nextPowerOfTwo(val)
}

func nextPowerOfTwo(val uint64) uint64 {
	if isPowerOfTwo(val) {
		return val
	}
	return 1 << bits.Len64(uint64(val))
}

// Returns true if val is a power of two, otherwise returns false
func isPowerOfTwo(val uint64) bool {
	return val&(val-1) == 0
}
