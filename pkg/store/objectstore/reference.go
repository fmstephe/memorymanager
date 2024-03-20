package objectstore

type Reference[O any] struct {
	allocIdx uint64
}

func newReference[O any](allocIdx uint64) Reference[O] {
	return Reference[O]{
		allocIdx: allocIdx + 1,
	}
}

func (r *Reference[O]) chunkAndOffset(chunkSize uint64) (chunkIdx uint64, offsetIdx uint64) {
	allocIdx := r.allocIdx - 1
	// TODO do some power of 2 work here, to eliminate all this division
	chunkIdx = allocIdx / chunkSize
	offsetIdx = allocIdx % chunkSize
	return chunkIdx, offsetIdx
}

func (r *Reference[O]) IsNil() bool {
	return r.allocIdx == 0
}
