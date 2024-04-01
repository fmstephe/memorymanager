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

// Determines if a point lies inside at least one of a slice of *View
func contains(vs []View, x, y float64) bool {
	for _, v := range vs {
		if v.containsPoint(x, y) {
			return true
		}
	}
	return false
}

// Determines if a view overlaps at least one of a slice of *View
func overlaps(vs []View, oV View) bool {
	for _, v := range vs {
		if oV.overlaps(v) {
			return true
		}
	}
	return false
}
