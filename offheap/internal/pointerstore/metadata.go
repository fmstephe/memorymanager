// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

// A RefPointer smuggles a generation tag. Only references with the same gen
// value can access/free objects they point to. This is a best-effort safety
// check to try to catch use-after-free type errors.
//
// Metadata smuggles a generation tag and an is-free tag into its
// taggedAddress. When accessing the data from a RefPointer we check that the
// metadata indicates the object is not free and that the RefPointer's
// generation matches that of the metadat.
type metadata struct {
	dataAddressAndGen taggedAddress
}

//gcassert:noescape
func (m *metadata) gen() uint8 {
	return m.dataAddressAndGen.gen()
}

//gcassert:noescape
func (m *metadata) incGen() uint8 {
	m.dataAddressAndGen = m.dataAddressAndGen.withIncGen()
	return m.dataAddressAndGen.gen()
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
