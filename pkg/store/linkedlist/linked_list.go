package linkedlist

import (
	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

// A node contains the list data as well as the forward and previous References
// to next nodes in the list. The next and prev References are never nil. If the
// list contains only one node then the next and prev References point back to
// this node.
type node[O any] struct {
	data O
	next objectstore.Reference[node[O]]
	prev objectstore.Reference[node[O]]
}

// Convenience method to get a pointer to the embedded data.
func (n *node[O]) getData() *O {
	return &n.data
}

// The store for linked lists. It is used to create a new list, but must also
// be passed into any method which operates on linkedlists.
type Store[O any] struct {
	nodeStore *objectstore.Store
}

// Creates a new Store.
func New[O any]() *Store[O] {
	return &Store[O]{
		nodeStore: objectstore.New(),
	}
}

// Creates a new empty list
// l.getReference().IsNil() will return true
// This method isn't strictly necessary. The zero value of a List[O] behaves
// exactly the same as one created by a Store[O]. However we include the method
// because it makes code easier to follow.
func (s *Store[O]) NewList() List[O] {
	return List[O]{}
}

// A List is simply a Reference to a node
type List[O any] objectstore.Reference[node[O]]

// casts a List to the raw Reference
func (l *List[O]) getReference() objectstore.Reference[node[O]] {
	return objectstore.Reference[node[O]](*l)
}

// sets the value of a List using a raw Reference value
func (l *List[O]) setReference(r objectstore.Reference[node[O]]) {
	*l = List[O](r)
}

// Pushes a single new node into the first position of an existing list. The
// list may be empty. The list node is allocated internally and a pointer to
// the embedded data is returned. The embedded data can then be mutated via
// this pointer.
func (l *List[O]) PushHead(store *Store[O]) *O {
	newR, newNode := objectstore.Alloc[node[O]](store.nodeStore)
	l.pushTail(store, newR, newNode)
	l.setReference(newR)
	return newNode.getData()
}

// Pushes a single new node into the last position of an existing list. The
// list may be empty. The list node is allocated internally and a pointer to
// the embedded data is returned. The embedded data can then be mutated via
// this pointer.
func (l *List[O]) PushTail(store *Store[O]) *O {
	newR, newNode := objectstore.Alloc[node[O]](store.nodeStore)
	l.pushTail(store, newR, newNode)
	return newNode.getData()
}

func (l *List[O]) pushTail(store *Store[O], newR objectstore.Reference[node[O]], newNode *node[O]) {
	firstR := l.getReference()

	// If we are inserting into an empty list, then make it point to itself
	// and directly update l
	if firstR.IsNil() {
		newNode.next = newR
		newNode.prev = newR
		l.setReference(newR)
		return
	}

	// Get the first and last nodes in the linked list
	firstNode := firstR.GetValue()
	lastNode := firstNode.prev.GetValue()

	// Make the last node point to the new node
	lastNode.next = newR
	newNode.prev = firstNode.prev

	// Make the new node point to the first node
	newNode.next = firstR
	firstNode.prev = newR
}

func (l *List[O]) PeakHead(store *Store[O]) *O {
	firstR := l.getReference()

	firstNode := firstR.GetValue()
	return firstNode.getData()
}

func (l *List[O]) PeakTail(store *Store[O]) *O {
	firstR := l.getReference()

	firstNode := firstR.GetValue()
	lastNode := firstNode.prev.GetValue()
	return lastNode.getData()
}

func (l *List[O]) RemoveHead(store *Store[O]) {
	l.remove(store, l.getReference())
}

func (l *List[O]) RemoveTail(store *Store[O]) {
	ref := l.getReference()
	origin := ref.GetValue()
	l.remove(store, origin.prev)
}

func (l *List[O]) remove(store *Store[O], r objectstore.Reference[node[O]]) {
	n := r.GetValue()
	if n.prev == r && n.next == r {
		// There is only one element in this list, now we empty it
		*l = List[O]{}

		// Free the removed node
		objectstore.Free(store.nodeStore, r)
		return
	}

	// Connect the previous and next nodes to each other
	prev := n.prev.GetValue()
	next := n.next.GetValue()
	prev.next = n.next
	next.prev = n.prev

	// If the removed node is the head of this list, point the list to next
	if r == l.getReference() {
		*l = List[O](n.next)
	}

	// Free the removed node
	objectstore.Free(store.nodeStore, r)
}

// Adds the nodes in attach to l. After this method is called attach should
// no longer be used.
func (l *List[O]) Append(store *Store[O], attach List[O]) {
	if attach.IsEmpty() {
		// There is nothing useful in attach - do nothing
		return
	}
	if l.IsEmpty() {
		// If l is empty then we just set it to point at attach
		*l = attach
		return
	}

	// Get nodes in this linked list
	lR := l.getReference()
	lElem := lR.GetValue()
	lPrev := lElem.prev
	lPrevElem := lPrev.GetValue()

	// Get nodes in the attaching linked list
	attachR := attach.getReference()
	attachElem := attachR.GetValue()
	attachPrev := attachElem.prev
	attachPrevElem := attachPrev.GetValue()

	// Connect end of h linked list to the start of attach linked list
	lPrevElem.next = attachR
	attachElem.prev = lPrev

	// Connect start of h linked list to the end of attach linked list
	attachPrevElem.next = lR
	lElem.prev = attachPrev
}

// This method iterates over every node in the list. For each node the function
// is called with a pointer to the embedded data for that node. It is possible,
// and idiomatic, to mutate the embedded data in the list via these pointers.
//
// The embedded data pointed to in each function call is owned by the list.
// Client code probably should not retain pointers outside the scope of the
// Survey method call. It is always acceptable to retain a copy of the embedded
// data value.
func (l *List[O]) Survey(store *Store[O], fun func(o *O) bool) bool {
	if l.IsEmpty() {
		// If p is nil, the list is empty, we have successfully
		// surveyed it
		return true
	}

	// Follow through the linked list until we return to head
	origin := l.getReference()
	current := origin
	for {
		n := current.GetValue()
		if !fun(n.getData()) {
			return false
		}
		current = n.next
		if current == origin {
			return true
		}
	}
}

// This method iterates over every node in the list. For each node the function
// is called with a pointer to the embedded data for that node. It is possible,
// although a bit unusual, to mutate the embedded data in the list via these
// pointers. For each function call, if the return value is true then the node
// will be removed from the list. Removing an element from a list causes the
// node to be freed and its memory returned to the store.
//
// The embedded data pointed to in each function call is owned by the list.
// Client code probably should not retain pointers outside the scope of the
// Survey method call. It is always acceptable to retain a copy of the embedded
// data value.
func (l *List[O]) Filter(store *Store[O], pred func(o *O) bool) {
	// Follow through the linked list until we return to origin
	origin := l.getReference()
	if origin.IsNil() {
		// If p is nil, the list is empty, we have successfully
		// filtered it
		return
	}
	defer func() {
		// The final act is to update the header to point to a valid
		// node in the list. If we filtered out the node pointed to by
		// h, we will have to modify it to point a node which is still
		// in the list. If we filtered all nodes from the list origin will
		// be a nil getReference.
		*l = List[O](origin)
	}()

	current := origin
	for {
		n := current.GetValue()

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
		objectstore.Free(store.nodeStore, current)

		if n.prev == current && n.next == current {
			// Special case where we are filtering the last node
			// nil out origin and return
			if current != origin {
				panic("We assumed that this case could only be hit when origin == current")
			}
			// Make origin a nil Reference
			origin = objectstore.Reference[node[O]]{}
			return
		}

		// Link prev and next together
		prevE := n.prev.GetValue()
		nextE := n.next.GetValue()
		prevE.next = n.next
		nextE.prev = n.prev

		// If we just filtered out origin - set origin to next (or we can
		// never exit this loop)
		if current == origin {
			origin = n.next
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

// Indicates whether this list is empty. This function has two advantages over Len, if you only care in whether a list is empty or not.
//
// 1: Because it does not actually look at any list data, no *Store argument is needed.
// 2: It is fast, because it doesn't need to iterate over the list
func (l *List[O]) IsEmpty() bool {
	r := l.getReference()
	return r.IsNil()
}
