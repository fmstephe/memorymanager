package objectstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// If we add/remove types for testing, update this number (or find a better way
// to manage this)
const numberOfTypes = 12

// A range of differently sized structs.

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
	ref any // Will be of type SizedArray*
}

func (a *MultitypeAllocation) getSlice() []byte {
	ref := a.ref
	switch t := ref.(type) {
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

func allocMultitype(os *Store, idx int) *MultitypeAllocation {
	switch idx % numberOfTypes {
	case 0:
		r, v := Alloc[SizedArray0](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 1:
		r, v := Alloc[SizedArray1](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 2:
		r, v := Alloc[SizedArray2](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 3:
		r, v := Alloc[SizedArray5Small](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 4:
		r, v := Alloc[SizedArray5](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 5:
		r, v := Alloc[SizedArray5Large](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 6:
		r, v := Alloc[SizedArray9Small](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 7:
		r, v := Alloc[SizedArray9](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 8:
		r, v := Alloc[SizedArray9Large](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 9:
		r, v := Alloc[SizedArray14Small](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 10:
		r, v := Alloc[SizedArray14](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	case 11:
		r, v := Alloc[SizedArray14Large](os)
		writeToField(v.Field[:], idx)
		return &MultitypeAllocation{r}
	default:
		panic("unreachable")
	}
}

// Demonstrate that we can create an object, modify that object and when we get
// that object from the store we can see the modifications
// We ensure that we allocate so many objects that we will need more than one slab
// to store all objects.
func Test_Object_NewModifyGet_Multitype(t *testing.T) {
	os := New()
	allocConfs := os.GetAllocationConfigs()
	allocConf := allocConfs[indexForType[SizedArray0]()]
	// perform a number of allocations which will force the creation of extra slabs
	totalAllocations := allocConf.ActualObjectsPerSlab * numberOfTypes * 3

	// Create all the objects and modify field
	allocs := make([]*MultitypeAllocation, totalAllocations)
	for i := range allocs {
		alloc := allocMultitype(os, i)
		allocs[i] = alloc
	}

	sizedStats := os.GetStats()
	stats := sizedStats[indexForType[SizedArray0]()]

	/* TODO re-enable these checks
	assert.Equal(t, len(refs), stats.Allocs)
	assert.Equal(t, len(refs), stats.Live)
	*/
	assert.Equal(t, 0, stats.Frees)

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
	os := New()
	allocConfs := os.GetAllocationConfigs()
	allocConf := allocConfs[indexForType[SizedArray0]()]
	// perform a number of allocations which will force the creation of extra slabs
	totalAllocations := allocConf.ActualObjectsPerSlab * numberOfTypes * 3

	// Create all the objects
	allocs := make([]*MultitypeAllocation, totalAllocations)
	for i := range allocs {
		alloc := allocMultitype(os, i)
		allocs[i] = alloc
	}

	sizedStats := os.GetStats()
	stats := sizedStats[indexForType[SizedArray0]()]

	/* TODO
	assert.Equal(t, len(refs), stats.Allocs)
	assert.Equal(t, len(refs), stats.Live)
	*/
	assert.Equal(t, 0, stats.Frees)

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
