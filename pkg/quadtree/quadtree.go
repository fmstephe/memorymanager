package quadtree

import (
	"fmt"
)

// Private interface for quadtree nodes. Implemented by both node and leaf.
type subtree[K any] interface {
	//
	insert(x, y float64, elems []K, p *subtree[K])
	//
	survey(view View, fun func(x, y float64, e K))
	//
	View() View
	//
	setView(view *View)
	//
	String() string
}

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

// A leaf struct implements the interface subtree. Like a node (see below),
// a leaf contains a View defining the rectangular area in which each vpoint
// could legally be located. A leaf struct may contain up to LEAF_SIZE non-zeroed
// vpoints.
// If any vpoint is non-empty, return false to zeroed(), then all vpoints
// of lesser index are also non-empty i.e. if ps[3] is non-empty then so
// are ps[2], ps[1] and ps[0], while ps[4] or greater have no such constraints.
// The vpoints are not ordered in any way with respect to their geometric locations.
type leaf[K any] struct {
	view View
	ps   [LEAF_SIZE]vpoint[K]
}

// Inserts each of the elements in elems into this leaf. There are three
// NB: We don't check that (x,y) is contained by this leaf's view, we rely of the parent
// node to ensure this.
// cases.
// If:			We find a non-empty vpoint at the exact location (x,y)
//   - Append elems to this vpoint
//
// Else-If:		We find an empty vpoint available
//   - Append elems to this vpoint and set the vpoint's location to (x,y)
//
// Else:		This leaf has overflowed
//   - Replace this leaf with an intermediate node and re-allocate
//     all of the elements in this leaf as well as those in elems into
//     the new node
func (l *leaf[K]) insert(x, y float64, elems []K, inPtr *subtree[K]) {
	for i := range l.ps {
		if l.ps[i].zeroed() {
			l.ps[i].x = x
			l.ps[i].y = y
			l.ps[i].elems = elems
			return
		}
		if l.ps[i].sameLoc(x, y) {
			l.ps[i].elems = append(l.ps[i].elems, elems...)
			return
		}
	}
	// This leaf is full we need to create an intermediary node to divide it up
	newIntNode[K](x, y, elems, inPtr, l)
}

// This function creates a new node and adds all of the elements contained in l to it,
// plus the new elements in elems. The pointer which previously pointed to l is
// pointed at the new node.
func newIntNode[K any](x, y float64, elems []K, inPtr *subtree[K], l *leaf[K]) {
	intNode := newNode[K](l.view)
	for _, p := range l.ps {
		intNode.insert(p.x, p.y, p.elems, nil) // Does not require an inPtr param as we are passing into a *node
	}
	intNode.insert(x, y, elems, nil) // Does not require an inPtr param as we are passing into a *node
	*inPtr = intNode                 // Redirect the old leaf's reference to this intermediate node
}

func newNode[K any](view View) *node[K] {
	n := &node[K]{view: view}
	v0, v1, v2, v3 := view.quarters()
	n.children[0] = &leaf[K]{view: v0}
	n.children[1] = &leaf[K]{view: v1}
	n.children[2] = &leaf[K]{view: v2}
	n.children[3] = &leaf[K]{view: v3}
	return n
}

// Applies fun to each of the elements contained in this leaf
// which appear within view.
func (l *leaf[K]) survey(view View, fun func(x, y float64, e K)) {
	for i := range l.ps {
		p := &l.ps[i]
		if !p.zeroed() && view.contains(p.x, p.y) {
			for i := range p.elems {
				fun(p.x, p.y, p.elems[i])
			}
		}
	}
}

// Restores the leaf invariant that "if any vpoint is non-empty, then all vpoints
// of lesser index are also non-empty" by rearranging the elements of ps.
func restoreOrder[K any](ps *[LEAF_SIZE]vpoint[K]) {
	for i := range ps {
		if ps[i].zeroed() {
			for j := i + 1; j < len(ps); j++ {
				if !ps[j].zeroed() {
					ps[i] = ps[j]
					ps[j].zeroOut()
					break
				}
			}
		}
	}
}

// Returns a pointer to the View of this leaf
func (l *leaf[K]) View() View {
	return l.view
}

// Sets the view for this leaf
func (l *leaf[K]) setView(view *View) {
	l.view = *view
}

// Returns a human friendly string representation of this leaf
func (l *leaf[K]) String() string {
	var str = l.view.String()
	for _, p := range l.ps {
		str += p.String()
	}
	return str
}

// A node struct implements the subtree interface.
// A node is the intermediate, non-leaf, storage structure for a
// quadtree.
// It contains a View, indicating the rectangular area this node covers.
// Each subtree will have a view containing one of four quarters of
// this node's view. Every subtree is guaranteed to be non-nil and
// may be either a node or a leaf struct.
type node[K any] struct {
	view     View
	children [4]subtree[K]
}

// Inserts elems into the single child subtree whose view contains (x,y)
func (n *node[K]) insert(x, y float64, elems []K, _ *subtree[K]) {
	for i := range n.children {
		view := n.children[i].View()
		if view.contains(x, y) {
			n.children[i].insert(x, y, elems, &n.children[i])
		}
	}
}

// Calls survey on each child subtree whose view overlaps with vs
func (n *node[K]) survey(view View, fun func(x, y float64, e K)) {
	for i := range n.children {
		child := n.children[i]
		if view.overlaps(child.View()) {
			n.children[i].survey(view, fun)
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
	return "<" + n.view.String() + "-\n" + n.children[0].String() + ", \n" + n.children[1].String() + ", \n" + n.children[2].String() + ", \n" + n.children[3].String() + ">"
}

// Each tree has a single root.
// The root is responsible for:
//   - Implementing the quadtree public interface T.
//   - Allocating and recycling leaf and node elements
type root[K any] struct {
	rootNode subtree[K]
}

// Returns a new root ready for use as an empty quadtree
//
// A root node is initialised and the tree is ready for service.
func newRoot[K any](view View) *root[K] {
	return &root[K]{
		rootNode: newNode[K](view),
	}
}

// Fills the array provided with new leaves each occupying
// a quarter of the view provided.
func (r *root[K]) newLeaves(view View, leaves *[4]subtree[K]) {
	v0, v1, v2, v3 := view.quarters()
	vs := []View{v0, v1, v2, v3}
	for i := range leaves {
		leaves[i] = &leaf[K]{view: vs[i]}
	}
}

// Inserts the value nval into this tree
func (r *root[K]) Insert(x, y float64, nval K) {
	elems := make([]K, 1, 1)
	elems[0] = nval
	r.rootNode.insert(x, y, elems, nil)
}

// Applies fun to every element occurring within view in this tree
func (r *root[K]) Survey(view View, fun func(x, y float64, e K)) {
	r.rootNode.survey(view, fun)
}

// Returns the View for this tree
func (r *root[K]) View() View {
	return r.rootNode.View()
}

func (r *root[K]) String() string {
	return r.rootNode.String()
}
