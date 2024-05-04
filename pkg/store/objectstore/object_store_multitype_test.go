package objectstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// If we add/remove types for testing, update this number (or find a better way
// to manage this)
const numberOfTypes = 13

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

type SizedArray2 struct {
	Field [1 << 2]byte
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
	case Reference[SizedArrayZero]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray0]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray1]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray2]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray5Small]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray5]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray5Large]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray9Small]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray9]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray9Large]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray14Small]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray14]:
		v := t.GetValue()
		return v.Field[:]
	case Reference[SizedArray14Large]:
		v := t.GetValue()
		return v.Field[:]
	default:
		panic(fmt.Errorf("Bad type %+v", t))
	}
}

func (a *MultitypeAllocation) free(s *Store) {
	ref := a.ref
	switch t := ref.(type) {
	case Reference[SizedArrayZero]:
		Free[SizedArrayZero](s, t)
	case Reference[SizedArray0]:
		Free[SizedArray0](s, t)
	case Reference[SizedArray1]:
		Free[SizedArray1](s, t)
	case Reference[SizedArray2]:
		Free[SizedArray2](s, t)
	case Reference[SizedArray5Small]:
		Free[SizedArray5Small](s, t)
	case Reference[SizedArray5]:
		Free[SizedArray5](s, t)
	case Reference[SizedArray5Large]:
		Free[SizedArray5Large](s, t)
	case Reference[SizedArray9Small]:
		Free[SizedArray9Small](s, t)
	case Reference[SizedArray9]:
		Free[SizedArray9](s, t)
	case Reference[SizedArray9Large]:
		Free[SizedArray9Large](s, t)
	case Reference[SizedArray14Small]:
		Free[SizedArray14Small](s, t)
	case Reference[SizedArray14]:
		Free[SizedArray14](s, t)
	case Reference[SizedArray14Large]:
		Free[SizedArray14Large](s, t)
	default:
		panic(fmt.Errorf("Bad type %+v", t))
	}
}

func multitypeAllocFunc(selector int) func(*Store) *MultitypeAllocation {
	switch selector % numberOfTypes {
	case 0:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArrayZero](os)
			return &MultitypeAllocation{r}
		}
	case 1:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray0](os)
			return &MultitypeAllocation{r}
		}
	case 2:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray1](os)
			return &MultitypeAllocation{r}
		}
	case 3:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray2](os)
			return &MultitypeAllocation{r}
		}
	case 4:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray5Small](os)
			return &MultitypeAllocation{r}
		}
	case 5:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray5](os)
			return &MultitypeAllocation{r}
		}
	case 6:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray5Large](os)
			return &MultitypeAllocation{r}
		}
	case 7:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray9Small](os)
			return &MultitypeAllocation{r}
		}
	case 8:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray9](os)
			return &MultitypeAllocation{r}
		}
	case 9:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray9Large](os)
			return &MultitypeAllocation{r}
		}
	case 10:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray14Small](os)
			return &MultitypeAllocation{r}
		}
	case 11:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray14](os)
			return &MultitypeAllocation{r}
		}
	case 12:
		return func(os *Store) *MultitypeAllocation {
			r, _ := Alloc[SizedArray14Large](os)
			return &MultitypeAllocation{r}
		}
	default:
		panic("unreachable")
	}
}

func allocAndWrite(os *Store, selector int) *MultitypeAllocation {
	allocFunc := multitypeAllocFunc(selector)
	allocation := allocFunc(os)
	allocSlice := allocation.getSlice()
	writeToField(allocSlice, selector)
	return allocation
}

func TestIndexForType(t *testing.T) {
	assert.Equal(t, 0, indexForType[SizedArrayZero](), "SizedArray0 %d", sizeForType[SizedArray0]())
	assert.Equal(t, 0, indexForType[SizedArray0](), "SizedArray0 %d", sizeForType[SizedArray0]())
	assert.Equal(t, 1, indexForType[SizedArray1](), "SizedArray1 %d", sizeForType[SizedArray1]())
	assert.Equal(t, 2, indexForType[SizedArray2](), "SizedArray2 %d", sizeForType[SizedArray2]())
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
	defer os.Destroy()

	testSizedStats[SizedArrayZero](t, os)
	testSizedStats[SizedArray1](t, os)
	testSizedStats[SizedArray2](t, os)
	testSizedStats[SizedArray5](t, os)
	testSizedStats[SizedArray5Large](t, os)
	testSizedStats[SizedArray9](t, os)
	testSizedStats[SizedArray9Large](t, os)
	testSizedStats[SizedArray14](t, os)
	testSizedStats[SizedArray14Large](t, os)
}

func testSizedStats[T any](t *testing.T, os *Store) {
	expectedStats := StatsForTypeSize[T](os)

	r1, _ := Alloc[T](os)
	r2, _ := Alloc[T](os)
	Free[T](os, r1)
	r3, _ := Alloc[T](os)
	Free[T](os, r2)
	Free[T](os, r3)

	expectedStats.Allocs = 3
	expectedStats.Frees = 3
	expectedStats.RawAllocs = 2
	expectedStats.Reused = 1

	conf := ConfForTypeSize[T](os)

	if conf.ObjectsPerSlab > 1 {
		// Only expect one slab to be allocated for smaller objects
		expectedStats.Slabs = 1
	} else {
		// Larger objects will require a slab per allocation
		expectedStats.Slabs = 2
	}

	actualStats := StatsForTypeSize[T](os)

	assert.Equal(t, expectedStats, actualStats)
}

// Demonstrate that we can create an object, modify that object and when we get
// that object from the store we can see the modifications
// We ensure that we allocate so many objects that we will need more than one slab
// to store all objects.
func Test_Object_NewModifyGet_Multitype(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	allocConf := ConfForTypeSize[SizedArray0](os)
	// perform a number of allocations which will force the creation of extra slabs
	totalAllocations := allocConf.ObjectsPerSlab * numberOfTypes * 3

	// Create all the objects and modify field
	allocs := make([]*MultitypeAllocation, totalAllocations)
	for i := range allocs {
		alloc := allocAndWrite(os, i)
		allocs[i] = alloc
	}

	// Assert that all of the modifications are visible
	for i, alloc := range allocs {
		s := alloc.getSlice()
		assert.Equal(t, generateField(len(s), i), s)
	}
}

// Demonstrate that we can create an object, then get that object and modify it
// we can then get that object again and will see the modification
// We ensure that we allocate so many objects that we will need more than one slab
// to store all objects.
func Test_Object_GetModifyGet_Multitype(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	allocConf := ConfForTypeSize[SizedArray0](os)
	// perform a number of allocations which will force the creation of extra slabs
	totalAllocations := allocConf.ObjectsPerSlab * numberOfTypes * 3

	// Create all the objects
	allocs := make([]*MultitypeAllocation, totalAllocations)
	for i := range allocs {
		alloc := allocAndWrite(os, i)
		allocs[i] = alloc
	}

	// Get each object and modify field
	for i, alloc := range allocs {
		s := alloc.getSlice()
		writeToField(s, i*2)
	}

	// Assert that all of the modifications are visible
	for i, alloc := range allocs {
		s := alloc.getSlice()
		assert.Equal(t, generateField(len(s), i*2), s)
	}
}

func writeToField(field []byte, value int) {
	for i := range field {
		field[i] = byte(value)
	}
}

func generateField(size int, value int) []byte {
	field := make([]byte, size)
	writeToField(field, value)
	return field
}