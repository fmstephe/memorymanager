package lowgc_quadtree

import (
	"container/list"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

const dups = 10

type point struct {
	x, y float64
}

// testRand will produce the same random numbers every time
// This is done to make the benchmarks consistent between runs
var testRand = rand.New(rand.NewSource(1))

func buildTestTrees() []T[string] {
	return []T[string]{
		NewQuadTree[string](0, 10, 0, 10),
		NewQuadTree[string](0, 1, 0, 2),
		NewQuadTree[string](0, 100, 0, 300),
		NewQuadTree[string](0, 20.4, 0, 35.6),
		NewQuadTree[string](0, 1e10, 0, 500.00000001),
		// Negative regions
		NewQuadTree[string](-10, 10, -10, 10),
		NewQuadTree[string](-1, 1, -2, 2),
		NewQuadTree[string](-100, 100, -300, 300),
		NewQuadTree[string](-20.4, 20.4, -35.6, 35.6),
		NewQuadTree[string](-1e10, 1e10, -500.00000001, 500.00000001),
	}
}

func TestOverflowLeaf(t *testing.T) {
	tree := NewQuadTree[string](0, 1, 0, 1)
	ps := fillView(tree.View(), 70)
	for i, p := range ps {
		tree.Insert(p.x, p.y, fmt.Sprintf("test-%d", i))
	}
	fun, results := SimpleSurvey[string]()
	tree.Survey(tree.View(), fun)
	if 70 != results.Len() {
		t.Errorf("Failed to retrieve 70 elements in scatter test, found only %d", results.Len())
	}
}

// Test that we can insert a single element into the tree and then retrieve it
func TestOneElement(t *testing.T) {
	testTrees := buildTestTrees()
	for _, tree := range testTrees {
		testOneElement(tree, t)
	}
}

func testOneElement(tree T[string], t *testing.T) {
	x, y := randomPosition(tree.View())
	tree.Insert(x, y, "test")
	fun, results := SimpleSurvey[string]()
	tree.Survey(tree.View(), fun)
	if results.Len() != 1 || "test" != results.Front().Value {
		t.Errorf("Failed to find required element at (%f,%f), in tree \n%v", x, y, tree)
	}
}

// Test that if we add 5 elements into a single quadrant of a fresh tree
// We can successfully retrieve those elements. This test is tied to
// the implementation detail that a quadrant with 5 elements will
// over-load a single leaf and must rearrange itself to fit the 5th
// element in.
func TestFullLeaf(t *testing.T) {
	testTrees := buildTestTrees()
	for _, tree := range testTrees {
		views := tree.View().quarters()
		for _, view := range views {
			testFullLeaf(tree, view, "v1", t)
		}
	}
}

func testFullLeaf(tree T[string], v View, msg string, t *testing.T) {
	for i := 0; i < LEAF_SIZE; i++ {
		x, y := randomPosition(v)
		name := "test" + strconv.Itoa(i)
		tree.Insert(x, y, name)
	}
	fun, results := SimpleSurvey[string]()
	tree.Survey(v, fun)
	if results.Len() != LEAF_SIZE {
		t.Error(msg, "Inserted 5 elements into a fresh quadtree and retrieved only ", results.Len())
	}
}

// Tests that we can add a large number of random elements to a tree
// and create random views for collecting from the populated tree.
func TestScatter(t *testing.T) {
	testTrees := buildTestTrees()
	for _, tree := range testTrees {
		testScatter(tree, t)
	}
	testTrees = buildTestTrees()
	for _, tree := range testTrees {
		testScatterDup(tree, t)
	}
}

func testScatter(tree T[string], t *testing.T) {
	t.Helper()
	ps := fillView(tree.View(), 1000)
	for _, p := range ps {
		tree.Insert(p.x, p.y, "test")
	}
	for i := 0; i < 1000; i++ {
		sv := subView(tree.View())
		var count int
		for _, v := range ps {
			if sv.contains(v.x, v.y) {
				count++
			}
		}
		fun, results := SimpleSurvey[string]()
		tree.Survey(sv, fun)
		if count != results.Len() {
			t.Errorf("Failed to retrieve %d elements in scatter test, found only %d", count, results.Len())
		}
	}
}

// Tests that we can add multiple elements to the same location
// and still retrieve all elements, including duplicates, using
// randomly generated views.
func testScatterDup(tree T[string], t *testing.T) {
	ps := fillView(tree.View(), 1000)
	for _, p := range ps {
		for i := 0; i < dups; i++ {
			tree.Insert(p.x, p.y, "test_"+strconv.Itoa(i))
		}
	}
	for i := 0; i < 1000; i++ {
		sv := subView(tree.View())
		var count int
		for _, v := range ps {
			if sv.contains(v.x, v.y) {
				count++
			}
		}
		fun, results := SimpleSurvey[string]()
		tree.Survey(sv, fun)
		if count*dups != results.Len() {
			t.Error("Failed to retrieve %i elements in duplicate scatter test, found only %i", count*dups, results.Len())
		}
	}
}

func testSurvey(tree T[string], view View, fun func(x, y float64, e string), collected, expCol *list.List, t *testing.T, errPfx string) {
	tree.Survey(view, fun)
	if collected.Len() != expCol.Len() {
		t.Errorf("%s: Expecting %v collected element(s), found %v", errPfx, expCol.Len(), collected.Len())
	}
	/* This code checks that every expected element is present
		   In practice this is too slow - disabled
	OUTER_LOOP:
		for i := 0; i < expCol.Len(); i++ {
			expVal := expCol.At(i)
			for j := 0; j < collected.Len(); j++ {
				colVal := collected.At(j)
				if expVal == colVal {
					continue OUTER_LOOP
				}
			}
			t.Errorf("%s: Expecting to find %v in collected vector, was not found", errPfx, expCol.At(i))
		}
	*/
}

func randomPosition(v View) (x, y float64) {
	x = testRand.Float64()*(v.rx-v.lx) + v.lx
	y = testRand.Float64()*(v.by-v.ty) + v.ty
	return
}

func fillView(v View, c int) []point {
	ps := make([]point, c)
	for i := 0; i < c; i++ {
		x, y := randomPosition(v)
		ps[i] = point{x: x, y: y}
	}
	return ps
}

func subView(v View) View {
	lx := testRand.Float64()*(v.rx-v.lx) + v.lx
	rx := testRand.Float64()*(v.rx-lx) + lx
	ty := testRand.Float64()*(v.by-v.ty) + v.ty
	by := testRand.Float64()*(v.by-ty) + ty
	return NewView(lx, rx, ty, by)
}
