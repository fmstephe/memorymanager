package quadtree

import (
	"fmt"

	"github.com/fmstephe/location-system/pkg/store/linkedlist"
	"github.com/fmstephe/location-system/pkg/store/objectstore"
)

// Returns a new empty quadtree.Tree whose View extends from
// leftX to rightX across the x axis and
// topY down to bottomY along the y axis
// leftX < rightX
// topY < bottomY
func NewQuadTree[T any](view View) Tree[T] {
	return newRoot[T](view)
}

// A point with a list of stored elements
type point[T any] struct {
	x, y float64
	list linkedlist.List[T]
}

// Indicates whether a point is isEmpty, i.e. uninitialised.  Because we never
// remove data from this quadtree, a point is only ever initialised when we add
// data to it.  This invariant will stop holding if this quadtree ever supports
// data deletion.
func (np *point[T]) isEmpty() bool {
	return np.list.IsEmpty()
}

// Indicates whether a point has the same x,y coords as those passed in
func (np *point[T]) sameLoc(x, y float64) bool {
	return np.x == x && np.y == y
}

func (np *point[T]) String() string {
	return fmt.Sprintf("(%v,%.3f,%.3f)", np.list, np.x, np.y)
}

const LEAF_SIZE = 16

// node structs make up the body of a quadtree.
// A node is either a leaf node, which contains actual data points.
// For a leaf all of the data points lie inside the bounds of this node's view.
//
// Or an internal node, internal nodes contain 4 children nodes.
// For an internal node all children nodes lie within the bounds of this node's view.
// The combined views of all the children nodes is the same as this node's view.
type node[T any] struct {
	view View

	// This is a count of the number of elements stored under this node. We
	// cache it to avoid needing to traverse the tree to answer this
	// question.
	cachedCount int64

	// A node is either a leaf, containing actual data, or an internal node
	// containing subtrees.
	isLeaf bool

	// Used if this node is a leaf
	ps [LEAF_SIZE]point[T]

	// Used if this node is not a leaf
	children [4]objectstore.Pointer[node[T]]
}

// Build an internal node, including allocating all of the children of this node.
// All of the child nodes are leaf nodes.
func makeNode[T any](view View, store *nodeStore[T]) objectstore.Pointer[node[T]] {
	nodePointer, newNode := store.allocNode(view)
	views := view.quarters()
	for i, view := range views {
		leafPointer := store.allocLeaf(view)
		newNode.children[i] = leafPointer
	}
	return nodePointer
}

// Inserts list into the single child subtree whose view contains (x,y)
func (n *node[T]) insert(x, y float64, list linkedlist.List[T], store *nodeStore[T]) {
	// We are adding an element to this node or one of its children, increment the count
	n.cachedCount++

	if n.isLeaf {
		// Node is a leaf - try to insert data directly into leaf
		for i := range n.ps {
			if n.ps[i].isEmpty() {
				n.ps[i].x = x
				n.ps[i].y = y
				n.ps[i].list = list
				return
			}
			if n.ps[i].sameLoc(x, y) {
				n.ps[i].list.Append(store.listStore, list)
				return
			}
		}

		// If we reach here then this leaf is full, convert to internal node
		n.convertToInternal(store)
		// After converting to internal node we fall down and execute internal node flow below
	}

	// Node is internal - find correct subtree to insert into
	for i := range n.children {
		childNode := store.getNode(n.children[i])
		if childNode.view.containsPoint(x, y) {
			childNode.insert(x, y, list, store)
			return
		}
	}
	panic("unreachable")
}

