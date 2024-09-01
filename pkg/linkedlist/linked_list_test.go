package linkedlist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestListData struct {
	intField int
}

// Show that the zero value of a value of type List[T] is an empty list
func TestLinkedList_ZeroValue(t *testing.T) {
	store := New[TestListData]()
	var l List[TestListData]

	// The zero value list should be empty
	assert.True(t, l.IsEmpty())
	assert.Equal(t, 0, l.Len(store))

	// Show we can survey/filter a zero value list, it won't do anything much,
	// we just prove it won't panic
	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		return true
	}))
	l.Filter(store, func(d *TestListData) bool {
		return false
	})
}

// Show that we can create a new list. A new list will be empty
func TestLinkedList_New(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// The new list should be empty
	assert.True(t, l.IsEmpty())
	assert.Equal(t, 0, l.Len(store))

	// Show we can survey/filter an empty list, it won't do anything much,
	// we just prove it won't panic
	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		return true
	}))
	l.Filter(store, func(d *TestListData) bool {
		return false
	})
}

// Show that we can add a new node to a list and set its embedded data
func TestLinkedList_SetData(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()
	n := l.PushTail(store)
	n.intField = 1

	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		assert.Equal(t, 1, d.intField)
		return true
	}))
}

// Show that we can add many nodes to the list and set all of their embedded
// data to different values
func TestLinkedList_SetData_Multiple(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// PushTail new nodes, collect copy of the inserted data
	inserted := []int{}
	for i := 0; i < 10; i++ {
		d := l.PushTail(store)
		d.intField = i
		inserted = append(inserted, d.intField)
	}

	assertContains(t, l, store, inserted)
}

// Show that we can insert nodes into a linked list, and then modify the
// embedded data of the existing nodes
func TestLinkedList_SetData_ModifyAfterPushTail(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// PushTail new nodes, give each one unique data values
	for i := 0; i < 10; i++ {
		d := l.PushTail(store)
		d.intField = i
	}

	// Survey the list, updating the embedded data of each node. Collect a
	// copy of the updated data value
	updated := []int{}
	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		// modify d
		d.intField *= 2
		// Store the updated version
		updated = append(updated, d.intField)
		return true
	}))

	assertContains(t, l, store, updated)
}

// Show that we can add many nodes to the list and then remove them all using
// filter
func TestLinkedList_AddManyRemoveAll(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	for i := 0; i < 10; i++ {
		l.PushTail(store)
	}
	l.Filter(store, func(_ *TestListData) bool {
		return false
	})

	assert.True(t, l.IsEmpty())
	assert.Equal(t, 0, l.Len(store))
}

// Show that we can add many nodes to a list, remove them all and then add new
// node. This demonstrates that a list still functions correctly after being
// filled and emptied.
func TestLinkedList_AddManyRemoveAll_AddOne(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// Add many nodes to the list
	for i := 0; i < 10; i++ {
		l.PushTail(store)
	}
	// Delete them all
	l.Filter(store, func(_ *TestListData) bool {
		return false
	})

	// PushTail a single node into the list, and check that it's there
	l.PushTail(store)
	assert.False(t, l.IsEmpty())
	assert.Equal(t, 1, l.Len(store))
}

// Show that we can add many nodes to a list and then use Filter but not
// actually remove any elements. All the inserted elements should still be in
// the list.
func TestLinkedList_AddManyRemoveNone(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// Add many nodes
	for i := 0; i < 10; i++ {
		l.PushTail(store)
	}
	// All elements are in the list
	assert.False(t, l.IsEmpty())
	assert.Equal(t, 10, l.Len(store))

	// Remove none
	l.Filter(store, func(_ *TestListData) bool {
		return true
	})
	// All elements are still in the list
	assert.False(t, l.IsEmpty())
	assert.Equal(t, 10, l.Len(store))
}

