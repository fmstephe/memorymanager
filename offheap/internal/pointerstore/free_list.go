// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package pointerstore

type freeStack struct {
	free []RefPointer
}

func (s *freeStack) push(r RefPointer) {
	s.free = append(s.free, r)
}

func (s *freeStack) pop() (r RefPointer, ok bool) {
	l := len(s.free)
	if l == 0 {
		return RefPointer{}, false
	}

	r = s.free[l-1]
	s.free = s.free[:l-1]
	return r, true
}
