package swisstable

import "math/bits"

type unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type bitSet[I unsigned] struct {
	values I
}

func newBitSet[I unsigned](values I) bitSet[I] {
	return bitSet[I]{
		values: values,
	}
}

func (s *bitSet) nextMatch() int {
	s := bits.TrailingZeros64(uint64(s.values))
	s.values &= ^(1 << s) // Clear out the bit position we will return
	return s >> 3         // divide by 8???
}
