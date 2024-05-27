package objectstore

import (
	"reflect"
	"unsafe"

	"github.com/fmstephe/flib/funsafe"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

// Allocates a new string copied from str
func AllocStringFromString(s *Store, str string) (RefString, string) {
	return AllocStringFromBytes(s, funsafe.StringToBytes(str))
}

// Allocates a new string copied from bytes
func AllocStringFromBytes(s *Store, bytes []byte) (RefString, string) {
	idx := indexForSize(len(bytes))

	// Allocate the string
	pRef := s.alloc(idx)
	sRef := newRefString(len(bytes), pRef)

	// Copy the byte data across to the allocated string
	allocBytes := sRef.ref.Bytes(len(bytes))
	copy(allocBytes, bytes)

	// Return the string-ref and string value
	return sRef, sRef.Value()
}

// Allocates a new string which contains the elements of strs concatenated together
func ConcatStrings(s *Store, strs ...string) (RefString, string) {
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
	for _, str := range strs[:len(strs)] {
		allocBytes = append(allocBytes, str...)
	}

	return sRef, sRef.Value()
}

// Append all of the values in 'fromSlice' to the end of the slice 'into'.  The
// reference 'into' is no longer valid after this function returns.  The
// returned reference should be used instead.
func AppendString(s *Store, into RefString, value string) RefString {
	pRef, newCapacity := resizeAndInvalidate[byte](s, into.ref, capacityForSlice(into.length), into.length+len(value))

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

func (r *RefString) Value() (str string) {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&str))
	stringHeader.Data = r.ref.DataPtr()
	stringHeader.Len = r.length
	return str
}

func (r *RefString) IsNil() bool {
	return r.ref.IsNil()
}
