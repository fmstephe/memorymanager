// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// If we add/remove types for testing, update this number (or find a better way
// to manage this)
const numberOfTypes = 15

// A range of differently sized structs.

type SizedArrayZero struct {
	Field [0]byte
}

type SizedArray0 struct {
	Field [1]byte
}

type SizedArray1 struct {
	Field [1 << 1]byte
}

type SizedArray2Small struct {
	Field [(1 << 2) - 1]byte
}

type SizedArray2 struct {
	Field [1 << 2]byte
}

type SizedArray2Large struct {
	Field [(1 << 2) + 1]byte
}

type SizedArray5Small struct {
	Field [(1 << 5) - 1]byte
}

type SizedArray5 struct {
	Field [1 << 5]byte
}

type SizedArray5Large struct {
	Field [(1 << 5) + 1]byte
}

type SizedArray9Small struct {
	Field [(1 << 9) - 1]byte
}

type SizedArray9 struct {
	Field [1 << 9]byte
}

type SizedArray9Large struct {
	Field [(1 << 9) + 1]byte
}

type SizedArray14Small struct {
	Field [(1 << 14) - 1]byte
}

type SizedArray14 struct {
	Field [1 << 14]byte
}

type SizedArray14Large struct {
	Field [(1 << 14) + 1]byte
}

type MultitypeAllocation struct {
	ref any // Will be of type Reference[SizedArray*]
}

func (a *MultitypeAllocation) getSlice() []byte {
	ref := a.ref
	switch t := ref.(type) {
	case RefObject[SizedArrayZero]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray0]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray1]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray2Small]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray2]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray2Large]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray5Small]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray5]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray5Large]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray9Small]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray9]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray9Large]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray14Small]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray14]:
		v := t.Value()
		return v.Field[:]
	case RefObject[SizedArray14Large]:
		v := t.Value()
		return v.Field[:]
	default:
		panic(fmt.Errorf("Bad type %+v", t))
	}
}

func (a *MultitypeAllocation) free(s *Store) {
	ref := a.ref
	switch t := ref.(type) {
	case RefObject[SizedArrayZero]:
		s.FreeObject(t)
	case RefObject[SizedArray0]:
		s.FreeObject(t)
	case RefObject[SizedArray1]:
		s.FreeObject(t)
	case RefObject[SizedArray2Small]:
		s.FreeObject(t)
	case RefObject[SizedArray2]:
		s.FreeObject(t)
	case RefObject[SizedArray2Large]:
		s.FreeObject(t)
	case RefObject[SizedArray5Small]:
		s.FreeObject(t)
	case RefObject[SizedArray5]:
		s.FreeObject(t)
	case RefObject[SizedArray5Large]:
		s.FreeObject(t)
	case RefObject[SizedArray9Small]:
		s.FreeObject(t)
	case RefObject[SizedArray9]:
		s.FreeObject(t)
	case RefObject[SizedArray9Large]:
		s.FreeObject(t)
	case RefObject[SizedArray14Small]:
		s.FreeObject(t)
	case RefObject[SizedArray14]:
		s.FreeObject(t)
	case RefObject[SizedArray14Large]:
		s.FreeObject(t)
	default:
		panic(fmt.Errorf("Bad type %+v", t))
	}
}

func multitypeAllocFunc(selector uint) func(*Store) *MultitypeAllocation {
	switch selector % numberOfTypes {
	case 0:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArrayZero]()
			return &MultitypeAllocation{r}
		}
	case 1:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray0]()
			return &MultitypeAllocation{r}
		}
	case 2:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray1]()
			return &MultitypeAllocation{r}
		}
	case 3:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray2Small]()
			return &MultitypeAllocation{r}
		}
	case 4:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray2]()
			return &MultitypeAllocation{r}
		}
	case 5:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray2Large]()
			return &MultitypeAllocation{r}
		}
	case 6:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray5Small]()
			return &MultitypeAllocation{r}
		}
	case 7:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray5]()
			return &MultitypeAllocation{r}
		}
	case 8:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray5Large]()
			return &MultitypeAllocation{r}
		}
	case 9:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray9Small]()
			return &MultitypeAllocation{r}
		}
	case 10:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray9]()
			return &MultitypeAllocation{r}
		}
	case 11:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray9Large]()
			return &MultitypeAllocation{r}
		}
	case 12:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray14Small]()
			return &MultitypeAllocation{r}
		}
	case 13:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray14]()
			return &MultitypeAllocation{r}
		}
	case 14:
		return func(os *Store) *MultitypeAllocation {
			r := os.AllocObject[SizedArray14Large]()
			return &MultitypeAllocation{r}
		}
	default:
		panic("unreachable")
	}
}

func allocAndWrite(os *Store, selector uint) *MultitypeAllocation {
	allocFunc := multitypeAllocFunc(selector)
	allocation := allocFunc(os)
	allocSlice := allocation.getSlice()
	writeToField(allocSlice, byte(selector))
	return allocation
}

