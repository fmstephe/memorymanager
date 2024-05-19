package objectstore

import (
	"fmt"
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
	idx := indexForSize(uint64(len(bytes)))
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", len(bytes)))
	}

	// Allocate the string
	pRef := s.alloc(idx)
	sRef := newRefStr(len(bytes), pRef)

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
	idx := indexForSize(uint64(totalLength))
	pRef := s.alloc(idx)
	sRef := newRefStr(totalLength, pRef)

	// Copy the byte data across to the allocated string
	allocBytes := sRef.ref.Bytes(totalLength)
	allocBytes = allocBytes[:0]
	for _, str := range strs[:len(strs)] {
		allocBytes = append(allocBytes, str...)
	}

	return sRef, sRef.Value()
}

func FreeStr(s *Store, r RefString) {
	idx := indexForSize(uint64(r.length))
	s.free(idx, r.ref)
}

// A reference to a string
// length is the len() of the string
type RefString struct {
	length int
	ref    pointerstore.RefPointer
}

func newRefStr(length int, ref pointerstore.RefPointer) RefString {
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
