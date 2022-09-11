package lowgc_quadtree

import (
	"fmt"

	"github.com/fmstephe/location-system/pkg/store"
)

// Returns a new empty QuadTree whose View extends from
// leftX to rightX across the x axis and
// topY down to bottomY along the y axis
// leftX < rightX
// topY < bottomY
func NewQuadTree[K any](leftX, rightX, topY, bottomY float64) T[K] {
	var newView = NewView(leftX, rightX, topY, bottomY)
	return newRoot[K](newView)
}

// A point with a slice of stored elements
type vpoint[K any] struct {
	x, y  float64
	elems []K
}

// Indicates whether a vpoint is zeroed, i.e. uninitialised
func (np *vpoint[K]) zeroed() bool {
	return np.elems == nil
}

// Resets this vpoint back to its uninitialised state
func (np *vpoint[K]) zeroOut() {
	np.elems = nil
}

// Indicates whether a vpoint has the same x,y coords as those passed in
func (np *vpoint[K]) sameLoc(x, y float64) bool {
	return np.x == x && np.y == y
}

func (np *vpoint[K]) String() string {
	return fmt.Sprintf("(%v,%.3f,%.3f)", np.elems, np.x, np.y)
}

const LEAF_SIZE = 16

// A node struct implements the subtree interface.
// A node is the intermediate, non-leaf, storage structure for a
// quadtree.
// It contains a View, indicating the rectangular area this node covers.
// Each subtree will have a view containing one of four quarters of
// this node's view. Every subtree is guaranteed to be non-nil and
// may be either a node or a leaf struct.
type node[K any] struct {
	view View

	isLeaf bool
	// Used if this node is a leaf
	ps [LEAF_SIZE]vpoint[K]

	// Used if this node is not a leaf
	children [4]store.ObjectPointer[node[K]]
}

func makeNode[K any](view View, store *treeStore[K]) store.ObjectPointer[node[K]] {
	nodePointer, newNode := store.newNode(view)
	views := view.quarters()
	for i, view := range views {
		leafPointer := store.newLeaf(view)
		newNode.children[i] = leafPointer
	}
	return nodePointer
}

// Inserts elems into the single child subtree whose view contains (x,y)
func (n *node[K]) insert(x, y float64, elems []K, store *treeStore[K]) {
	if n.isLeaf {
		// Node is a leaf - try to insert data directly into leaf
		for i := range n.ps {
			if n.ps[i].zeroed() {
				n.ps[i].x = x
				n.ps[i].y = y
				n.ps[i].elems = elems
				return
			}
			if n.ps[i].sameLoc(x, y) {
				n.ps[i].elems = append(n.ps[i].elems, elems...)
				return
			}
		}

		// If we reach here then this leaf is full, convert to internal node
		n.convertToInternal(store)
		// After converting to internal node we fall down and execute internal node flow below
	}
	// Node is internal - find correct subtree to insert into

	for i := range n.children {
		childNode := store.get(n.children[i])
		if childNode.view.contains(x, y) {
			childNode.insert(x, y, elems, store)
			return
		}
	}
	panic("unreachable")
}

func (n *node[K]) convertToInternal(store *treeStore[K]) {
	n.isLeaf = false
	views := n.view.quarters()
	for i, view := range views {
		leafPointer := store.newLeaf(view)
		n.children[i] = leafPointer
	}

	// re-insert data for the new leaves
	for i := range n.ps {
		p := &n.ps[i]
		n.insert(p.x, p.y, p.elems, store)
	}
}

// Calls survey on each child subtree whose view overlaps with vs
func (n *node[K]) survey(view View, fun func(x, y float64, e K), store *treeStore[K]) {
	if n.isLeaf {
		for i := range n.ps {
			p := &n.ps[i]
			if !p.zeroed() && view.contains(p.x, p.y) {
				for i := range p.elems {
					fun(p.x, p.y, p.elems[i])
				}
			}
		}
	} else {
		for _, p := range n.children {
			st := store.get(p)
			if view.overlaps(st.view) {
				st.survey(view, fun, store)
			}
		}
	}
}

// Returns the View for this node
func (n *node[K]) View() View {
	return n.view
}

// Sets the view for this node
func (n *node[K]) setView(view *View) {
	n.view = *view
}

// Returns a human friendly string representing this node, including its children.
func (n *node[K]) String() string {
	// TODO
	return "TODO"
	//return "<" + n.view.String() + "-\n" + n.children[0].String() + ", \n" + n.children[1].String() + ", \n" + n.children[2].String() + ", \n" + n.children[3].String() + ">"
}

// Each tree has a single root.
// The root is responsible for:
//   - Implementing the quadtree public interface T.
//   - Allocating and recycling leaf and node elements
type root[K any] struct {
	store       *treeStore[K]
	rootPointer store.ObjectPointer[node[K]]
	view        View
}

// Returns a new root ready for use as an empty quadtree
//
// A root node is initialised and the tree is ready for service.
func newRoot[K any](view View) *root[K] {
	store := newTreeStore[K]()
	st := makeNode[K](view, store)
	return &root[K]{
		store:       store,
		rootPointer: st,
		view:        view,
	}
}

// Inserts the value nval into this tree
func (r *root[K]) Insert(x, y float64, nval K) {
	elems := make([]K, 1, 1)
	elems[0] = nval
	st := r.store.get(r.rootPointer)
	st.insert(x, y, elems, r.store)
}

// Applies fun to every element occurring within view in this tree
func (r *root[K]) Survey(view View, fun func(x, y float64, e K)) {
	st := r.store.get(r.rootPointer)
	st.survey(view, fun, r.store)
}

// Returns the View for this tree
func (r *root[K]) View() View {
	return r.view
}

func (r *root[K]) String() string {
	st := r.store.get(r.rootPointer)
	return st.String()
}
