package objectstore

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/fmstephe/flib/funsafe"
	"github.com/fmstephe/location-system/pkg/store/internal/pointerstore"
)

func AllocStringFromString(s *Store, str string) (RefString, string) {
	return AllocStringFromBytes(s, funsafe.StringToBytes(str))
}

func AllocStringFromBytes(s *Store, bytes []byte) (RefString, string) {
	idx := indexForSize(uint64(len(bytes)))
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", len(bytes)))
	}

	// Allocate the string
	pRef := s.alloc(idx)
	oRef := newRefStr(len(bytes), pRef)

	// Copy the byte data across to the allocated string
	allocBytes := oRef.ref.Bytes(len(bytes))
	copy(allocBytes, bytes)

	// Return the string-ref and string value
	return oRef, oRef.Value()
}

func FreeStr(s *Store, r RefString) {
	idx := indexForSize(uint64(r.length))
	s.free(idx, r.ref)
}

// A reference to a string
// length is the len() of the string
type RefString struct {
	length int
	ref    pointerstore.Reference
}

func newRefStr(length int, ref pointerstore.Reference) RefString {
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
