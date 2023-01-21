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
// This method isn't strictly necessary. The zero value of a List[O] behaves
// exactly the same as one created by a Store[O]. However we include the method
// because it makes code easier to follow.
func (s *Store[O]) NewList() List[O] {
	return List[O]{}
}

// A List is simply a pointer to a node
type List[O any] objectstore.Pointer[node[O]]

// converts a List into the raw pointer value
func (l *List[O]) getPointer() objectstore.Pointer[node[O]] {
	return objectstore.Pointer[node[O]](*l)
}

// sets the value of a List using a raw pointer value
func (l *List[O]) setPointer(p objectstore.Pointer[node[O]]) {
	*l = List[O](p)
}

// Inserts a single new node into an existing list. The list may be empty.
// The list node is allocated internally and a pointer to the embedded data
// is returned.  The embedded data can then be mutated via this pointer.
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

	// Get the first and last nodes in the linked list
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
	lP := l.getPointer()
	hElem := store.nodeStore.Get(lP)
	hPrev := hElem.prev
	hPrevElem := store.nodeStore.Get(hPrev)

	// Get nodes in the attaching linked list
	attachP := attach.getPointer()
	attachElem := store.nodeStore.Get(attachP)
	attachPrev := attachElem.prev
	attachPrevElem := store.nodeStore.Get(attachPrev)

	// Connect end of h linked list to the start of attach linked list
	hPrevElem.next = attachP
	attachElem.prev = hPrev

	// Connect start of h linked list to the end of attach linked list
	attachPrevElem.next = lP
	hElem.prev = attachPrev
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
	origin := l.getPointer()
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
	p := l.getPointer()
	return p.IsNil()
}