// Show that we can add many nodes to a list and then remove a single one. This
// test is repeated to show that we can remove any one of the nodes in a list.
// This is intended to show that we haven't missed any corner cases in the
// Filter function.
func TestLinkedList_AddMany_RemoveOne(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("Deleting the %dth node", i), func(t *testing.T) {
			store := New[TestListData]()
			l := store.NewList()

			// PushTail many nodes
			for j := 0; j < 10; j++ {
				d := l.PushTail(store)
				d.intField = j
			}
			// Filter the i'th one
			l.Filter(store, func(d *TestListData) bool {
				return d.intField != i
			})

			// Show that we removed one of the elements of the list
			assert.False(t, l.IsEmpty())
			assert.Equal(t, 9, l.Len(store))
		})
	}
}

// Show that we can add multiple elements to a list, and then remove some using filter.
func TestLinkedList_AddSomeRemoveSome(t *testing.T) {
	for _, testData := range []struct {
		name      string
		inserts   []int
		remove    []int
		remaining []int
	}{
		{"insert nothing, remove nothing, nothing remains", []int{}, []int{}, []int{}},
		{"insert one thing, remove nothing, the inserted data remains", []int{1}, []int{}, []int{1}},
		{"insert one thing, remove it, nothing remains", []int{1}, []int{1}, []int{}},
		{"insert two things, remove nothing, two items remain", []int{1, 2}, []int{}, []int{1, 2}},
		{"insert two things, remove the first, the second remains", []int{1, 2}, []int{1}, []int{2}},
		{"insert two things, remove the second, the first remains", []int{1, 2}, []int{2}, []int{1}},
		{"insert two things, remove them both, nothing remains", []int{1, 2}, []int{1, 2}, []int{}},
		{"insert three things, remove nothing, three items remain", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},
		{"insert three things, remove the first, the last two remain", []int{1, 2, 3}, []int{1}, []int{2, 3}},
		{"insert three things, remove the second, the first and third remain", []int{1, 2, 3}, []int{2}, []int{1, 3}},
		{"insert three things, remove the third, the first and second remain", []int{1, 2, 3}, []int{3}, []int{1, 2}},
		{"insert three things, remove the first and second, the third remains", []int{1, 2, 3}, []int{1, 2}, []int{3}},
		{"insert three things, remove the first and third, the second remains", []int{1, 2, 3}, []int{1, 3}, []int{2}},
		{"insert three things, remove the second and third, the first remains", []int{1, 2, 3}, []int{2, 3}, []int{1}},
		{"insert three things, remove all of them, nothing remains", []int{1, 2, 3}, []int{1, 2, 3}, []int{}},
		{"insert four things, remove nothing, four items remain", []int{1, 2, 3, 4}, []int{}, []int{1, 2, 3, 4}},
		{"insert four things, remove the first, the second, third and fourth items remain", []int{1, 2, 3, 4}, []int{1}, []int{2, 3, 4}},
		{"insert four things, remove the second, the first, third and fourth items remain", []int{1, 2, 3, 4}, []int{2}, []int{1, 3, 4}},
		{"insert four things, remove the third, the first, second and fourth items remain", []int{1, 2, 3, 4}, []int{3}, []int{1, 2, 4}},
		{"insert four things, remove the fourth, the first, second and third items remain", []int{1, 2, 3, 4}, []int{4}, []int{1, 2, 3}},
		{"insert four things, remove the first and second, the third and fourth items remain", []int{1, 2, 3, 4}, []int{1, 2}, []int{3, 4}},
		{"insert four things, remove the first and third, the second and fourth items remain", []int{1, 2, 3, 4}, []int{1, 3}, []int{2, 4}},
		{"insert four things, remove the first and fourth, the second and third items remain", []int{1, 2, 3, 4}, []int{1, 4}, []int{2, 3}},
		{"insert four things, remove the second and third, the first and fourth items remain", []int{1, 2, 3, 4}, []int{2, 3}, []int{1, 4}},
		{"insert four things, remove the second and fourth, the first and third items remain", []int{1, 2, 3, 4}, []int{2, 4}, []int{1, 3}},
		{"insert four things, remove the third and fourth, the first and second items remain", []int{1, 2, 3, 4}, []int{3, 4}, []int{1, 2}},
		{"insert four things, remove the first, second and third, the fourth item remains", []int{1, 2, 3, 4}, []int{1, 2, 3}, []int{4}},
		{"insert four things, remove the first, second and fourth, the third item remains", []int{1, 2, 3, 4}, []int{1, 2, 4}, []int{3}},
		{"insert four things, remove the first, third and fourth, the second item remains", []int{1, 2, 3, 4}, []int{1, 3, 4}, []int{2}},
		{"insert four things, remove the second, third and fourth, the second item remains", []int{1, 2, 3, 4}, []int{2, 3, 4}, []int{1}},
	} {
		t.Run(testData.name, func(t *testing.T) {
			store := New[TestListData]()
			l := makeList(store, testData.inserts)

			// Delete the element from the list
			l.Filter(store, func(d *TestListData) bool {
				for _, delVal := range testData.remove {
					if d.intField == delVal {
						return false
					}
				}
				return true
			})
			assertContains(t, l, store, testData.remaining)
		})
	}
}

