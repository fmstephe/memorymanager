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
	//
	floatLock     sync.Mutex
	floatInterned map[float64]offheap.RefString
	floatStats    Stats
	//
	intLock     sync.Mutex
	intInterned map[int64]offheap.RefString
	intStats    Stats
	//
	bytesLock     sync.Mutex
	bytesInterned map[uint64]offheap.RefString
	bytesStats    Stats
}

func newInternShard(controller *internController, store *offheap.Store) internShard {
	return internShard{
		controller: controller,
		store:      store,

		floatInterned: make(map[float64]offheap.RefString),
		intInterned:   make(map[int64]offheap.RefString),
		bytesInterned: make(map[uint64]offheap.RefString),
	}
}

func (i *internShard) getFromFloat64(floatVal float64) string {
	// Avoid trying to add NaN values into our map
	if math.IsNaN(floatVal) {
		i.floatStats.returned++
		return "NaN"
	}

	i.floatLock.Lock()
	defer i.floatLock.Unlock()

	if refString, ok := i.floatInterned[floatVal]; ok {
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
	i.floatInterned[floatVal] = refString

	interned := refString.Value()
	i.floatStats.interned++
	return interned
}

func (i *internShard) getFloatStats() Stats {
	i.floatLock.Lock()
	defer i.floatLock.Unlock()

	return i.floatStats
}

func (i *internShard) getFromInt64(intVal int64) string {
	i.intLock.Lock()
	defer i.intLock.Unlock()

	if refString, ok := i.intInterned[intVal]; ok {
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
	i.intInterned[intVal] = refString

	interned := refString.Value()
	i.intStats.interned++
	return interned
}

func (i *internShard) getIntStats() Stats {
	i.intLock.Lock()
	defer i.intLock.Unlock()

	return i.intStats
}

func (i *internShard) getFromBytes(bytes []byte, hash uint64) string {
	if len(bytes) == 0 {
		// Return the interned version of the string
		i.bytesStats.returned++
		return ""
	}

	i.bytesLock.Lock()
	defer i.bytesLock.Unlock()

	str := unsafe.String(&bytes[0], len(bytes))
	// Perform lookup for existing interned string based on hash
	if refString, ok := i.bytesInterned[hash]; ok {
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
	i.bytesInterned[hash] = refString

	interned := refString.Value()
	i.bytesStats.interned++
	return interned
}

func (i *internShard) getBytesStats() Stats {
	i.bytesLock.Lock()
	defer i.bytesLock.Unlock()

	return i.bytesStats
}
