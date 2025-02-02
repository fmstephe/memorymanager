// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import (
	"fmt"
	"unsafe"
)

const maskShift = 56 // This leaves 8 bits for the generation data
const genMask = taggedAddress(0xFF << maskShift)
const pointerMask = ^genMask

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

func (a taggedAddress) withGen(gen uint8) taggedAddress {
	return (a & pointerMask) | (taggedAddress(gen) << maskShift)
}

func (a taggedAddress) bytes(size int) []byte {
	return ([]byte)(unsafe.Slice((*byte)((unsafe.Pointer)(a.address())), size))
}

func (a taggedAddress) isNil() bool {
	return a.address() == 0
}
