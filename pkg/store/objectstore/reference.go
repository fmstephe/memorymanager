package objectstore

type Reference[O any] struct {
	chunk  uint32
	offset uint32
}

func (r *Reference[O]) IsNil() bool {
	return r.chunk == 0 && r.offset == 0
}
