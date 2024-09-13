// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"strconv"

	"github.com/fmstephe/memorymanager/pkg/intern/internbase"
)

type int64Interner struct {
	interner internbase.InternerWithUint64Id[int64Converter]
	base     int
}

func NewInt64Interner(config internbase.Config, base int) Interner[int64] {
	return &int64Interner{
		interner: internbase.NewInternerWithUint64Id[int64Converter](config),
		base:     base,
	}
}

func (i *int64Interner) Get(value int64) string {
	return i.interner.Get(newInt64Converter(value, i.base))
}

func (i *int64Interner) GetStats() internbase.StatsSummary {
	return i.interner.GetStats()
}

var _ internbase.ConverterWithUint64Id = int64Converter{}

// A converter for int64 values. Here the identity is just the value itself.
type int64Converter struct {
	value int64
	base  int
}

func newInt64Converter(value int64, base int) int64Converter {
	return int64Converter{
		value: value,
		base:  base,
	}
}

func (c int64Converter) Identity() uint64 {
	return uint64(c.value)
}

func (c int64Converter) String() string {
	return strconv.FormatInt(c.value, c.base)
}
