package objectstore

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/fmstephe/flib/funsafe"
	"github.com/fmstephe/location-system/pkg/store/pointerstore"
)

func AllocFromStr(s *Store, str string) (RefStr, string) {
	return AllocFromBytes(s, funsafe.StringToBytes(str))
}

func AllocFromBytes(s *Store, bytes []byte) (RefStr, string) {
	idx := indexForSize(uint64(len(bytes)))
	if idx >= len(s.sizedStores) {
		panic(fmt.Errorf("Allocation too large at %d", len(bytes)))
	}

	// Allocate the string
	pRef := s.alloc(idx)
	oRef := newRefStr(len(bytes), pRef)

	// Copy the string across to the allocated string
	data := oRef.ref.GetData(len(bytes))
	copy(data, bytes)

	// Return the string-ref and string value
	return oRef, oRef.Value()
}

func FreeStr(s *Store, r RefStr) {
	idx := indexForSize(uint64(r.size))
	s.free(idx, r.ref)
}

// The address field holds a pointer to an object, but also sneaks a
// generation value in the top 8 bits of the address field.
//
// The generation must be masked out to get a usable pointer value. The object
// pointed to must have the same generation value in order to access/free that
// object.
type RefStr struct {
	size int
	ref  pointerstore.Reference
}

func newRefStr(size int, ref pointerstore.Reference) RefStr {
	if ref.IsNil() {
		panic("cannot create new RefStr with nil pointerstore.RefStr")
	}

	return RefStr{
		size: size,
		ref:  ref,
	}
}

func (r *RefStr) Value() (str string) {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&str))
	stringHeader.Data = r.ref.GetDataPtr()
	stringHeader.Len = r.size
	return str
}

func (r *RefStr) IsNil() bool {
	return r.ref.IsNil()
}

func (r *RefStr) gen() uint8 {
	return r.ref.GetGen()
}
