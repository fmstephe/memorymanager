package fuzzhelper

type deque[T any] struct {
	values []T
}

func newDeque[T any]() *deque[T] {
	return &deque[T]{
		values: []T{},
	}
}

func (d *deque[T]) addMany(newValues []T) {
	d.values = append(d.values, newValues...)
}

func (d *deque[T]) popFirst() T {
	value := d.values[0]
	d.values = d.values[1:]
	return value
}

//lint:ignore U1000 This method is not used right now, but if we change the search strategy in visit_types it may be used
func (d *deque[T]) popLast() T {
	value := d.values[len(d.values)-1]
	d.values = d.values[:len(d.values)-1]
	return value
}

func (d *deque[T]) len() int {
	return len(d.values)
}
