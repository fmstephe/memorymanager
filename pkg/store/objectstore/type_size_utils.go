package objectstore

import (
	"math/bits"
	"reflect"
)

func indexForType[T any]() int {
	size := sizeForType[T]()

	return indexForSize(size)
}

func sizeForType[T any]() uint64 {
	t := reflect.TypeFor[T]()
	return nextPowerOfTwo(uint64(t.Size()))
}

func sizeForSlice[T any](capacity int) uint64 {
	tSize := sizeForType[T]()
	return nextPowerOfTwo(tSize * uint64(capacity))
}

func sliceCapacityFromSize[T any](capacitySize uint64) int {
	tSize := sizeForType[T]()
	// TODO both are a power of 2, division can be eliminated
	return int(capacitySize / tSize)
}

func indexForSize(size uint64) int {
	if size == 0 {
		return 0
	}
	return bits.Len64(uint64(size) - 1)
}

// Returns the smallest power of two >= val
// If val is 0, we return 0
func nextPowerOfTwo(val uint64) uint64 {
	if val == 0 {
		return 0
	}
	if isPowerOfTwo(val) {
		return val
	}
	return 1 << bits.Len64(uint64(val))
}

// Returns true if val is a power of two, otherwise returns false
func isPowerOfTwo(val uint64) bool {
	return val&(val-1) == 0
}
