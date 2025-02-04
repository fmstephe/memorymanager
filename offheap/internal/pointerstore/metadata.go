// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

import "fmt"

// If the object's metadata has a non-nil nextFree pointer then the object is
// currently free. Object's which have never been allocated are implicitly
// free, but have a nil nextFree.
//
// A RefPointer smuggles a generation tag. Only references with the same gen
// value can access/free objects they point to. This is a best-effort safety
// check to try to catch use-after-free type errors.
type metadata struct {
	dataAddressAndGen taggedAddress
}

//gcassert:noescape
func (m *metadata) gen() uint8 {
	return m.dataAddressAndGen.gen()
}

//gcassert:noescape
func (m *metadata) setGen(gen uint8) {
	m.dataAddressAndGen = m.dataAddressAndGen.withGen(gen)
}

//gcassert:noescape
func (m *metadata) isFree() bool {
	return m.dataAddressAndGen.isFree()
}

//gcassert:noescape
func (m *metadata) setFree() {
	m.dataAddressAndGen = m.dataAddressAndGen.withFree()
}

//gcassert:noescape
func (m *metadata) setNotFree() {
	m.dataAddressAndGen = m.dataAddressAndGen.withNotFree()
}

// Check that the metadata for a reference agrees with the generation tag.
// Failure results in a panic.
//
// Please note, we never set the is-free bit in the taggedAddress of the
// RefPointer. So we don't expect them to match and don't check that here.
//
//gcassert:noescape
func (m *metadata) checkReference(r *RefPointer) {
	if m.gen() != r.Gen() {
		panic(fmt.Errorf("generation mismatch between metadata (%d) and reference (%d)", m.gen(), r.Gen()))
	}
}
