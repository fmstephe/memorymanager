// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

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
