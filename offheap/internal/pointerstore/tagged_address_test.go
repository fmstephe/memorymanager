// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const maxAddress = (1 << (64 - maskShift)) - 1

// For a nil uintptr the address and generation tag are zero
func TestTaggedAddress_ZeroValue(t *testing.T) {
	var ta taggedAddress

	// address and gen are 0
	assert.True(t, ta.isNil())
	assert.Equal(t, uintptr(unsafe.Pointer(nil)), ta.pointer())
	assert.Equal(t, uint8(0), ta.gen())
}

// For a nil uintptr the address and generation tag are zero
func TestTaggedAddress_NewTaggedAddress_Zero(t *testing.T) {
	nilPtr := uintptr(unsafe.Pointer(nil))
	ta := newTaggedAddress(nilPtr)

	// address and gen are 0
	assert.True(t, ta.isNil())
	assert.Equal(t, nilPtr, ta.pointer())
	assert.Equal(t, uint8(0), ta.gen())
}

// For any legal address and generation tag combination the address and
// generation tag are available unaltered
func TestTaggedAddress_AddressAndGenAreSeparate(t *testing.T) {
	for range 10 {
		newPtr := uintptr(rand.Int63n(maxAddress + 1))
		ta := newTaggedAddress(newPtr)

		for i := 1; i <= maxGen; i++ {
			gen := uint8(i)
			// set tagged address generation
			ta = ta.withGen(gen)

			// If the address is not 0 then isNil() must be false
			assert.Equal(t, newPtr == nilPtr, ta.isNil())
			// Assert that the original address is preserved
			assert.Equal(t, newPtr, ta.pointer())
			// Assert that the generation is preserved
			assert.Equal(t, gen, ta.gen())
		}
	}
}

// For any address with bits set in the generation tag region (i.e. upper 8
// bits) newTaggedAddress() panics
func TestTaggedAddress_PointerWithGenBitsPanics(t *testing.T) {
	for range 10 {
		newPtr := uintptr(rand.Int63n(maxAddress + 1))

		for i := 1; i <= maxGen; i++ {
			genBits := uintptr(i)
			// Create a pointer with bits set in the gen tag bits
			badPtr := newPtr | ((genBits) << maskShift)
			assert.Panics(t, func() { newTaggedAddress(badPtr) })
		}
	}
}

func TestTaggedAddress_IsFree(t *testing.T) {
	for range 10 {
		newPtr := uintptr(rand.Int63n(maxAddress + 1))

		ta := newTaggedAddress(newPtr)
		address := ta.pointer()
		gen := ta.gen()
		// A tagged address is created not-free. It is assumed that a
		// tagged address is only created for an allocation
		assert.False(t, ta.isFree())
		// Neither the address nor the generation tag are changed by the is-free bit
		assert.Equal(t, address, ta.pointer())
		assert.Equal(t, gen, ta.gen())

		ta = ta.withFree()
		assert.True(t, ta.isFree())
		// Neither the address nor the generation tag are changed by the is-free bit
		assert.Equal(t, address, ta.pointer())
		assert.Equal(t, gen, ta.gen())

		// Side-quest, set the generation tag and ensure this doesn't
		// interfere with the is-free bit
		ta = ta.withGen(maxGen)
		assert.True(t, ta.isFree())
		// Neither the address nor the generation tag are changed by the is-free bit
		assert.Equal(t, address, ta.pointer())
		assert.Equal(t, uint8(maxGen), ta.gen())

		// Set the generation back
		ta = ta.withGen(gen)

		ta = ta.withNotFree()
		assert.False(t, ta.isFree())
		// Neither the address nor the generation tag are changed by the is-free bit
		assert.Equal(t, address, ta.pointer())
		assert.Equal(t, gen, ta.gen())
	}
}
