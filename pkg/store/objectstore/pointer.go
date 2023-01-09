package objectstore

type Pointer[O any] struct {
	chunk  uint32
	offset uint32
}

func (p *Pointer[O]) IsNil() bool {
	return p.chunk == 0 && p.offset == 0
}