// Converts an existing leaf node to an internal node.  To do this we allocate
// a new set of leaf nodes and reinsert all of the data into these leaves.
func (n *node[T]) convertToInternal(store *nodeStore[T]) {
	n.isLeaf = false
	views := n.view.quarters()
	for i, view := range views {
		leafPointer := store.allocLeaf(view)
		n.children[i] = leafPointer
	}

	// re-insert data for the new leaves
	for i := range n.ps {
		p := &n.ps[i]
		x := p.x
		y := p.y
		list := p.list
		for i := range n.children {
			childNode := store.getNode(n.children[i])
			if childNode.view.containsPoint(x, y) {
				childNode.insert(x, y, list, store)
				break
			}
		}
	}
}

// Calls survey on each child subtree whose view overlaps with view
func (n *node[T]) survey(view View, fun func(x, y float64, data *T) bool, store *nodeStore[T]) bool {
	// Survey each point in this leaf
	if n.isLeaf {
		for i := range n.ps {
			p := &n.ps[i]
			if !p.isEmpty() && view.containsPoint(p.x, p.y) {
				if !p.list.Survey(store.listStore, func(data *T) bool { return fun(p.x, p.y, data) }) {
					return false
				}
			}
		}
		return true
	}

	// Survey each subtree
	for _, p := range n.children {
		st := store.getNode(p)
		if view.overlaps(st.view) {
			if !st.survey(view, fun, store) {
				return false
			}
		}
	}
	return true
}

func (n *node[T]) count(view View, store *nodeStore[T]) int64 {
	// In the case that the counting view completely covers this node
	// Then we can just quickly return the cached count
	if view.containsView(n.view) {
		return n.cachedCount
	}

	// count individual leaf elements
	if n.isLeaf {
		counted := int64(0)
		for i := range n.ps {
			p := &n.ps[i]
			if !p.isEmpty() && view.containsPoint(p.x, p.y) {
				// Visit all the elements stored here and count them
				p.list.Survey(store.listStore, func(_ *T) bool { counted++; return true })
			}
		}
		return counted
	}

	// Collect the count of the subtrees
	counted := int64(0)
	for _, p := range n.children {
		st := store.getNode(p)
		if view.overlaps(st.view) {
			counted += st.count(view, store)
		}
	}
	return counted
}

// Returns a human friendly string representing this node, including its children.
func (n *node[T]) String() string {
	// TODO
	return "TODO"
	//return "<" + n.view.String() + "-\n" + n.children[0].String() + ", \n" + n.children[1].String() + ", \n" + n.children[2].String() + ", \n" + n.children[3].String() + ">"
}

// Each tree has a single root.
// The root is responsible for:
//   - Implementing the quadtree public interface T.
//   - Allocating and recycling leaf and node elements.
type root[T any] struct {
	store       *nodeStore[T]
	rootPointer objectstore.Pointer[node[T]]
	view        View
}

// Returns a new root ready for use as an empty quadtree
//
// A root node is initialised and the tree is ready for service.
func newRoot[T any](view View) *root[T] {
	store := newTreeStore[T]()
	st := makeNode[T](view, store)
	return &root[T]{
		store:       store,
		rootPointer: st,
		view:        view,
	}
}

// Inserts data into this tree
func (r *root[T]) Insert(x, y float64, data T) error {
	if !r.view.containsPoint(x, y) {
		return fmt.Errorf("cannot insert x(%f) y(%f) into view %s", x, y, r.view)
	}
	list := r.store.newList(data)
	st := r.store.getNode(r.rootPointer)
	st.insert(x, y, list, r.store)
	return nil
}

// Applies fun to every element occurring within view in this tree
func (r *root[T]) Survey(view View, fun func(x, y float64, data *T) bool) {
	st := r.store.getNode(r.rootPointer)
	st.survey(view, fun, r.store)
}

// Applies fun to every element occurring within view in this tree
func (r *root[T]) Count(view View) int64 {
	st := r.store.getNode(r.rootPointer)
	return st.count(view, r.store)
}

// Returns the View for this tree
func (r *root[T]) View() View {
	return r.view
}

func (r *root[T]) String() string {
	st := r.store.getNode(r.rootPointer)
	return st.String()
}
