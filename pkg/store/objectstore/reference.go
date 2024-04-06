package objectstore

import (
	"unsafe"

	"github.com/fmstephe/location-system/pkg/store/pointerstore"
)

// The address field holds a pointer to an object[O], but also sneaks a
// generation value in the top 8 bits of the address field.
//
// The generation must be masked out to get a usable pointer value. The object
// pointed to must have the same generation value in order to access/free that
// object.
type Reference[O any] struct {
	ref pointerstore.Reference
}

func newReference[O any](ref pointerstore.Reference) Reference[O] {
	if ref.IsNil() {
		panic("cannot create new Reference with nil pointerstore.Reference")
	}

	return Reference[O]{
		ref: ref,
	}
}

func (r *Reference[O]) GetValue() *O {
	return (*O)((unsafe.Pointer)(r.ref.GetDataPtr()))
}

func (r *Reference[O]) getGen() uint8 {
	return r.ref.GetGen()
}

func (r *Reference[O]) IsNil() bool {
	return r.ref.IsNil()
}
