// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package testutil

import (
	"math/rand"
	"strings"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type RandomStringMaker struct {
	r *rand.Rand
}

func NewRandomStringMaker() *RandomStringMaker {
	return &RandomStringMaker{
		r: rand.New(rand.NewSource(1)),
	}
}

func (rsm *RandomStringMaker) MakeSizedBytes(length int) []byte {
	bytes := make([]byte, 0, length)
	for range length {
		bytes = append(bytes, letters[rsm.r.Intn(len(letters))])
	}
	return bytes
}

func (rsm *RandomStringMaker) MakeSizedString(length int) string {
	builder := strings.Builder{}
	builder.Grow(length)
	for range length {
		builder.WriteByte(letters[rsm.r.Intn(len(letters))])
	}
	return builder.String()
}
