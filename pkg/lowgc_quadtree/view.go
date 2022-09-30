package lowgc_quadtree

import (
	"fmt"
	"strconv"
)

// A View is a rectangle defined by four points, from two x coords and two y
// coords. It is acceptable for any of the coordinates to be negative values,
// so long as the invariants are respected.
// lx - left most x
// rx - right most x
// Invariant: lx <= rx
// ty - top most y
// by - bottom most y
// Invariant: ty >= by
// The rectangle is defined the four points;
// (lx,ty),(lx,by),(rx,ty),(rx,by)
// The zeroed View is a zero area plane at origin (0,0)
type View struct {
	lx float64
	rx float64
	ty float64
	by float64
}

// Returns a new View struct with the four
func NewView(lx, rx, ty, by float64) View {
	if rx < lx {
		panic(fmt.Sprintf("Cannot create view with inverted x coordinates. lx : %10.3f rx : %10.3f, ty : %10.3f by %10.3f", lx, rx, ty, by))
	}
	if ty < by {
		panic(fmt.Sprintf("Cannot create view with inverted y coordinates. lx : %10.3f rx : %10.3f, ty : %10.3f by %10.3f", lx, rx, ty, by))
	}
	return View{lx, rx, ty, by}
}

// Convenience function to get a View which captures the entire earth in Lon/Lat
// We use our X coordinates as longitudes west to east (-180...180)
// We use our Y coordinates as lattitudes north to south (90...-90)
func NewLongLatView() View {
	return View{
		// Leftmost x is all the way west
		lx: -180,
		// Rightmost x is all the way east
		rx: 180,
		// Top y is at the north pole, very north
		ty: 90,
		// Bottom y is at the south pole, very south
		by: -90,
	}
}

// Indicates whether this View contains the point (x,y)
func (v View) contains(x, y float64) bool {
	return x >= v.lx && x <= v.rx && y >= v.by && y <= v.ty
}

// Indicates whether any of the four edges
// of ov pass through v
func (v View) crossedBy(ov View) bool {
	if v.crossedVertically(ov.lx, ov.ty, ov.by) {
		return true
	}
	if v.crossedVertically(ov.rx, ov.ty, ov.by) {
		return true
	}
	if v.crossedHorizontally(ov.ty, ov.lx, ov.rx) {
		return true
	}
	if v.crossedHorizontally(ov.by, ov.lx, ov.rx) {
		return true
	}
	return false
}

// Indicates whether the line running vertically
// along x from ty down to by passes through v
// Invariant: ty >= by
func (v View) crossedVertically(x, ty, by float64) bool {
	if x < v.lx || x > v.rx {
		return false
	}
	if by > v.ty {
		return false
	}
	if ty < v.by {
		return false
	}
	return true
}

// Indicates whether the line running horizontally
// along y from lx leftwards to rx passes through v
// Invariant: lx <= rx
func (v View) crossedHorizontally(y, lx, rx float64) bool {
	if y > v.ty || y < v.by {
		return false
	}
	if rx < v.lx {
		return false
	}
	if lx > v.rx {
		return false
	}
	return true
}

// One View overlaps with another if the two Views intersect at
// their borders or if either is contained entirely within the other.
// Reflexive, symmetric, and *not* transitive
func (v View) overlaps(ov View) bool {
	if v.crossedBy(ov) {
		return true
	}
	if ov.crossedBy(v) {
		return true
	}
	return false
}

// Returns four views representing v divided into four non-overlapping equal sized sections
// These four quarters completely cover v
func (v View) quarters() [4]View {
	lx := v.lx
	rx := v.rx
	ty := v.ty
	by := v.by
	midx := lx + (rx-lx)/2
	midy := by + (ty-by)/2
	return [4]View{
		NewView(lx, midx, ty, midy),
		NewView(midx, rx, ty, midy),
		NewView(lx, midx, midy, by),
		NewView(midx, rx, midy, by),
	}
}

// Indicates whether v and ov are equivalent to each other
// Two views are equivalent iff each of the four corners are equal in both views
func (v View) eq(ov View) bool {
	return v == ov
}

// Human readable (sort of) representation of v
func (v View) String() string {
	lx := strconv.FormatFloat(v.lx, 'f', 6, 64)
	rx := strconv.FormatFloat(v.rx, 'f', 6, 64)
	ty := strconv.FormatFloat(v.ty, 'f', 6, 64)
	by := strconv.FormatFloat(v.by, 'f', 6, 64)
	return "[" + lx + " " + rx + " " + ty + " " + by + "]"
}
