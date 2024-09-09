// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap

// This type constraint allows us to build generic types which accept any
// Reference type.  There is an awkward problem if your type is RefString,
// because you are forced to include an _unused_ parameterised type. We may
// learn to live with this peacefully in time.
type Reference[T any] interface {
	RefString | RefSlice[T] | RefObject[T]
}
