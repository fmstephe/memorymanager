// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"fmt"
	"unsafe"
)

const (
	// handy nil pointer constant
	nilPtr = uintptr(0)
	// maximum value of the generation tag (127)
	maxGen = 0x7F
	// This leaves 8 bits for the generation and isFree tags
	maskShift = 56
	// Highest bit indicates if address is free
	isFreeMask = taggedAddress(0x80 << maskShift)
	// Next 7 bits indicate generation tag
	genMask = taggedAddress(maxGen << maskShift)
	// A generation value of one. This is used to increment generation
	// values
	genOne = taggedAddress(1 << maskShift)
	// Mask revealing all 8 tag bits
	tagMask = isFreeMask | genMask
	// Mask revealing 56 address bits
	addressMask = ^tagMask
	// Mask revealing address and is-free bit, allows us to safely set generation tag
	addressAndIsFreeMask = addressMask | isFreeMask
	// Mask which can be used to directly set the is-free bit to 1
	setFree = isFreeMask
	// Mask which can be used to directly set the is-free bit to 0
	setNotFree = ^setFree
)

type taggedAddress uint64

func newTaggedAddress(ptr uintptr) taggedAddress {
	ta := taggedAddress(ptr)
	if ta.gen() != 0 {
		panic(fmt.Errorf("cannot create tagged address from pointer (%d) with bits set in the generation tag (%d)", ptr, ta.gen()))
	}
	return ta
}

func (a taggedAddress) gen() uint8 {
	return (uint8)((a & genMask) >> maskShift)
}

func (a taggedAddress) pointer() uintptr {
	return uintptr(a & addressMask)
}

func (a taggedAddress) bytes(size int) []byte {
	return ([]byte)(unsafe.Slice((*byte)((unsafe.Pointer)(a.pointer())), size))
}

func (a taggedAddress) isFree() bool {
	return a&setFree == setFree
}

func (a taggedAddress) isNil() bool {
	return a.pointer() == nilPtr
}

func (a taggedAddress) withGen(gen uint8) taggedAddress {
	// move the new gen into postion
	newGen := (taggedAddress(gen) << maskShift) & genMask
	// Set the new generation value back into the address
	return (a & addressAndIsFreeMask) | newGen
}

func (a taggedAddress) withIncGen() taggedAddress {
	// increment the generation value alone
	newGen := (a + genOne) & genMask
	// Set the new generation value back into the address
	return (a & addressAndIsFreeMask) | newGen
}

func (a taggedAddress) withFree() taggedAddress {
	return a | setFree
}

func (a taggedAddress) withNotFree() taggedAddress {
	return a & setNotFree
}
