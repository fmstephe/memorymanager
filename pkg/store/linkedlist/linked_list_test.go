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
	h, n := store.NewList()
	assert.NotNil(t, h)
	assert.NotNil(t, n)
}

func TestLinkedList_SetData(t *testing.T) {
	store := New[TestListData]()
	h, n := store.NewList()
	n.intField = 1
	n.boolField = true
	n.floatField = 1.234

	r := h.Survey(store, func(d *TestListData) bool {
		assert.Equal(t, 1, d.intField)
		assert.Equal(t, true, d.boolField)
		assert.Equal(t, 1.234, d.floatField)
		return true
	})
	assert.True(t, r)
}

func TestLinkedList_SetData_Multiple(t *testing.T) {
	store := New[TestListData]()
	h, _ := store.NewList()
	inserted := map[TestListData]struct{}{}
	for i := 1; i < 11; i++ {
		d := h.Insert(store)
		d.intField = i
		d.floatField = float64(i) / 10
		d.boolField = i%2 == 0
		inserted[*d] = struct{}{}
	}
	surveyed := map[TestListData]struct{}{}
	r := h.Survey(store, func(d *TestListData) bool {
		surveyed[*d] = struct{}{}
		return true
	})
	assert.True(t, r)
	assert.Subset(t, inserted, surveyed)
	assert.Subset(t, surveyed, inserted)
}

func TestLinkedList_AddOneRemoveIt(t *testing.T) {
	store := New[TestListData]()
	h, _ := store.NewList()

	h.Filter(store, func(_ *TestListData) bool {
		return false
	})

	p := h.pointer()
	assert.True(t, p.IsNil())
	assert.Equal(t, 0, h.Len(store))
}

func TestLinkedList_AddManyRemoveAll(t *testing.T) {
	store := New[TestListData]()
	h, _ := store.NewList()

	for i := 1; i < 11; i++ {
		h.Insert(store)
	}
	h.Filter(store, func(_ *TestListData) bool {
		return false
	})

	p := h.pointer()
	assert.True(t, p.IsNil())
	assert.Equal(t, 0, h.Len(store))
}

func TestLinkedList_AddManyRemoveNone(t *testing.T) {
	store := New[TestListData]()
	h, _ := store.NewList()

	for i := 1; i < 11; i++ {
		h.Insert(store)
	}
	h.Filter(store, func(_ *TestListData) bool {
		return true
	})

	p := h.pointer()
	assert.False(t, p.IsNil())
	assert.Equal(t, 11, h.Len(store))
}

func TestLinkedList_AddMany_RemoveOne(t *testing.T) {
	for i := 0; i < 11; i++ {
		t.Run(fmt.Sprintf("Deleting the %dth node", i), func(t *testing.T) {
			store := New[TestListData]()
			h, _ := store.NewList()

			for j := 1; j < 11; j++ {
				d := h.Insert(store)
				d.intField = j
			}
			h.Filter(store, func(d *TestListData) bool {
				return d.intField != i
			})

			p := h.pointer()
			assert.False(t, p.IsNil())
			assert.Equal(t, 10, h.Len(store))
		})
	}
}

func TestLinkedList_Len(t *testing.T) {
	store := New[TestListData]()
	h, _ := store.NewList()
	for i := 1; i < 11; i++ {
		h.Insert(store)
	}
	assert.Equal(t, 11, h.Len(store))
}