func TestIndexForType(t *testing.T) {
	assert.Equal(t, 0, indexForType[SizedArrayZero](), "SizedArray0 %d", sizeForType[SizedArray0]())
	assert.Equal(t, 0, indexForType[SizedArray0](), "SizedArray0 %d", sizeForType[SizedArray0]())
	assert.Equal(t, 1, indexForType[SizedArray1](), "SizedArray1 %d", sizeForType[SizedArray1]())
	assert.Equal(t, 2, indexForType[SizedArray2Small](), "SizedArray2 %d", sizeForType[SizedArray2]())
	assert.Equal(t, 2, indexForType[SizedArray2](), "SizedArray2 %d", sizeForType[SizedArray2]())
	assert.Equal(t, 3, indexForType[SizedArray2Large](), "SizedArray2 %d", sizeForType[SizedArray2]())
	assert.Equal(t, 5, indexForType[SizedArray5Small](), "SizedArray5Small %d", sizeForType[SizedArray5Small]())
	assert.Equal(t, 5, indexForType[SizedArray5](), "SizedArray5 %d", sizeForType[SizedArray5]())
	assert.Equal(t, 6, indexForType[SizedArray5Large](), "SizedArray5Large %d", sizeForType[SizedArray5Large]())
	assert.Equal(t, 9, indexForType[SizedArray9Small](), "SizedArray9Small %d", sizeForType[SizedArray9Small]())
	assert.Equal(t, 9, indexForType[SizedArray9](), "SizedArray9 %d", sizeForType[SizedArray9]())
	assert.Equal(t, 10, indexForType[SizedArray9Large](), "SizedArray9Large %d", sizeForType[SizedArray9Large]())
	assert.Equal(t, 14, indexForType[SizedArray14Small](), "SizedArray14Small %d", sizeForType[SizedArray14Small]())
	assert.Equal(t, 14, indexForType[SizedArray14](), "SizedArray14 %d", sizeForType[SizedArray14]())
	assert.Equal(t, 15, indexForType[SizedArray14Large](), "SizedArray14Large %d", sizeForType[SizedArray14Large]())
}

// These tests are a bit fragile, as we have to _carefully_ only allocate
// objects of each size class only once. Because we track the number of slabs
// allocated as well as raw/reused allocations asserting the correct metrics
// quickly becomes difficult when we exercise the same size class multiple
// times.
func TestSizedStats(t *testing.T) {
	os := New()
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	testSizedStats[SizedArrayZero](t, os)
	testSizedStats[SizedArray1](t, os)
	testSizedStats[SizedArray2](t, os)
	testSizedStats[SizedArray2Large](t, os)
	testSizedStats[SizedArray5](t, os)
	testSizedStats[SizedArray5Large](t, os)
	testSizedStats[SizedArray9](t, os)
	testSizedStats[SizedArray9Large](t, os)
	testSizedStats[SizedArray14](t, os)
	testSizedStats[SizedArray14Large](t, os)
}

func testSizedStats[T any](t *testing.T, os *Store) {
	expectedStats := StatsForType[T](os)

	r1 := os.AllocObject[T]()
	r2 := os.AllocObject[T]()
	os.FreeObject(r1)
	r3 := os.AllocObject[T]()
	os.FreeObject(r2)
	os.FreeObject(r3)

	expectedStats.Allocs = 3
	expectedStats.Frees = 3
	expectedStats.RawAllocs = 2
	expectedStats.Reused = 1

	conf := ConfForType[T](os)

	if conf.ObjectsPerSlab > 1 {
		// Only expect one slab to be allocated for smaller objects
		expectedStats.Slabs = 1
	} else {
		// Larger objects will require a slab per allocation
		expectedStats.Slabs = 2
	}

	actualStats := StatsForType[T](os)

	assert.Equal(t, expectedStats, actualStats)
}

// Demonstrate that we can create an object, modify that object and when we get
// that object from the store we can see the modifications
// We ensure that we allocate so many objects that we will need more than one slab
// to store all objects.
func Test_Object_NewModifyGet_Multitype(t *testing.T) {
	os := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	allocConf := ConfForType[SizedArray0](os)
	// perform a number of allocations which will force the creation of extra slabs
	totalAllocations := allocConf.ObjectsPerSlab * numberOfTypes * 3

	// Create all the objects and modify field
	allocs := make([]*MultitypeAllocation, totalAllocations)
	for i := range allocs {
		alloc := allocAndWrite(os, uint(i))
		allocs[i] = alloc
	}

	// Assert that all of the modifications are visible
	for i, alloc := range allocs {
		s := alloc.getSlice()
		assert.Equal(t, generateField(len(s), byte(i)), s)
	}
}

// Demonstrate that we can create an object, then get that object and modify it
// we can then get that object again and will see the modification
// We ensure that we allocate so many objects that we will need more than one slab
// to store all objects.
func Test_Object_GetModifyGet_Multitype(t *testing.T) {
	os := NewSized(1 << 8)
	defer func() {
		assert.NoError(t, os.Destroy())
	}()

	allocConf := ConfForType[SizedArray0](os)
	// perform a number of allocations which will force the creation of extra slabs
	totalAllocations := allocConf.ObjectsPerSlab * numberOfTypes * 3

	// Create all the objects
	allocs := make([]*MultitypeAllocation, totalAllocations)
	for i := range allocs {
		alloc := allocAndWrite(os, uint(i))
		allocs[i] = alloc
	}

	// Get each object and modify field
	for i, alloc := range allocs {
		s := alloc.getSlice()
		writeToField(s, byte(i*2))
	}

	// Assert that all of the modifications are visible
	for i, alloc := range allocs {
		s := alloc.getSlice()
		assert.Equal(t, generateField(len(s), byte(i*2)), s)
	}
}

func writeToField(field []byte, value byte) {
	for i := range field {
		field[i] = value
	}
}

func generateField(size int, value byte) []byte {
	field := make([]byte, size)
	writeToField(field, value)
	return field
}
