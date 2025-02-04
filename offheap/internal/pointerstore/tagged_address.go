// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"fmt"
	"unsafe"
)

const nilPtr = uintptr(0)
const maxGen = 0x7F
const maskShift = 56                                // This leaves 8 bits for the generation and isFree tags
const isFreeMask = taggedAddress(0x80 << maskShift) // Highest bit indicates if address is free
const genMask = taggedAddress(maxGen << maskShift)  // Next 7 bits indicate generation tag
const tagMask = isFreeMask | genMask                // Mask revealing all 8 tag bits
const pointerMask = ^tagMask                        // Mask revealing 56 address bits

const setFree = isFreeMask
const setNotFree = ^setFree

type taggedAddress uint64

func newTaggedAddress(ptr uintptr) taggedAddress {
	ta := taggedAddress(ptr)
	if ta.gen() != 0 {
		panic(fmt.Errorf("Cannot create tagged address from pointer (%d) with bits set in the generation tag (%d).", ptr, ta.gen()))
	}
	return ta
}

func (a taggedAddress) gen() uint8 {
	return (uint8)((a & genMask) >> maskShift)
}

func (a taggedAddress) address() uintptr {
	return uintptr(a & pointerMask)
}

func (a taggedAddress) bytes(size int) []byte {
	return ([]byte)(unsafe.Slice((*byte)((unsafe.Pointer)(a.address())), size))
}

func (a taggedAddress) isFree() bool {
	return a&setFree == setFree
}

func (a taggedAddress) isNil() bool {
	return a.address() == nilPtr
}

func (a taggedAddress) withGen(gen uint8) taggedAddress {
	return (a & (pointerMask | isFreeMask)) | (taggedAddress(gen) << maskShift)
}

func (a taggedAddress) withFree() taggedAddress {
	return a | setFree
}

func (a taggedAddress) withNotFree() taggedAddress {
	return a & setNotFree
}
