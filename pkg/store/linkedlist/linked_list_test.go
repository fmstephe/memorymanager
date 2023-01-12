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

func TestLinkedList_New(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()
	assert.NotNil(t, l)
}

func TestLinkedList_SetData(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()
	n := l.Insert(store)
	n.intField = 1
	n.boolField = true
	n.floatField = 1.234

	r := l.Survey(store, func(d *TestListData) bool {
		assert.Equal(t, 1, d.intField)
		assert.Equal(t, true, d.boolField)
		assert.Equal(t, 1.234, d.floatField)
		return true
	})
	assert.True(t, r)
}

func TestLinkedList_SetData_Multiple(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()
	inserted := []TestListData{}
	for i := 0; i < 10; i++ {
		d := l.Insert(store)
		d.intField = i
		d.floatField = float64(i) / 10
		d.boolField = i%2 == 0
		inserted = append(inserted, *d)
	}
	surveyed := []TestListData{}
	r := l.Survey(store, func(d *TestListData) bool {
		surveyed = append(surveyed, *d)
		return true
	})
	assert.True(t, r)
	assert.ElementsMatch(t, inserted, surveyed)
}

func TestLinkedList_SetData_ModifyAfterInsert(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()
	for i := 0; i < 10; i++ {
		d := l.Insert(store)
		d.intField = i
		d.floatField = float64(i) / 10
		d.boolField = i%2 == 0
	}
	updated := []TestListData{}
	r := l.Survey(store, func(d *TestListData) bool {
		// modify d
		d.intField *= 2
		d.boolField = !d.boolField
		d.floatField *= 2
		// Store the updated version
		updated = append(updated, *d)
		return true
	})
	assert.True(t, r)
	surveyed := []TestListData{}
	r = l.Survey(store, func(d *TestListData) bool {
		surveyed = append(surveyed, *d)
		return true
	})
	assert.True(t, r)
	assert.ElementsMatch(t, updated, surveyed)
}

func TestLinkedList_AddOneRemoveIt(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	l.Filter(store, func(_ *TestListData) bool {
		return false
	})

	p := l.getPointer()
	assert.True(t, p.IsNil())
	assert.Equal(t, 0, l.Len(store))
}

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

func TestLinkedList_AddManyRemoveAll_AddOne(t *testing.T) {
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

	l.Insert(store)

	p = l.getPointer()
	assert.False(t, p.IsNil())
	assert.Equal(t, 1, l.Len(store))
}

func TestLinkedList_AddManyRemoveNone(t *testing.T) {
	store := New[TestListData]()
	l := store.NewList()

	for i := 0; i < 10; i++ {
		l.Insert(store)
	}
	l.Filter(store, func(_ *TestListData) bool {
		return true
	})

	p := l.getPointer()
	assert.False(t, p.IsNil())
	assert.Equal(t, 10, l.Len(store))
}

func TestLinkedList_AddMany_RemoveOne(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("Deleting the %dth node", i), func(t *testing.T) {
			store := New[TestListData]()
			l := store.NewList()

			for j := 0; j < 10; j++ {
				d := l.Insert(store)
				d.intField = j
			}
			l.Filter(store, func(d *TestListData) bool {
				return d.intField != i
			})

			p := l.getPointer()
			assert.False(t, p.IsNil())
			assert.Equal(t, 9, l.Len(store))
		})
	}
}
