package quadtree

import (
	"github.com/fmstephe/location-system/pkg/store/linkedlist"
	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

type nodeStore[T any] struct {
	nodes     *objectstore.Store[node[T]]
	listStore *linkedlist.Store[T]
}

func newTreeStore[T any]() *nodeStore[T] {
	return &nodeStore[T]{
		nodes: objectstore.New[node[T]](),
		//elems: objectstore.New[elem[T]](),
		listStore: linkedlist.New[T](),
	}
}

func (s *nodeStore[T]) newNode(view View) (objectstore.Pointer[node[T]], *node[T]) {
	p, newNode := s.nodes.Alloc()
	newNode.view = view
	newNode.isLeaf = false
	return p, newNode
}

func (s *nodeStore[T]) newLeaf(view View) objectstore.Pointer[node[T]] {
	p, newLeaf := s.nodes.Alloc()
	newLeaf.view = view
	newLeaf.isLeaf = true
	return p
}

func (s *nodeStore[T]) getNode(p objectstore.Pointer[node[T]]) *node[T] {
	return s.nodes.Get(p)
}

func (s *nodeStore[T]) newElem(data T) linkedlist.List[T] {
	list := s.listStore.NewList()
	dataP := list.Insert(s.listStore)
	*dataP = data
	return list
}
