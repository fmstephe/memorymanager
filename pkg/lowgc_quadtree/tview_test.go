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

func TestOrigView(t *testing.T) {
	for _, testValue := range []struct {
		width, height float64
	}{
		{10.0, 10.0},
		{5.5, 30.03},
		{0.0, 0.0},
		{9999.9999, 1.123456},
		{10.01e8, 4.4e3},
	} {
		v := OrigView(testValue.width, testValue.height)
		if v.lx != 0 {
			t.Error("Left x not at origin")
		}
		if v.rx != testValue.width {
			t.Errorf("Right x %10.3f : expecting %10.3f", v.rx, testValue.width)
		}
		if v.ty != 0 {
			t.Error("Top y not at origin")
		}
		if v.by != testValue.height {
			t.Errorf("Bottom y %10.3f : expecting %10.3f", v.by, testValue.height)
		}
	}
}

func TestNewView(t *testing.T) {
	for _, testValue := range []fourPoints{
		{10.0, 10.0, 10.0, 10.0},
		{10.0, 10.0, 10.0, 10.0},
		{5.5, 30.03, 3.45, 5.96},
		{0.0, 0.0, 0.0, 0.0},
		{1.123456, 9999.9999, 9876.5, 12345.5},
		// Negative ones
		{-4.4e3, 10.01e8, -45.0e4, 5.0e5},
		{-5.5, 30.03, -3.45, 5.96},
		{-0.0, 0.0, -0.0, 0.0},
		{-1.123456, 9999.9999, -9876.5, 12345.5},
		{-4.4e3, 10.01e8, -45.0e4, 5.0e5},
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
		{5.5, 5.4, 3.45, 5.96},
		{5.5, 5.6, 3.45, 3.44},
		{5.5, 5.4, 3.45, 3.44},
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

func TestSubtract(t *testing.T) {
	var v1, v2 View
	var eqv []View
	// Left side
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(1, 2, 0, 2)
	eqv = []View{NewView(0, 1, 0, 2)}
	testSubtract(t, v1, v2, eqv, 1, "Left side")
	// Right Side
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(0, 1, 0, 2)
	eqv = []View{NewView(1, 2, 0, 2)}
	testSubtract(t, v1, v2, eqv, 1, "Right side")
	// Bottom Side
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(0, 2, 1, 2)
	eqv = []View{NewView(0, 2, 0, 1)}
	testSubtract(t, v1, v2, eqv, 1, "Bottom side")
	// Top Side
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(0, 2, 0, 1)
	eqv = []View{NewView(0, 2, 1, 2)}
	testSubtract(t, v1, v2, eqv, 1, "Top side")
	// Bottom Left
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(1, 2, 0, 1)
	eqv = []View{NewView(0, 1, 0, 2), NewView(0, 2, 1, 2)}
	testSubtract(t, v1, v2, eqv, 2, "Bottom left")
	// Bottom Right
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(0, 1, 0, 1)
	eqv = []View{NewView(1, 2, 0, 2), NewView(0, 2, 1, 2)}
	testSubtract(t, v1, v2, eqv, 2, "Bottom right")
	// Top Left
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(1, 2, 1, 2)
	eqv = []View{NewView(0, 1, 0, 2), NewView(0, 2, 0, 1)}
	testSubtract(t, v1, v2, eqv, 2, "Top left")
	// Top Right
	v1 = NewView(0, 2, 0, 2)
	v2 = NewView(0, 1, 1, 2)
	eqv = []View{NewView(1, 2, 0, 2), NewView(0, 2, 0, 1)}
	testSubtract(t, v1, v2, eqv, 2, "Top left")
	// Left Right
	v1 = NewView(0, 3, 0, 3)
	v2 = NewView(1, 2, -20, 20)
	eqv = []View{NewView(0, 1, 0, 3), NewView(2, 3, 0, 3)}
	testSubtract(t, v1, v2, eqv, 2, "Left right")
	// Top Bottom
	v1 = NewView(0, 3, 0, 3)
	v2 = NewView(-20, 20, 1, 2)
	eqv = []View{NewView(0, 3, 0, 1), NewView(0, 3, 2, 3)}
	testSubtract(t, v1, v2, eqv, 2, "Top bottom")
	// Centre
	v1 = NewView(0, 4, 0, 4)
	v2 = NewView(1, 2, 1, 2)
	eqv = []View{NewView(0, 1, 0, 4), NewView(2, 4, 0, 4), NewView(0, 4, 0, 1), NewView(0, 4, 2, 4)}
	testSubtract(t, v1, v2, eqv, 4, "Centre")
}

func testSubtract(t *testing.T, v1, v2 View, eqv []View, vNum int, prefix string) {
	vs := v1.Subtract(v2)
	if len(vs) != vNum {
		t.Errorf("%s subtract, expecting %d view found %d", prefix, vNum, len(vs))
	}
	for ei := range eqv {
		for vi := range vs {
			if eqv[ei].eq(vs[vi]) {
				break
			}
			if vi == len(vs)-1 {
				t.Errorf("%s subtract, expected %v found %v", prefix, eqv, vs)
				return
			}
		}
	}
}

func overlap() (v1, v2 View) {
	lx, rx := oPair(negRFLoat64(), negRFLoat64())
	ty, by := oPair(negRFLoat64(), negRFLoat64())
	x1 := rand.Float64()
	y1 := rand.Float64()
	nx, ny := nearest(x1, y1, lx, rx, ty, by)
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
	ty2, by2 := oPair(y1, y2)
	return NewView(lx, rx, ty, by), NewView(lx2, rx2, ty2, by2)
}

func disjoint() (v1, v2 View) {
	lx, rx := oPair(negRFLoat64(), negRFLoat64())
	ty, by := oPair(negRFLoat64(), negRFLoat64())
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
	ty2, by2 := oPair(y1, y2)
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
