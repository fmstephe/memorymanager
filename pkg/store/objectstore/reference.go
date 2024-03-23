package objectstore

import (
	"fmt"
	"unsafe"
)

type Reference[O any] struct {
	address uintptr
}

func newReference[O any](obj *object[O]) Reference[O] {
	return Reference[O]{
		address: (uintptr)(unsafe.Pointer(obj)),
	}
}

func (r *Reference[O]) GetValue() *O {
	obj := (*object[O])(unsafe.Pointer(r.address))
	if !obj.nextFree.IsNil() {
		panic(fmt.Errorf("attempted to get freed object %v", r))
	}
	return &obj.value
}

func (r *Reference[O]) getObject() *object[O] {
	obj := (*object[O])(unsafe.Pointer(r.address))
	return obj
}

func (r *Reference[O]) IsNil() bool {
	return r.address == 0
}
