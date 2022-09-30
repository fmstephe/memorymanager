package lowgc_quadtree

// Public interface for quadtrees.
type Tree[K any] interface {
	View() View
	// Inserts e into this quadtree at point (x,y)
	Insert(x, y float64, e K) error
	// Applies fun to every element in this quadtree that lies within view
	// If fun returns false, then the surveying terminates
	Survey(view View, fun func(x, y float64, e K) bool)
	// Returns the number of elements found within the view
	Count(view View) int64
	// Provides a human readable (as far as possible) string representation of this tree
	String() string
}