// Show that we can append a list to another list.
func TestAppend(t *testing.T) {
	for _, testData := range []struct {
		name       string
		firstList  []int
		secondList []int
		combined   []int
	}{
		// Empty lists
		{"Start with an empty list, combine with an empty list", []int{}, []int{}, []int{}},

		// First list is empty
		{"Start with an empty list, combine with a one node list", []int{}, []int{10}, []int{10}},
		{"Start with an empty list, combine with a two node list", []int{}, []int{10, 20}, []int{10, 20}},
		{"Start with an empty list, combine with a three node list", []int{}, []int{10, 20, 30}, []int{10, 20, 30}},

		// Second list is empty
		{"Start with a one node list, combine with an empty list", []int{1}, []int{}, []int{1}},
		{"Start with a two node list, combine with an empty list", []int{1, 2}, []int{}, []int{1, 2}},
		{"Start with a three node list, combine with an empty list", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},

		// First list varies, second list has one node
		{"Start with a one node list, combine with a one node list", []int{1}, []int{10}, []int{1, 10}},
		{"Start with a two node list, combine with a one node list", []int{1, 2}, []int{10}, []int{1, 2, 10}},
		{"Start with a three node list, combine with a one node list", []int{1, 2, 3}, []int{10}, []int{1, 2, 3, 10}},

		// First list has one node, second list varies
		{"Start with a one node list, combine with a two node list", []int{1}, []int{10, 20}, []int{1, 10, 20}},
		{"Start with a one node list, combine with a three node list", []int{1}, []int{10, 20, 30}, []int{1, 10, 20, 30}},

		// First list has two nodes, second list varies
		{"Start with a two node list, combine with a two node list", []int{1, 2}, []int{10, 20}, []int{1, 2, 10, 20}},
		{"Start with a two node list, combine with a three node list", []int{1, 2}, []int{10, 20, 30}, []int{1, 2, 10, 20, 30}},

		// First list has three nodes, second list varies
		{"Start with a three node list, combine with a two node list", []int{1, 2, 3}, []int{10, 20}, []int{1, 2, 3, 10, 20}},
		{"Start with a three node list, combine with a three node list", []int{1, 2, 3}, []int{10, 20, 30}, []int{1, 2, 3, 10, 20, 30}},
	} {
		t.Run(testData.name, func(t *testing.T) {
			store := New[TestListData]()
			l1 := makeList(store, testData.firstList)
			l2 := makeList(store, testData.secondList)
			l1.Append(store, l2)
			assertContains(t, l1, store, testData.combined)
		})
	}
}

func makeList(store *Store[TestListData], datas []int) List[TestListData] {
	l := store.NewList()

	// Add the items to the list
	for _, val := range datas {
		d := l.PushTail(store)
		d.intField = val
	}

	return l
}

func assertContains(t *testing.T, l List[TestListData], store *Store[TestListData], expected []int) {
	t.Helper()

	// Survey the list, collecting a copy of the embedded data values
	surveyed := []int{}
	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		surveyed = append(surveyed, d.intField)
		return true
	}))

	assert.ElementsMatch(t, expected, surveyed)
	assert.Equal(t, len(expected), l.Len(store))
}
