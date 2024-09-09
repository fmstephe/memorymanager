// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package intern

import (
	"sync"

	"github.com/fmstephe/memorymanager/offheap"
)

// An ConverterWithUint64Id converts types to strings which are able to be canonically
// identified by a uint64 value.
//
// A good example of this is an actual uint64 value. Another example would be a
// time.Time value which is identified by its UnixNanos() value.
type ConverterWithUint64Id interface {
	Identity() uint64
	String() string
}

// A InternerWithUint64Id is the type which manages the interning of strings.
type InternerWithUint64Id[C ConverterWithUint64Id] struct {
	indexMask  uint64
	controller *internController
	store      *offheap.Store
	shards     []internerWithUint64IdShard[C]
}

// Construct a new InternerWithUint64Id with the provided config.
func NewInternerWithUint64Id[C ConverterWithUint64Id](config Config) InternerWithUint64Id[C] {
	controller := newController(config.getMaxLen(), config.getMaxBytes())
	store := config.getStore()
	shardCount := config.getShards()

	shards := make([]internerWithUint64IdShard[C], shardCount)
	for i := range shards {
		shards[i] = newInternerWithUint64IdShard[C](controller, store)
	}

	return InternerWithUint64Id[C]{
		indexMask:  uint64(shardCount - 1),
		controller: controller,
		store:      store,
		shards:     shards,
	}
}

// Returns the string representation of converter.
//
// The string value may be retrieved from an interning cache or stored in the
// cache.  Regardless of whether the string is or was interned, the correct
// string value is returned.
func (i *InternerWithUint64Id[C]) Get(converter C) string {
	idx := i.getIndex(converter.Identity())
	return i.shards[idx].get(converter)
}

// Retrieves the summarised stats for interned int strings
func (i *InternerWithUint64Id[C]) GetStats() StatsSummary {
	intShards := make([]Stats, 0, len(i.shards))
	for idx := range i.shards {
		intShards = append(intShards, i.shards[idx].getStats())
	}
	return makeSummary(intShards, i.controller)
}

func (i *InternerWithUint64Id[C]) getIndex(hash uint64) uint64 {
	return i.indexMask & hash
}

type internerWithUint64IdShard[C ConverterWithUint64Id] struct {
	controller *internController
	store      *offheap.Store
	//
	lock     sync.Mutex
	interned map[uint64]offheap.RefString
	stats    Stats
}

func newInternerWithUint64IdShard[C ConverterWithUint64Id](controller *internController, store *offheap.Store) internerWithUint64IdShard[C] {
	return internerWithUint64IdShard[C]{
		controller: controller,
		store:      store,
		//
		interned: make(map[uint64]offheap.RefString),
	}
}

func (i *internerWithUint64IdShard[C]) get(converter C) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	identity := converter.Identity()

	if refString, ok := i.interned[identity]; ok {
		i.stats.Returned++
		return refString.Value()
	}

	str := converter.String()

	if !i.controller.canInternMaxLen(str) {
		i.stats.MaxLenExceeded++
		return str
	}

	if !i.controller.canInternUsedBytes(str) {
		i.stats.UsedBytesExceeded++
		return str
	}

	// intern int-string and then return interned version
	refString := offheap.AllocStringFromString(i.store, str)
	i.interned[identity] = refString

	interned := refString.Value()
	i.stats.Interned++
	return interned
}

func (i *internerWithUint64IdShard[C]) getStats() Stats {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.stats
}
