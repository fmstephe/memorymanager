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
	isFree            bool
}

//gcassert:noescape
func (m *metadata) gen() uint8 {
	return m.dataAddressAndGen.gen()
}

//gcassert:noescape
func (m *metadata) setGen(gen uint8) {
	m.dataAddressAndGen = m.dataAddressAndGen.withGen(gen)
}

// Check that the metadata for a reference agrees with the generation tag and the data address.
// Failure results in a panic.
//
//gcassert:noescape
func (m *metadata) checkReference(r *RefPointer) {
	if m.dataAddressAndGen != r.dataAddressAndGen {
		mGen := m.gen()
		rGen := r.Gen()
		mAddress := m.dataAddressAndGen.address()
		rAddress := r.dataAddressAndGen.address()

		genMismatchMessage := fmt.Sprintf("generation mismatch between metadata (%d) and reference (%d)", m.gen(), r.Gen())
		addressMismatchMessage := fmt.Sprintf("address mismatch between metadata (%#0*X) and reference (%#0*X)", 14, m.dataAddressAndGen.address(), 14, r.dataAddressAndGen.address())

		switch {
		case mGen != rGen && mAddress != rAddress:
			panic(fmt.Errorf(genMismatchMessage + " and " + addressMismatchMessage))

		case mGen != rGen:
			panic(fmt.Errorf(genMismatchMessage))

		case mAddress != rAddress:
			panic(fmt.Errorf(addressMismatchMessage))
		}
	}
}
