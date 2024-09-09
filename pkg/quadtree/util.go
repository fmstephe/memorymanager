// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package quadtree

// A Simple quadtree collector which will push every element into col
func LimitSurvey[K any](limit int) (fun func(x, y float64, e *K) bool, colP *[]K) {
	count := 0
	col := []K{}
	colP = &col

	fun = func(x, y float64, e *K) bool {
		if count >= limit {
			return false
		}

		col = *colP
		col = append(col, *e)
		colP = &col
		count++
		return true
	}
	return fun, colP
}

// A Simple quadtree collector which will push every element into col
func SliceSurvey[K any]() (fun func(x, y float64, e *K) bool, colP *[]K) {
	col := []K{}
	colP = &col
	fun = func(x, y float64, e *K) bool {
		col = *colP
		col = append(col, *e)
		colP = &col
		return true
	}
	return fun, colP
}
