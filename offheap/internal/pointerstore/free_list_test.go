// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFreeStack_Empty(t *testing.T) {
	s := freeStack{}
	r, ok := s.pop()
	assert.True(t, r.IsNil())
	assert.False(t, ok)
}

// Push 3 references onto the stack. Pop them off in reverse order
func TestFreeStack_PushPop(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadatas := MmapSlab(allocConfig)

	s := freeStack{}

	r1Push := NewReference(objects[0], metadatas[0])
	r2Push := NewReference(objects[1], metadatas[1])
	r3Push := NewReference(objects[2], metadatas[2])

	s.push(r1Push)
	s.push(r2Push)
	s.push(r3Push)

	r3Pop, ok3 := s.pop()
	assert.Equal(t, r3Push, r3Pop)
	assert.True(t, ok3)

	r2Pop, ok2 := s.pop()
	assert.Equal(t, r2Push, r2Pop)
	assert.True(t, ok2)

	r1Pop, ok1 := s.pop()
	assert.Equal(t, r1Push, r1Pop)
	assert.True(t, ok1)

	r0Pop, ok0 := s.pop()
	assert.True(t, r0Pop.IsNil())
	assert.False(t, ok0)
}

// Push three references off, pop two, push two more. Pop remaining references.
func TestFreeStack_PushPopComplex(t *testing.T) {
	allocConfig := NewAllocConfigBySize(8, 32*8)
	objects, metadatas := MmapSlab(allocConfig)

	s := freeStack{}

	r1Push := NewReference(objects[0], metadatas[0])
	r2Push := NewReference(objects[1], metadatas[1])
	r3Push := NewReference(objects[2], metadatas[2])
	r4Push := NewReference(objects[3], metadatas[3])
	r5Push := NewReference(objects[4], metadatas[4])

	s.push(r1Push)
	s.push(r2Push)
	s.push(r3Push)

	r3Pop, ok3 := s.pop()
	assert.Equal(t, r3Push, r3Pop)
	assert.True(t, ok3)

	r2Pop, ok2 := s.pop()
	assert.Equal(t, r2Push, r2Pop)
	assert.True(t, ok2)

	s.push(r4Push)
	s.push(r5Push)

	r5Pop, ok5 := s.pop()
	assert.Equal(t, r5Push, r5Pop)
	assert.True(t, ok5)

	r4Pop, ok4 := s.pop()
	assert.Equal(t, r4Push, r4Pop)
	assert.True(t, ok4)

	r1Pop, ok1 := s.pop()
	assert.Equal(t, r1Push, r1Pop)
	assert.True(t, ok1)

	r0Pop, ok0 := s.pop()
	assert.True(t, r0Pop.IsNil())
	assert.False(t, ok0)
}
