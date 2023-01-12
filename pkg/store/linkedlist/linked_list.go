package linkedlist

import (
	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

// A node contains the list data as well as the forward and previous getPointers
// to next nodes in the list.  The next and prev getPointers are never nil. If the
// list contains only one node then the next and prev getPointers point back to
// themselves.
type node[O any] struct {
	data O
	next objectstore.Pointer[node[O]]
	prev objectstore.Pointer[node[O]]
}

// Convenience method to get a getPointer to the embedded data.
func (n *node[O]) getData() *O {
	return &n.data
}

// The store for linked lists. It is used to create a new list, but must also
// be passed into any method which operates on linkedlists.
type Store[O any] struct {
	nodeStore *objectstore.Store[node[O]]
}

// Creates a new Store.
func New[O any]() *Store[O] {
	return &Store[O]{
		nodeStore: objectstore.New[node[O]](),
	}
}

// Creates a new empty list
// l.getPointer().IsNil() will return true
func (s *Store[O]) NewList() *List[O] {
	return &List[O]{}
}

type List[O any] objectstore.Pointer[node[O]]

func (l *List[O]) getPointer() objectstore.Pointer[node[O]] {
	return objectstore.Pointer[node[O]](*l)
}

func (l *List[O]) setPointer(p objectstore.Pointer[node[O]]) {
	*l = List[O](p)
}

func (l *List[O]) Insert(store *Store[O]) *O {
	firstP := l.getPointer()
	newP, newNode := store.nodeStore.Alloc()

	// If we are inserting into an empty list, then make it point to itself
	// and directly update l
	if firstP.IsNil() {
		newNode.next = newP
		newNode.prev = newP
		l.setPointer(newP)
		return newNode.getData()
	}

	// Get the first and last elements in the linked list
	firstNode := store.nodeStore.Get(firstP)
	lastNode := store.nodeStore.Get(firstNode.prev)

	// Make the last node point to the new node
	lastNode.next = newP
	newNode.prev = firstNode.prev

	// Make the new node point to the first node
	newNode.next = firstP
	firstNode.prev = newP

	return newNode.getData()
}

// TODO add tests which terminate the survey early. Do we actually need this behaviour?
func (l *List[O]) Survey(store *Store[O], fun func(o *O) bool) bool {
	// Follow through the linked list until we return to head
	origin := l.getPointer()
	if origin.IsNil() {
		// If p is nil, the list is empty, we have successfully
		// surveyed it
		return true
	}
	current := origin
	for {
		n := store.nodeStore.Get(current)
		if !fun(n.getData()) {
			return false
		}
		current = n.next
		if current == origin {
			return true
		}
	}
}

// TODO add a suite of tests on lists of size 2 and 3. These two cases are specifically interesting because the sized 2 list has first and last element directly connected to each other, and the 3 sized list has an intermediate node. We will assume that lists larger than 3 will behave the same as we sized 3 lists - so we don't have to test an infinite number of lists. These tests should specifically track the values of individual elements added and removed.
func (l *List[O]) Filter(store *Store[O], pred func(o *O) bool) {
	// Follow through the linked list until we return to origin
	origin := l.getPointer()
	if origin.IsNil() {
		// If p is nil, the list is empty, we have successfully
		// filtered it
		return
	}
	defer func() {
		// The final act is to update the header to point to a valid
		// node in the list.  If we filtered out the node pointed to by
		// h, we will have to modify it to point a node which is still
		// in the list.  If we filtered all nodes from the list origin will
		// be a nil getPointer.
		*l = List[O](origin)
	}()

	current := origin
	for {
		n := store.nodeStore.Get(current)

		// Test if node should be filtered
		if pred(n.getData()) {
			// Don't filter node
			current = n.next

			// Check if we have looped around the list
			if current == origin {
				return
			}
			continue
		}

		// Filter current node
		store.nodeStore.Free(current)

		if n.prev == current && n.next == current {
			// Special case where we are filtering the last node
			// nil out origin and return
			if current != origin {
				panic("We assumed that this case could only be hit when origin == current")
			}
			// Make origin a nil getPointer
			origin = objectstore.Pointer[node[O]]{}
			return
		}

		// Link prev and next together
		prevE := store.nodeStore.Get(n.prev)
		nextE := store.nodeStore.Get(n.next)
		prevE.next = n.next
		nextE.prev = n.prev

		// If we just filtered out origin - set origin to prev (or we can
		// never exit this loop)
		if current == origin {
			origin = n.prev
		}
		current = n.next
	}
}

// Consider a way to do this which doesn't need to iterate over the whole list
func (l *List[O]) Len(store *Store[O]) int {
	count := 0

	l.Survey(store, func(_ *O) bool {
		count++
		return true
	})
	return count
}
