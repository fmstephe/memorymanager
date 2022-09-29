package lowgc_quadtree

// Public interface for quadtrees.
type Tree[K any] interface {
	View() View
	// Inserts e into this quadtree at point (x,y)
	Insert(x, y float64, e K) error
	// Applies fun to every element in this quadtree that lies within any view in views
	Survey(view View, fun func(x, y float64, e K))
	// Provides a human readable (as far as possible) string representation of this tree
	String() string
}
