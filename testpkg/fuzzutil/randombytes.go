// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package fuzzutil

import "math/rand"

func MakeRandomTestCases() [][]byte {
	r := rand.New(rand.NewSource(1))
	return [][]byte{
		[]byte{},
		randomBytes(r, 1),
		randomBytes(r, 10),
		randomBytes(r, 50),
		randomBytes(r, 100),
		randomBytes(r, 500),
		randomBytes(r, 1000),
		randomBytes(r, 5000),
		randomBytes(r, 10000),
		randomBytes(r, 50000),
	}
}

func randomBytes(r *rand.Rand, size int) []byte {
	bytes := make([]byte, size)
	r.Read(bytes)
	return bytes
}
