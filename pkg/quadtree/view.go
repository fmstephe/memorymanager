package quadtree

import (
	"fmt"
	"math"
	"strconv"
)

// A View is a rectangle defined by four points, from two x coords and two y
// coords. The coordinate system places the origin (0,0) in the top left
// hand corner if you drew it on a piece of paper. However, it is allowed
// to have view's above and to the left of origin extending into negative coords.
// Valid is initially false, and is used to construct Views which cover no area at all.
// lx - left-x
// rx - right-x
// Invariant: lx <= rx
// ty - top-y
// by - bottom-y
// Invariant: ty <= by
// The rectangle is defined the four points;
// (lx,ty),(lx,by),(rx,ty),(rx,by)
// Top left point on 2D plane is (0,0)
// The zeroed View is a zero area plane at origin (0,0)
type View struct {
	lx, rx, ty, by float64
	valid          bool
}

// Returns a new View struct with top left-hand corner
// at the point (0,0) and with the width and height as provided
// Providing negative values for width or height will cause a panic
func OrigView(width, height float64) View {
	if width < 0 || height < 0 {
		msg := fmt.Sprintf("Cannot create view with negative elements. width : %10.3f height : %10.3f", width, height)
		panic(msg)
	}
	return View{0, width, 0, height, true}
}

// Returns a new View which covers a single point
// i.e. a view with no area
func PointView(x, y float64) View {
	return NewView(x, x, y, y)
}

// Returns a new View struct with the four
func NewView(lx, rx, ty, by float64) View {
	if rx < lx || by < ty {
		msg := fmt.Sprintf("Cannot create view with inverted corners. lx : %10.3f rx : %10.3f, ty : %10.3f by %10.3f", lx, rx, ty, by)
		panic(msg)
	}
	return View{lx, rx, ty, by, true}
}

// Indicates whether this View contains the point (x,y)
func (v View) contains(x, y float64) bool {
	return x >= v.lx && x <= v.rx && y <= v.by && y >= v.ty
}

// Indicates whether any of the four edges
// of ov pass through v
func (v View) xBy(ov View) bool {
	if v.xV(ov.lx, ov.ty, ov.by) {
		return true
	}
	if v.xV(ov.rx, ov.ty, ov.by) {
		return true
	}
	if v.xH(ov.ty, ov.lx, ov.rx) {
		return true
	}
	if v.xH(ov.by, ov.lx, ov.rx) {
		return true
	}
	return false
}

// Indicates whether the line running vertically
// along point x from ty down to by passes through v
// Invariant: ty <= by
func (v View) xV(x, ty, by float64) bool {
	if x < v.lx || x > v.rx {
		return false
	}
	if by < v.ty {
		return false
	}
	if ty > v.by {
		return false
	}
	return true
}

// Indicates whether the line running horizontally
// along point y from lx leftwards to rx passes through v
// Invariant: lx <= rx
func (v View) xH(y, lx, rx float64) bool {
	if y < v.ty || y > v.by {
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
	if v.xBy(ov) {
		return true
	}
	if ov.xBy(v) {
		return true
	}
	return false
}

// Returns the width of the View
func (v View) width() float64 {
	return v.rx - v.lx
}

// Returns the width of the View
func (v View) height() float64 {
	return v.by - v.ty
}

// Returns four views representing v divided into four non-overlapping equal sized sections
// These four quarters completely cover v
// TODO This function should return a slice of views created by dividing v an arbitrary number of times
func (v View) quarters() (v1, v2, v3, v4 View) {
	lx := v.lx
	rx := v.rx
	ty := v.ty
	by := v.by
	midx := lx + (rx-lx)/2
	midy := ty + (by-ty)/2
	v1 = NewView(lx, midx, ty, midy)
	v2 = NewView(midx, rx, ty, midy)
	v3 = NewView(lx, midx, midy, by)
	v4 = NewView(midx, rx, midy, by)
	return
}

// Indicates whether v and ov are equivalent to each other
// Two views are equivalent iff each of the four corners are equal in both views
func (v View) eq(ov View) bool {
	return v.lx == ov.lx && v.rx == ov.rx && v.ty == ov.ty && v.by == ov.by
}

// Returns a slice of views which satisfy:
// 1: None are overlapping with themselves or with ov
// 2: When combined with ov they completely cover v
// Intuitively imagine v is make up every view in []*View returned plus ov
// To subtract ov you just need to take it away and return the []*View
func (v View) Subtract(ov View) []View {
	if v.eq(ov) {
		return make([]View, 0, 0)
	}
	if !v.overlaps(ov) {
		return []View{v}
	}
	vs := make([]View, 0, 4)
	if ov.lx > v.lx {
		vs = append(vs, NewView(v.lx, ov.lx, v.ty, v.by)) // Grab the left most rectangle
	}
	if ov.rx < v.rx {
		vs = append(vs, NewView(ov.rx, v.rx, v.ty, v.by)) // Grab the right most rectangle
	}
	if ov.ty > v.ty {
		vs = append(vs, NewView(v.lx, v.rx, v.ty, ov.ty)) // Grab the top most rectangle
	}
	if ov.by < v.by {
		vs = append(vs, NewView(v.lx, v.rx, ov.by, v.by)) // Grab the bottom rectangle
	}
	return vs
}

// Returns a view which represents the overlapping region of v and ov
func (v View) Intersect(ov View) View {
	if v.eq(ov) {
		return v
	}
	if !v.overlaps(ov) {
		return View{}
	}
	ilx := math.Max(v.lx, ov.lx)
	irx := math.Min(v.rx, ov.rx)
	ity := math.Max(v.ty, ov.ty)
	iby := math.Min(v.by, ov.by)
	return NewView(ilx, irx, ity, iby)
}

// Human readable (sort of) representation of v
func (v View) String() string {
	lx := strconv.FormatFloat(v.lx, 'f', 6, 64)
	rx := strconv.FormatFloat(v.rx, 'f', 6, 64)
	ty := strconv.FormatFloat(v.ty, 'f', 6, 64)
	by := strconv.FormatFloat(v.by, 'f', 6, 64)
	return "[" + lx + " " + rx + " " + ty + " " + by + "]"
}
