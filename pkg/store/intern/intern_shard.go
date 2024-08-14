package intern

import (
	"math"
	"strconv"
	"sync"
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/offheap"
)

type internShard struct {
	controller *internController
	store      *offheap.Store

	// lock protected fields
	lock          sync.Mutex // consider a readwrite lock
	internedFloat map[float64]offheap.RefString
	floatStats    Stats
	internedInt   map[int64]offheap.RefString
	intStats      Stats
	internedBytes map[uint64]offheap.RefString
	bytesStats    Stats
}

func newInternShard(controller *internController, store *offheap.Store) internShard {
	return internShard{
		controller: controller,
		store:      store,

		// lock protected fields
		lock:          sync.Mutex{},
		internedFloat: make(map[float64]offheap.RefString),
		internedInt:   make(map[int64]offheap.RefString),
		internedBytes: make(map[uint64]offheap.RefString),
	}
}

func (i *internShard) getFromFloat64(floatVal float64) string {
	// Avoid trying to add NaN values into our map
	if math.IsNaN(floatVal) {
		i.floatStats.returned++
		return "NaN"
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	if refString, ok := i.internedFloat[floatVal]; ok {
		i.floatStats.returned++
		return refString.Value()
	}

	str := strconv.FormatFloat(floatVal, 'f', -1, 64)

	if !i.controller.canInternMaxLen(str) {
		i.floatStats.maxLenExceeded++
		return str
	}

	if !i.controller.canInternUsedBytes(str) {
		i.floatStats.usedBytesExceeded++
		return str
	}

	// intern int-string and then return interned version
	refString := offheap.AllocStringFromString(i.store, str)
	i.internedFloat[floatVal] = refString

	interned := refString.Value()
	i.floatStats.interned++
	return interned
}

func (i *internShard) getFloatStats() Stats {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.floatStats
}

func (i *internShard) getFromInt64(intVal int64) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	if refString, ok := i.internedInt[intVal]; ok {
		i.intStats.returned++
		return refString.Value()
	}

	str := strconv.FormatInt(intVal, 10)

	if !i.controller.canInternMaxLen(str) {
		i.intStats.maxLenExceeded++
		return str
	}

	if !i.controller.canInternUsedBytes(str) {
		i.intStats.usedBytesExceeded++
		return str
	}

	// intern int-string and then return interned version
	refString := offheap.AllocStringFromString(i.store, str)
	i.internedInt[intVal] = refString

	interned := refString.Value()
	i.intStats.interned++
	return interned
}

func (i *internShard) getIntStats() Stats {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.intStats
}

func (i *internShard) getFromBytes(bytes []byte, hash uint64) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	str := unsafe.String(&bytes[0], len(bytes))
	// Perform lookup for existing interned string based on hash
	if refString, ok := i.internedBytes[hash]; ok {
		iStr := refString.Value()
		// Because two different strings _might_ have the same hash we
		// test that the interned string and the submitted string are
		// equal.
		if iStr == str {
			// Return the interned version of the string
			i.bytesStats.returned++
			return iStr
		}
		// Hash collision, return string copy
		i.bytesStats.hashCollision++
		return string(bytes)
	}

	if !i.controller.canInternMaxLen(str) {
		i.bytesStats.maxLenExceeded++
		return string(bytes)
	}
	if !i.controller.canInternUsedBytes(str) {
		i.bytesStats.usedBytesExceeded++
		return string(bytes)
	}

	// intern string and then return interned version
	refString := offheap.AllocStringFromBytes(i.store, bytes)
	i.internedBytes[hash] = refString

	interned := refString.Value()
	i.bytesStats.interned++
	return interned
}

func (i *internShard) getBytesStats() Stats {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.bytesStats
}
