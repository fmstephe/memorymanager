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
	return residentObjectSize(uint64(t.Size()))
}

func sizeForSlice[T any](capacity int) uint64 {
	tSize := sizeForType[T]()
	return nextPowerOfTwo(tSize * uint64(capacity))
}

func sliceCapacity(capacity int) int {
	return int(nextPowerOfTwo(uint64(capacity)))
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
