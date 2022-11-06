package quadtree

import "github.com/fmstephe/location-system/pkg/store/objectstore"

type nodeStore[T any] struct {
	nodes *objectstore.Store[node[T]]
	elems *objectstore.Store[elem[T]]
}

func newTreeStore[T any]() *nodeStore[T] {
	return &nodeStore[T]{
		nodes: objectstore.New[node[T]](),
		elems: objectstore.New[elem[T]](),
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

type elem[T any] struct {
	// Linked list fields
	next objectstore.Pointer[elem[T]]
	prev objectstore.Pointer[elem[T]]

	// Actual data
	data T
}

func (s *nodeStore[T]) newElem(data T) objectstore.Pointer[elem[T]] {
	p, newE := s.elems.Alloc()
	// new element points to itself in a cycle
	newE.next = p
	newE.prev = p
	newE.data = data
	return p
}

func (s *nodeStore[T]) attachData(targetP, attachP objectstore.Pointer[elem[T]]) {
	// Get elements in the target linked list
	targetElem := s.elems.Get(targetP)
	targetTailP := targetElem.prev
	targetTailElem := s.elems.Get(targetTailP)

	// Get elements in the attaching linked list
	attachElem := s.elems.Get(attachP)
	attachTailP := attachElem.prev
	attachTailElem := s.elems.Get(attachTailP)

	// Connect end of target linked list to the start of attach linked list
	targetTailElem.next = attachP
	attachElem.prev = targetTailP

	// Connect start of target linked list to the end of attach linked list
	attachTailElem.next = targetP
	targetElem.prev = attachTailP
}

func (s *nodeStore[T]) survey(p objectstore.Pointer[elem[T]], fun func(e T) bool) bool {
	e := s.elems.Get(p)
	if !fun(e.data) {
		return false
	}

	// Follow through the linked list until we return to head
	next := e.next
	for next != p {
		e := s.elems.Get(next)
		if !fun(e.data) {
			return false
		}
		next = e.next
	}
	return true
}
