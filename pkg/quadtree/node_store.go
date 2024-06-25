package quadtree

import (
	"github.com/fmstephe/location-system/pkg/store/offheap"
)

type nodeStore[T any] struct {
	nodes *offheap.Store
}

func newTreeStore[T any]() *nodeStore[T] {
	return &nodeStore[T]{
		nodes: offheap.New(),
	}
}

func (s *nodeStore[T]) allocNode(view View) (offheap.RefObject[node[T]], *node[T]) {
	r := offheap.AllocObject[node[T]](s.nodes)
	newNode := r.Value()
	newNode.view = view
	newNode.isLeaf = false
	return r, newNode
}

func (s *nodeStore[T]) allocLeaf(view View) offheap.RefObject[node[T]] {
	r := offheap.AllocObject[node[T]](s.nodes)
	newLeaf := r.Value()
	newLeaf.view = view
	newLeaf.isLeaf = true
	return r
}

func (s *nodeStore[T]) newSlice(data T) offheap.RefSlice[T] {
	slc := offheap.AllocSlice[T](s.nodes, 1, 1)
	slc.Value()[0] = data
	return slc
}
