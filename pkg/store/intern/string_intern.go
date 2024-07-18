package intern

import (
	"strconv"
	"sync"
	"unsafe"

	xxhash "github.com/cespare/xxhash/v2"
	"github.com/fmstephe/location-system/pkg/store/offheap"
)

type Interner struct {
	// Immutable fields
	maxBytes int // set to -1 for unlimited bytes
	maxLen   int // set of -1 for unlimited string length

	// Mutable fields
	lock          sync.Mutex // consider a readwrite lock
	internedFloat map[float64]offheap.RefString
	internedInt   map[int64]offheap.RefString
	internedBytes map[uint64]offheap.RefString
	usedBytes     int
	store         *offheap.Store
}

func (i *Interner) GetFromFloat64(val float64) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	if refString, ok := i.internedFloat[val]; ok {
		return refString.Value()
	}

	str := strconv.FormatFloat(val, 'f', -1, 64)
	if !i.canIntern(str) {
		return str
	}

	// intern int-string and then return interned version
	refString := offheap.AllocStringFromString(i.store, str)
	i.internedFloat[val] = refString
	return refString.Value()
}

func (i *Interner) GetFromInt64(val int64) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	if refString, ok := i.internedInt[val]; ok {
		return refString.Value()
	}

	str := strconv.FormatInt(val, 10)
	if !i.canIntern(str) {
		return str
	}

	// intern int-string and then return interned version
	refString := offheap.AllocStringFromString(i.store, str)
	i.internedInt[val] = refString
	return refString.Value()
}

func (i *Interner) GetFromBytes(bytes []byte) string {
	h := xxhash.Sum64(bytes)

	i.lock.Lock()
	defer i.lock.Unlock()

	str := unsafe.String(&bytes[0], len(bytes))
	// Perform lookup for existing interned string based on hash
	if refString, ok := i.internedBytes[h]; ok {
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

	if !i.canIntern(str) {
		// Can't intern, return string copy
		return string(bytes)
	}

	// intern string and then return interned version
	refString := offheap.AllocStringFromBytes(i.store, bytes)
	i.internedBytes[h] = refString
	return refString.Value()
}

func (i *Interner) canIntern(str string) bool {
	if i.maxLen != -1 && len(str) > i.maxLen {
		return false
	}

	if i.maxBytes != -1 && len(str)+i.usedBytes > i.maxBytes {
		return false
	}

	// We assume the caller will actually intern the string now
	// So we account for the increased storage
	i.usedBytes += len(str)
	return true
}
