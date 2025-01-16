package swisstable

import (
	"github.com/dolthub/maphash"
)

type SwissTable[K comparable, V any] struct {
	ctrl   []lowHashMatch
	groups []group[K, V]
	hash   maphash.Hasher[K]
	//
}

const (
	groupSize           = 8
	highHashMask uint64 = 0xffff_ffff_ffff_ff80
	lowHashMask  uint64 = 0x0000_0000_0000_007f
	empty        int8   = -128 // 0b1000_0000
	tombstone    int8   = -2   // 0b1111_1110
)

// highHash is a 57 bit hash prefix
type highHash uint64

// lowHash is a 7 bit hash suffix
type lowHash int8

// lowHashMatch is a short part of the second hash. We match on this hash fragment
// to efficiently check if a hash matches before testing the whole hash
type lowHashMatch [groupSize]uint8

// Group is a set of 8 key/values pairs
// The original version claims there are 16 key/value pairs here???
type group[K comparable, V any] struct {
	keys   [groupSize]K
	values [groupSize]V
}

func (m *SwissTable[K, V]) Has(key K) (ok bool) {
	hi, lo := splitHash(m.hash.Hash(key))
	g := probeStart(hi, len(m.groups))
	for {
		matches := metaMatchLow(&m.ctrl[g], lo)
	}
	return false
}

func splitHash(hash uint64) (highHash, lowHash) {
	return highHash((hash & highHashMask) >> 7), lowHash(hash & lowHashMask)
}

func probeStart(hi highHash, groups int) uint32 {
	return fastModN(uint32(hi), groups)
}

func fastModN(x, n uint32) uint32 {
	// lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32((uint64(x) * uint64(n)) >> 32)
}
