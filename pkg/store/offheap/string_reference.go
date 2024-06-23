package offheap

import (
	"unsafe"

	"github.com/fmstephe/flib/funsafe"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

// Allocates a new string whose size and contents will be the same as found in
// str.
func AllocStringFromString(s *Store, str string) RefString {
	return AllocStringFromBytes(s, funsafe.StringToBytes(str))
}

// Allocates a new string whose size and contents will be the same as found in
// bytes.
func AllocStringFromBytes(s *Store, bytes []byte) RefString {
	idx := indexForSize(len(bytes))

	// Allocate the string
	pRef := s.alloc(idx)
	sRef := newRefString(len(bytes), pRef)

	// Copy the byte data across to the allocated string
	allocBytes := sRef.ref.Bytes(len(bytes))
	copy(allocBytes, bytes)

	// Return the string-ref and string value
	return sRef
}

// Allocates a new string which contains the elements of strs concatenated together.
func ConcatStrings(s *Store, strs ...string) RefString {
	// Calculate the total string size needed
	totalLength := 0
	for _, str := range strs {
		totalLength += len(str)
	}

	// Allocate the string
	idx := indexForSize(totalLength)
	pRef := s.alloc(idx)
	sRef := newRefString(totalLength, pRef)

	// Copy the byte data across to the allocated string
	allocBytes := sRef.ref.Bytes(totalLength)
	allocBytes = allocBytes[:0]
	for _, str := range strs {
		allocBytes = append(allocBytes, str...)
	}

	return sRef
}

// Returns a new RefString pointing to a string whose size and contents is the
// same as into.Value() + value.
//
// After this function returns into is no longer a valid RefString, and will
// behave as if Free(...) was called on it.  Internally there is an
// optimisation which _may_ reuse the existing allocation slot if possible. But
// externally this function behaves as if a new allocation is made and the old
// one freed.
func AppendString(s *Store, into RefString, value string) RefString {
	pRef, newCapacity := resizeAndInvalidate[byte](s, into.ref, capacityForSlice(into.length), into.length, len(value))

	// We have the capacity available, append the element
	newRef := newRefString(into.length, pRef)
	str := newRef.ref.Bytes(newCapacity)
	copy(str[into.length:], value)
	newRef.length += len(value)

	return newRef
}

// Frees the allocation referenced by r. After this call returns r must never
// be used again. Any use of the string referenced by r will have
// unpredicatable behaviour.
func FreeString(s *Store, r RefString) {
	idx := indexForSize(r.length)
	s.free(idx, r.ref)
}

// A reference to a string. This reference allows us to gain access to an
// allocated string directly.
//
// It is acceptable, and enouraged, to use RefString in fields of types which
// will be managed by a Store. This is acceptable because RefString does not
// contain any conventional Go pointers, unlike native strings.
type RefString struct {
	length int
	ref    pointerstore.RefPointer
}

func newRefString(length int, ref pointerstore.RefPointer) RefString {
	if ref.IsNil() {
		panic("cannot create new RefStr with nil pointerstore.RefStr")
	}

	return RefString{
		length: length,
		ref:    ref,
	}
}

// Returns the raw string pointed to by this RefString.
//
// Care must be taken not to use this string after FreeString(...) has been
// called on this RefString.
func (r *RefString) Value() string {
	return unsafe.String((*byte)((unsafe.Pointer)(r.ref.DataPtr())), r.length)
}

// Returns true if this RefString does not point to an allocated string, false otherwise.
func (r *RefString) IsNil() bool {
	return r.ref.IsNil()
}

// Returns the stats for the allocations size of a string of length.
//
// It is important to note that these statistics apply to the size class
// indicated here. The statistics allocations will capture all allocations for
// this _size_ including allocations for non-slice types.
func StatsForString(s *Store, length int) pointerstore.Stats {
	stats := s.Stats()
	idx := indexForSize(length)
	return stats[idx]
}

// Returns the allocation config for the allocations size of a string of length.
//
// It is important to note that this config apply to the size class indicated
// here. The config apply to all allocations for this _size_ including
// allocations for non-string types.
func ConfForString(s *Store, length int) pointerstore.AllocConfig {
	configs := s.AllocConfigs()
	idx := indexForSize(length)
	return configs[idx]
}
