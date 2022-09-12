package lowgc_quadtree

import "github.com/fmstephe/location-system/pkg/store"

type nodeStore[T any] struct {
	nodes *store.ObjectStore[node[T]]
	elems *store.ObjectStore[elem[T]]
}

func newTreeStore[T any]() *nodeStore[T] {
	return &nodeStore[T]{
		nodes: store.NewObjectStore[node[T]](),
		elems: store.NewObjectStore[elem[T]](),
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

func (s *nodeStore[T]) getNode(p store.ObjectPointer[node[T]]) *node[T] {
	return s.nodes.Get(p)
}

type elem[T any] struct {
	// Linked list fields
	next store.ObjectPointer[elem[T]]
	prev store.ObjectPointer[elem[T]]

	// Actual data
	data T
}

func (s *nodeStore[T]) newElem(data T) store.ObjectPointer[elem[T]] {
	p, newE := s.elems.New()
	// new element points to itself in a cycle
	newE.next = p
	newE.prev = p
	newE.data = data
	return p
}

func (s *nodeStore[T]) attachData(targetP, attachP store.ObjectPointer[elem[T]]) {
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

func (s *nodeStore[T]) survey(p store.ObjectPointer[elem[T]], fun func(e T)) {
	e := s.elems.Get(p)
	fun(e.data)

	// Follow through the linked list until we return to head
	next := e.next
	for next != p {
		e := s.elems.Get(next)
		fun(e.data)
		next = e.next
	}
}
