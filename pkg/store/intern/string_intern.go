package intern

import (
	"math"
	"strconv"
	"sync"
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/offheap"
)

type Interner struct {
	controller *internController
	store      *offheap.Store

	// lock protected fields
	lock          sync.Mutex // consider a readwrite lock
	internedFloat map[float64]offheap.RefString
	internedInt   map[int64]offheap.RefString
	internedBytes map[uint64]offheap.RefString
}

func newInterner(controller *internController) Interner {
	return Interner{
		controller: controller,
		store:      offheap.New(),

		// lock protected fields
		lock:          sync.Mutex{},
		internedFloat: make(map[float64]offheap.RefString),
		internedInt:   make(map[int64]offheap.RefString),
		internedBytes: make(map[uint64]offheap.RefString),
	}
}

func (i *Interner) GetFromFloat64(floatVal float64) string {
	// Avoid trying to add NaN values into our map
	if math.IsNaN(floatVal) {
		return "NaN"
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	if refString, ok := i.internedFloat[floatVal]; ok {
		return refString.Value()
	}

	str := strconv.FormatFloat(floatVal, 'f', -1, 64)
	if !i.controller.canIntern(str) {
		return str
	}

	// intern int-string and then return interned version
	refString := offheap.AllocStringFromString(i.store, str)
	i.internedFloat[floatVal] = refString

	interned := refString.Value()
	return interned
}

func (i *Interner) GetFromInt64(intVal int64) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	if refString, ok := i.internedInt[intVal]; ok {
		return refString.Value()
	}

	str := strconv.FormatInt(intVal, 10)
	if !i.controller.canIntern(str) {
		return str
	}

	// intern int-string and then return interned version
	refString := offheap.AllocStringFromString(i.store, str)
	i.internedInt[intVal] = refString

	interned := refString.Value()
	return interned
}

func (i *Interner) GetFromBytes(bytes []byte, hash uint64) string {
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
			return iStr
		}
		// Hash collision, return string copy
		return string(bytes)
	}

	if !i.controller.canIntern(str) {
		// Can't intern, return string copy
		return string(bytes)
	}

	// intern string and then return interned version
	refString := offheap.AllocStringFromBytes(i.store, bytes)
	i.internedBytes[hash] = refString

	interned := refString.Value()
	return interned
}
