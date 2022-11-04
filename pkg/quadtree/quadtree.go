package quadtree

import (
	"fmt"

	"github.com/fmstephe/location-system/pkg/store"
)

// Returns a new empty QuadTree whose View extends from
// leftX to rightX across the x axis and
// topY down to bottomY along the y axis
// leftX < rightX
// topY < bottomY
func NewQuadTree[K any](view View) Tree[K] {
	return newRoot[K](view)
}

// A point with a slice of stored elements
type vpoint[K any] struct {
	x, y   float64
	elemsP store.ObjectPointer[elem[K]]
}

// Indicates whether a vpoint is zeroed, i.e. uninitialised
func (np *vpoint[K]) zeroed() bool {
	return np.elemsP.IsNil()
}

// Indicates whether a vpoint has the same x,y coords as those passed in
func (np *vpoint[K]) sameLoc(x, y float64) bool {
	return np.x == x && np.y == y
}

func (np *vpoint[K]) String() string {
	return fmt.Sprintf("(%v,%.3f,%.3f)", np.elemsP, np.x, np.y)
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

	cachedCount int64

	isLeaf bool
	// Used if this node is a leaf
	ps [LEAF_SIZE]vpoint[K]

	// Used if this node is not a leaf
	children [4]store.ObjectPointer[node[K]]
}

func makeNode[K any](view View, store *nodeStore[K]) store.ObjectPointer[node[K]] {
	nodePointer, newNode := store.newNode(view)
	views := view.quarters()
	for i, view := range views {
		leafPointer := store.newLeaf(view)
		newNode.children[i] = leafPointer
	}
	return nodePointer
}

// Inserts elems into the single child subtree whose view contains (x,y)
func (n *node[K]) insert(x, y float64, elemsP store.ObjectPointer[elem[K]], store *nodeStore[K]) {
	// We are adding an element to this node or one of its children, increment the count
	n.cachedCount++

	if n.isLeaf {
		// Node is a leaf - try to insert data directly into leaf
		for i := range n.ps {
			if n.ps[i].zeroed() {
				n.ps[i].x = x
				n.ps[i].y = y
				n.ps[i].elemsP = elemsP
				return
			}
			if n.ps[i].sameLoc(x, y) {
				store.attachData(n.ps[i].elemsP, elemsP)
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
			childNode.insert(x, y, elemsP, store)
			return
		}
	}
	panic("unreachable")
}

func (n *node[K]) convertToInternal(store *nodeStore[K]) {
	n.isLeaf = false
	views := n.view.quarters()
	for i, view := range views {
		leafPointer := store.newLeaf(view)
		n.children[i] = leafPointer
	}

	// re-insert data for the new leaves
	for i := range n.ps {
		p := &n.ps[i]
		x := p.x
		y := p.y
		elemsP := p.elemsP
		for i := range n.children {
			childNode := store.getNode(n.children[i])
			if childNode.view.containsPoint(x, y) {
				childNode.insert(x, y, elemsP, store)
				break
			}
		}
	}
}

// Calls survey on each child subtree whose view overlaps with vs
func (n *node[K]) survey(view View, fun func(x, y float64, e K) bool, store *nodeStore[K]) bool {
	// Survey each point in this leaf
	if n.isLeaf {
		for i := range n.ps {
			p := &n.ps[i]
			if !p.zeroed() && view.containsPoint(p.x, p.y) {
				if !store.survey(p.elemsP, func(data K) bool { return fun(p.x, p.y, data) }) {
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

func (n *node[K]) count(view View, store *nodeStore[K]) int64 {
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
			if !p.zeroed() && view.containsPoint(p.x, p.y) {
				// Visit all the elements stored here and count them
				store.survey(p.elemsP, func(data K) bool { counted++; return true })
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
	store       *nodeStore[K]
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
func (r *root[K]) Insert(x, y float64, nval K) error {
	if !r.view.containsPoint(x, y) {
		return fmt.Errorf("cannot insert x(%f) y(%f) into view %s", x, y, r.view)
	}
	elemsP := r.store.newElem(nval)
	st := r.store.getNode(r.rootPointer)
	st.insert(x, y, elemsP, r.store)
	return nil
}

// Applies fun to every element occurring within view in this tree
func (r *root[K]) Survey(view View, fun func(x, y float64, e K) bool) {
	st := r.store.getNode(r.rootPointer)
	st.survey(view, fun, r.store)
}

// Applies fun to every element occurring within view in this tree
func (r *root[K]) Count(view View) int64 {
	st := r.store.getNode(r.rootPointer)
	return st.count(view, r.store)
}

// Returns the View for this tree
func (r *root[K]) View() View {
	return r.view
}

func (r *root[K]) String() string {
	st := r.store.getNode(r.rootPointer)
	return st.String()
}
