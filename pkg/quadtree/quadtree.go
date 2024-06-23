package quadtree

import (
	"fmt"

	"github.com/fmstephe/location-system/pkg/store/offheap"
)

// This struct is the exported root of a quad tree
type Tree[T any] struct {
	store         *nodeStore[T]
	treeReference offheap.RefObject[node[T]]
	view          View
}

// Returns a new Tree ready for use as an empty quadtree
//
// A Tree node is initialised and the tree is ready for service.
func NewTree[T any](view View) *Tree[T] {
	store := newTreeStore[T]()
	st := makeNode[T](view, store)
	return &Tree[T]{
		store:         store,
		treeReference: st,
		view:          view,
	}
}

// Inserts data into this tree
func (r *Tree[T]) Insert(x, y float64, data T) error {
	if !r.view.containsPoint(x, y) {
		return fmt.Errorf("cannot insert x(%f) y(%f) into view %s", x, y, r.view)
	}
	list := r.store.newList(data)
	st := r.treeReference.Value()
	st.insert(x, y, list, r.store)
	return nil
}

// Applies fun to every element occurring within view in this tree
func (r *Tree[T]) Survey(view View, fun func(x, y float64, data *T) bool) {
	st := r.treeReference.Value()
	st.survey(view, fun, r.store)
}

// Applies fun to every element occurring within view in this tree
func (r *Tree[T]) Count(view View) int64 {
	st := r.treeReference.Value()
	return st.count(view, r.store)
}

// Returns the View for this tree
func (r *Tree[T]) View() View {
	return r.view
}

func (r *Tree[T]) String() string {
	st := r.treeReference.Value()
	return st.String()
}
