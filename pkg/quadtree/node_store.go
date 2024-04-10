package quadtree

import (
	"github.com/fmstephe/location-system/pkg/store/linkedlist"
	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

type nodeStore[T any] struct {
	nodes     *objectstore.Store
	listStore *linkedlist.Store[T]
}

func newTreeStore[T any]() *nodeStore[T] {
	return &nodeStore[T]{
		nodes:     objectstore.New(),
		listStore: linkedlist.New[T](),
	}
}

func (s *nodeStore[T]) allocNode(view View) (objectstore.Reference[node[T]], *node[T]) {
	r, newNode := objectstore.Alloc[node[T]](s.nodes)
	newNode.view = view
	newNode.isLeaf = false
	return r, newNode
}

func (s *nodeStore[T]) allocLeaf(view View) objectstore.Reference[node[T]] {
	r, newLeaf := objectstore.Alloc[node[T]](s.nodes)
	newLeaf.view = view
	newLeaf.isLeaf = true
	return r
}

func (s *nodeStore[T]) newList(data T) linkedlist.List[T] {
	list := s.listStore.NewList()
	dataP := list.PushTail(s.listStore)
	*dataP = data
	return list
}
