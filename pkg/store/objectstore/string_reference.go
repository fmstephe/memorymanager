package objectstore

import (
	"unsafe"

	"github.com/fmstephe/flib/funsafe"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

// Allocates a new string copied from str
func AllocStringFromString(s *Store, str string) RefString {
	return AllocStringFromBytes(s, funsafe.StringToBytes(str))
}

// Allocates a new string copied from bytes
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

// Allocates a new string which contains the elements of strs concatenated together
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

// Append all of the values in 'fromSlice' to the end of the slice 'into'.  The
// reference 'into' is no longer valid after this function returns.  The
// returned reference should be used instead.
func AppendString(s *Store, into RefString, value string) RefString {
	pRef, newCapacity := resizeAndInvalidate[byte](s, into.ref, capacityForSlice(into.length), into.length, len(value))

	// We have the capacity available, append the element
	newRef := newRefString(into.length, pRef)
	str := newRef.ref.Bytes(newCapacity)
	copy(str[into.length:], value)
	newRef.length += len(value)

	return newRef
}

func FreeStr(s *Store, r RefString) {
	idx := indexForSize(r.length)
	s.free(idx, r.ref)
}

// A reference to a string
// length is the len() of the string
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

func (r *RefString) Value() string {
	return unsafe.String((*byte)((unsafe.Pointer)(r.ref.DataPtr())), r.length)
}

func (r *RefString) IsNil() bool {
	return r.ref.IsNil()
}
