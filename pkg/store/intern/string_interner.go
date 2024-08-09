package intern

import (
	"math"
	"math/bits"
	"runtime"

	xxhash "github.com/cespare/xxhash/v2"
)

type StringInterner struct {
	indexMask uint64
	shards    []internShard
}

func New(maxLen, maxBytes int) *StringInterner {
	shardCount := runtime.NumCPU()
	return NewWithShards(maxLen, maxBytes, shardCount)
}

func NewWithShards(maxLen, maxBytes, shardCount int) *StringInterner {
	shardCount = nextPowerOfTwo(shardCount)
	indexMask := uint64(shardCount - 1)

	controller := newController(maxLen, maxBytes)

	shards := make([]internShard, shardCount)
	for i := range shards {
		shards[i] = newInternShard(controller)
	}
	return &StringInterner{
		indexMask: indexMask,
		shards:    shards,
	}
}

func (i *StringInterner) GetFromFloat64(floatVal float64) string {
	idx := i.getIndex(math.Float64bits(floatVal))
	return i.shards[idx].getFromFloat64(floatVal)
}

func (i *StringInterner) GetFromInt64(intVal int64) string {
	idx := i.getIndex(uint64(intVal))
	return i.shards[idx].getFromInt64(intVal)
}

func (i *StringInterner) GetFromBytes(bytes []byte) string {
	hash := xxhash.Sum64(bytes)

	idx := i.getIndex(hash)
	return i.shards[idx].getFromBytes(bytes, hash)
}

func (i *StringInterner) getIndex(hash uint64) uint64 {
	return i.indexMask & hash
}

// Returns the smallest power of two >= val
func nextPowerOfTwo(val int) int {
	if val <= 1 {
		return 1
	}
	// Test if val is a power of two
	if val > 0 && val&(val-1) == 0 {
		return val
	}
	return 1 << bits.Len64(uint64(val))
}
