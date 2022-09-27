package lowgc_quadtree

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Struct of four points
type fourPoints struct {
	lx, rx, ty, by float64
}

func TestNewView(t *testing.T) {
	for _, testValue := range []fourPoints{
		{10.0, 10.0, 10.0, 10.0},
		{10.0, 10.0, 10.0, 10.0},
		{5.5, 30.03, 5.96, 3.45},
		{0.0, 0.0, 0.0, 0.0},
		{1.123456, 9999.9999, 12345.5, 9876.5},
		// Negative ones
		{-4.4e3, 10.01e8, 5.0e5, -45.0e4},
		{-5.5, 30.03, 5.96, -3.45},
		{-0.0, 0.0, 0.0, -0.0},
		{-1.123456, 9999.9999, 12345.5, -9876.5},
		{-4.4e3, 10.01e8, 5.0e5, -45.0e4},
	} {
		v := NewView(testValue.lx, testValue.rx, testValue.ty, testValue.by)
		if v.lx != testValue.lx {
			t.Errorf("Left x %10.3f : expecting %10.3f", v.lx, testValue.lx)
		}
		if v.rx != testValue.rx {
			t.Errorf("Right x %10.3f : expecting %10.3f", v.rx, testValue.rx)
		}
		if v.ty != testValue.ty {
			t.Errorf("Right x %10.3f : expecting %10.3f", v.ty, testValue.ty)
		}
		if v.by != testValue.by {
			t.Errorf("Bottom y %10.3f : expecting %10.3f", v.by, testValue.by)
		}
	}
}

func TestIllegalView(t *testing.T) {
	for _, testValue := range []fourPoints{
		{5.5, 5.4, 5.96, 3.45},
		{5.5, 5.6, 3.44, 3.45},
		{5.5, 5.4, 3.44, 3.45},
	} {
		require.Panics(t, func() {
			NewView(testValue.lx, testValue.rx, testValue.ty, testValue.by)
		})
	}
}

func TestOverLap(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10000; i++ {
		v1, v2 := overlap()
		if !v1.overlaps(v2) {
			t.Errorf("View %v and %v not reported as overlapping", v1, v2)
			t.Error("+------------------------------------------------+")
		}
		if !v2.overlaps(v1) {
			t.Errorf("View %v and %v not reported as overlapping", v2, v1)
			t.Error("<------------------------------------------------>")
		}
	}
}

func TestDisjoint(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 1000; i++ {
		v1, v2 := disjoint()
		if v1.overlaps(v2) {
			t.Errorf("View %v and %v are reported as overlapping", v1, v2)
			t.Error("+------------------------------------------------+")
		}
		if v2.overlaps(v1) {
			t.Errorf("View %v and %v are reported as overlapping", v2, v1)
			t.Error("<------------------------------------------------>")
		}
	}
}

func overlap() (v1, v2 View) {
	// Generate a random view
	lx, rx := oPair(negRFLoat64(), negRFLoat64())
	by, ty := oPair(negRFLoat64(), negRFLoat64())

	// Generate a new random point
	x1 := rand.Float64()
	y1 := rand.Float64()

	// Find nearest point to x1,y1 point from corners of the view
	nx, ny := nearest(x1, y1, lx, rx, ty, by)

	// Carefully generate a second random point so that the x1,x2,y1,y2 view overlaps
	var x2, y2 float64
	if x1 > nx {
		x2 = nx - rand.Float64()
	} else {
		x2 = rand.Float64() + nx
	}
	if y1 > ny {
		y2 = ny - rand.Float64()
	} else {
		y2 = rand.Float64() + ny
	}

	lx2, rx2 := oPair(x1, x2)
	by2, ty2 := oPair(y1, y2)
	return NewView(lx, rx, ty, by), NewView(lx2, rx2, ty2, by2)
}

func disjoint() (v1, v2 View) {
	lx, rx := oPair(negRFLoat64(), negRFLoat64())
	by, ty := oPair(negRFLoat64(), negRFLoat64())
	v1 = NewView(lx, rx, ty, by)
	var x1, y1 float64
	for true {
		x1 = negRFLoat64()
		y1 = negRFLoat64()
		if !v1.contains(x1, y1) {
			break
		}
	}
	nx, ny := nearest(x1, y1, lx, rx, ty, by)
	var x2, y2 float64
	if x1 < nx {
		x2 = nx - rand.Float64()
	} else {
		x2 = rand.Float64() + nx
	}
	if y1 < ny {
		y2 = ny - rand.Float64()
	} else {
		y2 = rand.Float64() + ny
	}
	lx2, rx2 := oPair(x1, x2)
	by2, ty2 := oPair(y1, y2)
	v2 = NewView(lx2, rx2, ty2, by2)
	return
}

func oPair(f1, f2 float64) (r1, r2 float64) {
	r1 = math.Min(f1, f2)
	r2 = math.Max(f1, f2)
	return
}

func nearest(x, y, lx, rx, ty, by float64) (nx, ny float64) {
	d1 := dist(x, y, lx, ty)
	d2 := dist(x, y, rx, ty)
	d3 := dist(x, y, lx, by)
	d4 := dist(x, y, rx, by)

	n1 := math.Min(d1, d2)
	n2 := math.Min(n1, d3)
	n3 := math.Min(n2, d4)

	switch {
	case n3 == d1:
		return lx, ty
	case n3 == d2:
		return rx, ty
	case n3 == d3:
		return lx, by
	case n3 == d4:
		return rx, by
	default:
		panic("unreachable")
	}
}

func dist(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2.0) + math.Pow(y1-y2, 2.0))
}

func negRFLoat64() float64 {
	f := rand.Float64()
	d := rand.Float64()
	if d < 0.5 {
		return -f
	}
	return f
}
