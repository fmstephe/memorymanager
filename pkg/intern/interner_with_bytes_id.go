package intern

import (
	"sync"
	"unsafe"

	xxhash "github.com/cespare/xxhash/v2"
	"github.com/fmstephe/memorymanager/offheap"
)

// An ConverterWithBytesId converts types to strings which are able to be canonically
// identified by a []byte value.
//
// A good example of this is a plain []byte. But many complex types could use
// this converter with values which can't be canonically identified by a single
// uint64.
//
// We don't include the String() method here, because we will _always_ directly
// convert the []byte into a string. This is important because the []byte value
// is compared, byte-wise, to the existing interned string value identified by
// the hash. If the string value could be different from the []byte from the
// identity then this comparison wouldn't work.
type ConverterWithBytesId interface {
	Identity() []byte
}

// A InternerWithBytesId is the type which manages the interning of strings.
type InternerWithBytesId[C ConverterWithBytesId] struct {
	indexMask  uint64
	controller *internController
	store      *offheap.Store
	shards     []internerWithBytesIdShard
}

// Construct a new InternerWithBytesId with the provided config.
func NewInternerWithBytesId[C ConverterWithBytesId](config Config) InternerWithBytesId[C] {
	controller := newController(config.getMaxLen(), config.getMaxBytes())
	store := config.getStore()
	shardCount := config.getShards()

	shards := make([]internerWithBytesIdShard, nextPowerOfTwo(shardCount))
	for i := range shards {
		shards[i] = newInternerWithBytesIdShard(controller, store)
	}

	return InternerWithBytesId[C]{
		indexMask:  uint64(shardCount - 1),
		controller: controller,
		store:      store,
		shards:     shards,
	}
}

// Converts converter into a string representation
//
// The string value may be retrieved from an interning cache or stored in the
// cache.  Regardless of whether the string is or was interned, the correct
// string value is returned.
func (i *InternerWithBytesId[C]) Get(converter C) string {
	bytes := converter.Identity()
	hash := xxhash.Sum64(bytes)
	idx := i.getIndex(hash)
	return i.shards[idx].get(hash, bytes)
}

// Retrieves the summarised stats for interned strings
func (i *InternerWithBytesId[C]) GetStats() StatsSummary {
	intShards := make([]Stats, 0, len(i.shards))
	for idx := range i.shards {
		intShards = append(intShards, i.shards[idx].getStats())
	}
	return makeSummary(intShards, i.controller)
}

func (i *InternerWithBytesId[C]) getIndex(hash uint64) uint64 {
	return i.indexMask & hash
}

type internerWithBytesIdShard struct {
	controller *internController
	store      *offheap.Store
	//
	lock     sync.Mutex
	interned map[uint64]offheap.RefString
	stats    Stats
}

func newInternerWithBytesIdShard(controller *internController, store *offheap.Store) internerWithBytesIdShard {
	return internerWithBytesIdShard{
		controller: controller,
		store:      store,
		//
		interned: make(map[uint64]offheap.RefString),
	}
}

func (i *internerWithBytesIdShard) get(hash uint64, bytes []byte) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	if len(bytes) == 0 {
		// We hardcode the empty string case here
		i.stats.Returned++
		return ""
	}

	unsafeStr := unsafe.String(&bytes[0], len(bytes))
	if refString, ok := i.interned[hash]; ok {
		internedStr := refString.Value()
		// Because two different strings _might_ have the same hash we
		// test that the interned string and the submitted string are
		// equal.
		if internedStr == unsafeStr {
			// Return the interned version of the string
			i.stats.Returned++
			return internedStr
		}
		// Hash collision, can't intern this string.  Return string
		// copy
		i.stats.HashCollision++
		return string(bytes)
	}

	if !i.controller.canInternMaxLen(unsafeStr) {
		// Too long, can't intern this string. Return string copy
		i.stats.MaxLenExceeded++
		return string(bytes)
	}

	if !i.controller.canInternUsedBytes(unsafeStr) {
		// Too many bytes interned, can't intern this string. Return
		// string copy
		i.stats.UsedBytesExceeded++
		return string(bytes)
	}

	// intern string and then return interned version
	refString := offheap.AllocStringFromBytes(i.store, bytes)
	i.interned[hash] = refString

	i.stats.Interned++
	return refString.Value()
}

func (i *internerWithBytesIdShard) getStats() Stats {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.stats
}
