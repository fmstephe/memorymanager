package linkedlist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestListData struct {
	intField   int
	boolField  bool
	floatField float64
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
	n.boolField = true
	n.floatField = 1.234

	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		assert.Equal(t, 1, d.intField)
		assert.Equal(t, true, d.boolField)
		assert.Equal(t, 1.234, d.floatField)
		return true
	}))
}

// Show that we can add many nodes to the list and set all of their embedded
// data to different values
func TestLinkedList_SetData_Multiple(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// Insert new nodes, collect copy of the inserted data
	inserted := []TestListData{}
	for i := 0; i < 10; i++ {
		d := l.Insert(store)
		d.intField = i
		d.floatField = float64(i) / 10
		d.boolField = i%2 == 0
		inserted = append(inserted, *d)
	}

	// Survey the list and collect copy of surveyed data
	surveyed := []TestListData{}
	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		surveyed = append(surveyed, *d)
		return true
	}))

	// Inserted and surveyed data must match
	assert.ElementsMatch(t, inserted, surveyed)
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
		d.floatField = float64(i) / 10
		d.boolField = i%2 == 0
	}

	// Survey the list, updating the embedded data of each node. Collect a
	// copy of the updated data value
	updated := []TestListData{}
	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		// modify d
		d.intField *= 2
		d.boolField = !d.boolField
		d.floatField *= 2
		// Store the updated version
		updated = append(updated, *d)
		return true
	}))

	// Survey the list, collecting a copy of the embedded data values
	surveyed := []TestListData{}
	assert.True(t, l.Survey(store, func(d *TestListData) bool {
		surveyed = append(surveyed, *d)
		return true
	}))

	// The updated and surveyed data values must match
	assert.ElementsMatch(t, updated, surveyed)
}

// Show that we can add an element to the list and then remove it using Filter
func TestLinkedList_AddOneRemoveIt(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	// Insert a single element into the list
	l.Insert(store)
	p := l.getPointer()
	assert.False(t, p.IsNil())
	assert.Equal(t, 1, l.Len(store))

	// Delete the element from the list
	l.Filter(store, func(_ *TestListData) bool {
		return false
	})
	p = l.getPointer()
	assert.True(t, p.IsNil())
	assert.Equal(t, 0, l.Len(store))
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
