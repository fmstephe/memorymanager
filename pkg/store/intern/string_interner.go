package intern

import (
	"math"
	"math/bits"
	"runtime"

	xxhash "github.com/cespare/xxhash/v2"
	"github.com/fmstephe/location-system/pkg/store/offheap"
)

// A StringInterner is the type which manages the interning of strings.
type StringInterner struct {
	indexMask  uint64
	controller *internController
	store      *offheap.Store
	shards     []internShard
}

// Construct a new StringInterner.
//
// maxLen defines the longest string length which will be interned. 0 means no
// limit.
//
// maxBytes defines the maximum accumulated bytes interned, i.e. sum of
// len(string) for all interned strings. 0 means no limit (this may result in
// memory exhaustion if too many unique strings are interned)
//
// Internally the number of shards will be configured automatically based on
// the number CPUs available.
func New(maxLen, maxBytes int) *StringInterner {
	shardCount := runtime.NumCPU()
	return NewWithShards(maxLen, maxBytes, shardCount)
}

// Construct a new StringInterner.
//
// maxLen defines the longest string length which will be interned. 0 means no
// limit.
//
// maxBytes defines the maximum accumulated bytes interned, i.e. sum of
// len(string) for all interned strings. 0 means no limit (potentially
// dangerous).
//
// shardCount defines the desired number of shards to use to enable
// parallelism. It should be noted that this value is treated as a lower-bound
// and will be rounded up to the nearest power of two.
func NewWithShards(maxLen, maxBytes, shardCount int) *StringInterner {
	controller := newController(maxLen, maxBytes)
	store := offheap.New()

	shards := make([]internShard, nextPowerOfTwo(shardCount))
	for i := range shards {
		shards[i] = newInternShard(controller, store)
	}

	return &StringInterner{
		indexMask:  uint64(shardCount - 1),
		controller: controller,
		store:      store,
		shards:     shards,
	}
}

// Converts floatVal into a string representation using
//
// strconv.FormatFloat(floatVal, 'f', -1, 64)
//
// The string value may be retrieved from an interning cache or stored in the
// cache.  Regardless of whether the string is or was interned, the correct
// string value is returned.
func (i *StringInterner) GetFromFloat64(floatVal float64) string {
	idx := i.getIndex(math.Float64bits(floatVal))
	return i.shards[idx].getFromFloat64(floatVal)
}

// Retrieves the summarised stats for interned float strings
func (i *StringInterner) GetFloatStats() StatsSummary {
	floatShards := make([]Stats, 0, len(i.shards))
	for idx := range i.shards {
		floatShards = append(floatShards, i.shards[idx].getFloatStats())
	}
	return makeSummary(floatShards, i.controller)
}

// Converts intVal into a string representation using
//
// strconv.FormatInt(intVal,10)
//
// The string value may be retrieved from an interning cache or stored in the
// cache.  Regardless of whether the string is or was interned, the correct
// string value is returned.
func (i *StringInterner) GetFromInt64(intVal int64) string {
	idx := i.getIndex(uint64(intVal))
	return i.shards[idx].getFromInt64(intVal)
}

// Retrieves the summarised stats for interned int strings
func (i *StringInterner) GetIntStats() StatsSummary {
	intShards := make([]Stats, 0, len(i.shards))
	for idx := range i.shards {
		intShards = append(intShards, i.shards[idx].getIntStats())
	}
	return makeSummary(intShards, i.controller)
}

// Converts bytes into a string representation using
//
// string(bytes)
//
// The string value may be retrieved from an interning cache or stored in the
// cache.  Regardless of whether the string is or was interned, the correct
// string value is returned.
func (i *StringInterner) GetFromBytes(bytes []byte) string {
	hash := xxhash.Sum64(bytes)

	idx := i.getIndex(hash)
	return i.shards[idx].getFromBytes(bytes, hash)
}

// Retrieves the summarised stats for interned []byte strings
func (i *StringInterner) GetBytesStats() StatsSummary {
	bytesShards := make([]Stats, 0, len(i.shards))
	for idx := range i.shards {
		bytesShards = append(bytesShards, i.shards[idx].getBytesStats())
	}
	return makeSummary(bytesShards, i.controller)
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
