// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import "github.com/fmstephe/memorymanager/pkg/intern/internbase"

type Interner[T any] interface {
	Get(t T) string
	GetStats() internbase.StatsSummary
}
