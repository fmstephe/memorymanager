package quadtree

import (
	"container/list"
)

//
//	A Simple quadtree collector which will push every element into col
//
func SimpleSurvey[K any]() (fun func(x, y float64, e K), col *list.List) {
	col = list.New()
	fun = func(x, y float64, e K) {
		col.PushBack(e)
	}
	return
}

//
//	A Simple quadtree delete function which indicates that every element given to it should be deleted
//
func SimpleDelete[K any]() (pred func(x, y float64, e K) bool) {
	pred = func(x, y float64, e K) bool {
		return true
	}
	return
}

//
//	A quadtree delete function which indicates that every element given to it should be deleted.
//	Additionally each element deleted will be pushed into col
//
func CollectingDelete[K any]() (pred func(x, y float64, e K) bool, col *list.List) {
	col = list.New()
	pred = func(x, y float64, e K) bool {
		col.PushBack(e)
		return true
	}
	return
}

//
//	Determines if a point lies inside at least one of a slice of *View
//
func contains(vs []View, x, y float64) bool {
	for _, v := range vs {
		if v.contains(x, y) {
			return true
		}
	}
	return false
}

//
//	Determines if a view overlaps at least one of a slice of *View
//
func overlaps(vs []View, oV View) bool {
	for _, v := range vs {
		if oV.overlaps(v) {
			return true
		}
	}
	return false
}
