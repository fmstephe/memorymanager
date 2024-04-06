// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package fmath

// uint64 comparisons

// Returns 1 if x > y, 0 otherwise
func UIGT(x, y uint64) uint8 {
	return uint8((y - x) >> 63)
}

// Returns 1 if x >= y, 0 otherwise
func UIGTE(x, y uint64) uint8 {
	return uint8(((x - y) >> 63) ^ 1)
}

// Returns 1 if x < y, 0 otherwise
func UILT(x, y uint64) uint8 {
	return uint8(((x - y) >> 63))
}

// Returns 1 if x <= y, 0 otherwise
func UILTE(x, y uint64) uint8 {
	return uint8(((y - x) >> 63) ^ 1)
}

// Returns 1 if x == 0
// Returns 0 if x == 1
// Undfined for all other inputs
func UINot(x uint64) uint8 {
	return uint8(x ^ 1)
}

// uint8 comparisons

// Returns 1 if x > y, 0 otherwise
func UI8GT(x, y uint8) uint8 {
	return (y - x) >> 7
}

// Returns 1 if x >= y, 0 otherwise
func UI8GTE(x, y uint8) uint8 {
	return ((x - y) >> 7) ^ 1
}

// Returns 1 if x < y, 0 otherwise
func UI8LT(x, y uint8) uint8 {
	return ((x - y) >> 7)
}

// Returns 1 if x <= y, 0 otherwise
func UI8LTE(x, y uint8) uint8 {
	return ((y - x) >> 7) ^ 1
}

// Returns 1 if x == 0
// Returns 0 if x == 1
// Undfined for all other inputs
func UI8Not(x uint8) uint8 {
	return x ^ 1
}
