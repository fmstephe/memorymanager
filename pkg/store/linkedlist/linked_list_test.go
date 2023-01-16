package linkedlist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestListData struct {
	intField int
}

// Show that we can create a new list. A new list will be empty with a nil underlying pointer
func TestLinkedList_New(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// The new list should be empty
	assert.Equal(t, 0, l.Len(store))

	// The underlying pointer of an empty list should be nil
	p := l.getPointer()
	assert.True(t, p.IsNil())

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
	n := l.Insert(store)
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

	// Insert new nodes, collect copy of the inserted data
	inserted := []int{}
	for i := 0; i < 10; i++ {
		d := l.Insert(store)
		d.intField = i
		inserted = append(inserted, d.intField)
	}

	assertContains(t, l, store, inserted)
}

// Show that we can insert nodes into a linked list, and then modify the
// embedded data of the existing nodes
func TestLinkedList_SetData_ModifyAfterInsert(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// Insert new nodes, give each one unique data values
	for i := 0; i < 10; i++ {
		d := l.Insert(store)
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
		l.Insert(store)
	}
	l.Filter(store, func(_ *TestListData) bool {
		return false
	})

	p := l.getPointer()
	assert.True(t, p.IsNil())
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
		l.Insert(store)
	}
	// Delete them all
	l.Filter(store, func(_ *TestListData) bool {
		return false
	})

	// Insert a single node into the list, and check that it's there
	l.Insert(store)
	p := l.getPointer()
	assert.False(t, p.IsNil())
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
		l.Insert(store)
	}
	// All elements are in the list
	p := l.getPointer()
	assert.False(t, p.IsNil())
	assert.Equal(t, 10, l.Len(store))

	// Remove none
	l.Filter(store, func(_ *TestListData) bool {
		return true
	})
	// All elements are still in the list
	p = l.getPointer()
	assert.False(t, p.IsNil())
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

			// Insert many nodes
			for j := 0; j < 10; j++ {
				d := l.Insert(store)
				d.intField = j
			}
			// Filter the i'th one
			l.Filter(store, func(d *TestListData) bool {
				return d.intField != i
			})

			// Show that we removed one of the elements of the list
			p := l.getPointer()
			assert.False(t, p.IsNil())
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
			l := store.NewList()

			// Add the items to the list
			for _, val := range testData.inserts {
				d := l.Insert(store)
				d.intField = val
			}

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

func assertContains(t *testing.T, l *List[TestListData], store *Store[TestListData], expected []int) {
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
