package linkedlist

import (
	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

type node[O any] struct {
	data O
	next objectstore.Pointer[node[O]]
	prev objectstore.Pointer[node[O]]
}

func (n *node[O]) getData() *O {
	return &n.data
}

type Store[O any] struct {
	nodeStore *objectstore.Store[node[O]]
}

func New[O any]() *Store[O] {
	return &Store[O]{
		nodeStore: objectstore.New[node[O]](),
	}
}

func (s *Store[O]) NewList() (*List[O], *O) {
	p, newNode := s.nodeStore.Alloc()
	newNode.next = p
	newNode.prev = p
	return (*List[O])(&p), newNode.getData()
}

type List[O any] objectstore.Pointer[node[O]]

func (l *List[O]) pointer() objectstore.Pointer[node[O]] {
	return objectstore.Pointer[node[O]](*l)
}

func (l *List[O]) Insert(store *Store[O]) *O {
	p, data := store.NewList()
	l.combine(store, p.pointer())
	return data
}

func (l *List[O]) combine(store *Store[O], attach objectstore.Pointer[node[O]]) {
	// Get Elementents in the headers linked list
	hElem := store.nodeStore.Get(l.pointer())
	hPrev := hElem.prev
	hPrevElem := store.nodeStore.Get(hPrev)

	// Get Elementents in the attaching linked list
	attachElem := store.nodeStore.Get(attach)
	attachPrev := attachElem.prev
	attachPrevElem := store.nodeStore.Get(attachPrev)

	// Connect end of h linked list to the start of attach linked list
	hPrevElem.next = attach
	attachElem.prev = hPrev

	// Connect start of h linked list to the end of attach linked list
	attachPrevElem.next = l.pointer()
	hElem.prev = attachPrev
}

func (l *List[O]) Survey(store *Store[O], fun func(o *O) bool) bool {
	// Follow through the linked list until we return to head
	p := l.pointer()
	if p.IsNil() {
		// If p is nil, the list is empty, we have successfully
		// surveyed it
		return true
	}
	current := p
	for {
		n := store.nodeStore.Get(current)
		if !fun(n.getData()) {
			return false
		}
		current = n.next
		if current == p {
			return true
		}
	}
}

func (l *List[O]) Filter(store *Store[O], pred func(o *O) bool) {
	// Follow through the linked list until we return to origin
	origin := l.pointer()
	defer func() {
		// The final act is to update the header to point to a valid
		// node in the list.  If we filtered out the node pointed to by
		// h, we will have to modify it to point a node which is still
		// in the list.  If we filtered all nodes from the list origin will
		// be a nil pointer.
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
			// Make origin a nil pointer
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

func (l *List[O]) Len(store *Store[O]) int {
	count := 0

	l.Survey(store, func(_ *O) bool {
		count++
		return true
	})
	return count
}
