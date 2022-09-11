package lowgc_quadtree

import "github.com/fmstephe/location-system/pkg/store"

type nodeStore[T any] struct {
	nodes *store.ObjectStore[node[T]]
}

func newTreeStore[T any]() *nodeStore[T] {
	return &nodeStore[T]{
		nodes: store.NewObjectStore[node[T]](),
	}
}

func (s *nodeStore[T]) newNode(view View) (store.ObjectPointer[node[T]], *node[T]) {
	p, newNode := s.nodes.New()
	newNode.view = view
	newNode.isLeaf = false
	return p, newNode
}

func (s *nodeStore[T]) newLeaf(view View) store.ObjectPointer[node[T]] {
	p, newLeaf := s.nodes.New()
	newLeaf.view = view
	newLeaf.isLeaf = true
	return p
}

func (s *nodeStore[T]) get(p store.ObjectPointer[node[T]]) *node[T] {
	return s.nodes.Get(p)
}
