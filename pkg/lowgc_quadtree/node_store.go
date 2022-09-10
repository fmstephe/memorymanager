package lowgc_quadtree

type treeStore[T any] struct {
	nodeStore objectStore[node[T]]
}

func newTreeStore[T any]() *treeStore[T] {
	return &treeStore[T]{
		nodeStore: newObjectStore[node[T]](),
	}
}

func (s *treeStore[T]) newNode(view View) (pointer, *node[T]) {
	p, newNode := s.nodeStore.newObject()
	newNode.view = view
	newNode.isLeaf = false
	return p, newNode
}

func (s *treeStore[T]) newLeaf(view View) pointer {
	p, newLeaf := s.nodeStore.newObject()
	newLeaf.view = view
	newLeaf.isLeaf = true
	return p
}

func (s *treeStore[T]) get(p pointer) *node[T] {
	return s.nodeStore.getObject(p)
}

type objectStore[O any] struct {
	// Immutable fields
	chunkSize int32

	offset  int32
	objects [][]O
}

type pointer struct {
	chunk  int32
	offset int32
}

func newObjectStore[O any]() objectStore[O] {
	chunkSize := int32(1024)
	// Initialise the first chunk
	objects := [][]O{make([]O, chunkSize)}
	return objectStore[O]{
		chunkSize: chunkSize,
		offset:    0,
		objects:   objects,
	}
}

func (s *objectStore[O]) newObject() (pointer, *O) {
	chunk := int32(len(s.objects) - 1)
	offset := s.offset
	s.offset++
	if s.offset == s.chunkSize {
		// Create a new chunk
		s.objects = append(s.objects, make([]O, s.chunkSize))
		s.offset = 0
	}
	return pointer{
		chunk:  chunk,
		offset: offset,
	}, &s.objects[chunk][offset]
}

func (s *objectStore[O]) getObject(p pointer) *O {
	return &s.objects[p.chunk][p.offset]
}
