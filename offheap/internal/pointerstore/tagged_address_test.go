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

const addressShift = 64 - maskShift
const maxAddress = (1 << addressShift) - 1

// For a nil uintptr the address and generation tag are zero
func TestTaggedAddress_ZeroValue(t *testing.T) {
	var ta taggedAddress

	// address and gen are 0
	assert.True(t, ta.isNil())
	assert.Equal(t, uintptr(unsafe.Pointer(nil)), ta.address())
	assert.Equal(t, uint8(0), ta.gen())
}

// For a nil uintptr the address and generation tag are zero
func TestTaggedAddress_NewTaggedAddress_Zero(t *testing.T) {
	nilPtr := uintptr(unsafe.Pointer(nil))
	ta := newTaggedAddress(nilPtr)

	// address and gen are 0
	assert.True(t, ta.isNil())
	assert.Equal(t, nilPtr, ta.address())
	assert.Equal(t, uint8(0), ta.gen())
}

// For any legal address and generation tag combination the address and
// generation tag are available unaltered
func TestTaggedAddress_AddressAndGenAreSeparate(t *testing.T) {
	for range 10 {
		newPtr := uintptr(rand.Int63n(maxAddress + 1))
		ta := newTaggedAddress(newPtr)

		for i := 1; i <= 255; i++ {
			gen := uint8(i)
			// set tagged address generation
			ta = ta.withGen(gen)

			// If the address is not 0 then isNil() must be false
			assert.Equal(t, newPtr == 0, ta.isNil())
			// Assert that the original address is preserved
			assert.Equal(t, newPtr, ta.address())
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

		//for genBits := uintptr(1); genBits <= uintptr(255); genBits++ {
		for i := 1; i <= 255; i++ {
			genBits := uintptr(i)
			// Create a pointer with bits set in the gen tag bits
			badPtr := newPtr | ((genBits) << maskShift)
			assert.Panics(t, func() { newTaggedAddress(badPtr) })
		}
	}
}
