package stringstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests are quite sparse and limited. But stringstore.Store is just a
// wrapper around bytestore.Store. So we mostly reply on that package being
// correct and well tested.

// Demonstrate that we can allocate a new string and get that string from the store
func TestAllocAndGet(t *testing.T) {
	ss := New()
	str := "a very good string"

	p, allocStr := ss.Alloc(str)
	assert.Equal(t, str, allocStr)

	getStr := ss.Get(p)
	assert.Equal(t, str, getStr)
}

// Demonstrate that we can alloc a string, then free it. If we try to Get()
// the freed string Store panics
func Test_Strings_NewFreeGet_Panic(t *testing.T) {
	ss := New()
	p, _ := ss.Alloc("be free!")
	ss.Free(p)

	assert.Panics(t, func() { ss.Get(p) })
}

// Demonstrate that we can alloc a string, then free it. If we try to Free()
// the freed string Store panics
func Test_Strings_NewFreeFree_Panic(t *testing.T) {
	ss := New()
	p, _ := ss.Alloc("be free!")
	ss.Free(p)

	assert.Panics(t, func() { ss.Free(p) })
}
