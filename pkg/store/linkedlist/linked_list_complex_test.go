package linkedlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	pushH   = "push head"
	pushT   = "push tail"
	removeH = "remove head"
	removeT = "remove tail"
)

type ListAction struct {
	action string
	value  int
	head   int
	tail   int
}

func (a *ListAction) DoAction(t *testing.T, store *Store[TestListData], list *List[TestListData]) {
	switch a.action {
	case pushH:
		data := list.PushHead(store)
		data.intField = a.value
	case pushT:
		data := list.PushTail(store)
		data.intField = a.value
	case removeH:
		list.RemoveHead(store)
	case removeT:
		list.RemoveTail(store)
	}

	if !list.IsEmpty() {
		// Assert the head of the list
		data := list.PeakHead(store)
		assert.Equal(t, a.head, data.intField)

		// Assert the tail of the list
		data = list.PeakTail(store)
		assert.Equal(t, a.tail, data.intField)
	}
}

func TestLinkedList_CommonOperations(t *testing.T) {
	for _, testData := range []struct {
		name       string
		actions    []ListAction
		finalState []int
	}{
		{"do nothing, final result is an empty list", []ListAction{}, []int{}},

		// Push one element
		{"push head 1, final result is [1]", []ListAction{{pushH, 1, 1, 1}}, []int{1}},
		{"push tail 1, final result is [1]", []ListAction{{pushT, 1, 1, 1}}, []int{1}},

		// Push two elements
		{"push head 1, 2, final result is [2,1]", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}}, []int{2, 1}},
		{"push tail 1, 2, final result is [1,2]", []ListAction{{pushT, 1, 1, 1}, {pushT, 2, 1, 2}}, []int{1, 2}},
		{"push head 1, push tail 2, final result is [1,2]", []ListAction{{pushH, 1, 1, 1}, {pushT, 2, 1, 2}}, []int{1, 2}},
		{"push tail 1, push head 2, final result is [2,1]", []ListAction{{pushT, 1, 1, 1}, {pushH, 2, 2, 1}}, []int{2, 1}},

		// Push three elements
		{"push head 1, 2, 3, final result is [3,2,1]", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}, {pushH, 3, 3, 1}}, []int{3, 2, 1}},
		{"push tail 1, 2, 3, final result is [1,2,3]", []ListAction{{pushT, 1, 1, 1}, {pushT, 2, 1, 2}, {pushT, 3, 1, 3}}, []int{1, 2, 3}},
		{"push head 1, 2, push tail 3, final result is [2,1,3]", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}, {pushT, 3, 2, 3}}, []int{2, 1, 3}},
		{"push tail 1, 2, push head 3, final result is [3,1,2]", []ListAction{{pushT, 1, 1, 1}, {pushT, 2, 1, 2}, {pushH, 3, 3, 2}}, []int{3, 1, 2}},
		{"push head 1, push tail 2, push head 3, final result is [3,2,1]", []ListAction{{pushH, 1, 1, 1}, {pushT, 2, 1, 2}, {pushH, 3, 3, 2}}, []int{3, 2, 1}},
		{"push tail 1, push head 2, push tail 3, final result is [2,1,3]", []ListAction{{pushT, 1, 1, 1}, {pushH, 2, 2, 1}, {pushT, 3, 2, 3}}, []int{2, 1, 3}},

		// Push two elements and remove them
		{"push head 1, 2, remove them all from head", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}, {removeH, -1, 1, 1}, {removeH, -1, -1, -1}}, []int{}},
		{"push head 1, 2, remove them all from tail", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}, {removeT, -1, 2, 2}, {removeT, -1, -1, -1}}, []int{}},
		{"push tail 1, 2, remove them all from head", []ListAction{{pushT, 1, 1, 1}, {pushT, 2, 1, 2}, {removeH, -1, 2, 2}, {removeH, -1, -1, -1}}, []int{}},
		{"push tail 1, 2, remove them all from tail", []ListAction{{pushT, 1, 1, 1}, {pushT, 2, 1, 2}, {removeT, -1, 1, 1}, {removeT, -1, -1, -1}}, []int{}},

		// Push three elements and remove them
		{"push head 1, 2, 3, remove them all from head", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}, {pushH, 3, 3, 1}, {removeH, -1, 2, 1}, {removeH, -1, 1, 1}, {removeH, -1, -1, -1}}, []int{}},
		{"push head 1, 2, 3, remove them all from tail", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}, {pushH, 3, 3, 1}, {removeT, -1, 3, 2}, {removeT, -1, 3, 3}, {removeT, -1, -1, -1}}, []int{}},

		// Push some elements, remove some, add some more
		{"push head 1, 2, 3, remove from tail and then from head, push tail 4,5,6, final result is [2,4,5,6]", []ListAction{{pushH, 1, 1, 1}, {pushH, 2, 2, 1}, {pushH, 3, 3, 1}, {removeT, -1, 3, 2}, {removeH, -1, 2, 2}, {pushT, 4, 2, 4}, {pushT, 5, 2, 5}, {pushT, 6, 2, 6}}, []int{2, 4, 5, 6}},
	} {
		t.Run(testData.name, func(t *testing.T) {
			store := New[TestListData]()
			l := store.NewList()

			for _, action := range testData.actions {
				action.DoAction(t, store, &l)
			}

			assertContains(t, l, store, testData.finalState)
		})
	}
}
