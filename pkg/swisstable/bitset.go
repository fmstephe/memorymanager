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

func (s *bitSet[I]) nextMatch() int {
	v := bits.TrailingZeros64(uint64(s.values))
	s.values &= ^(1 << v) // Clear out the bit position we will return
	return v >> 3         // divide by 8???
}

func metaMatchH2(m *metadata, h h2) bitset {
	// https://graphics.stanford.edu/~seander/bithacks.html##ValueInWord
	return hasZeroByte(castUint64(m) ^ (loBits * uint64(h)))
}
