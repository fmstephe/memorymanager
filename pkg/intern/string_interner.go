// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"unsafe"

	"github.com/fmstephe/memorymanager/pkg/intern/internbase"
)

type stringInterner struct {
	interner internbase.InternerWithBytesId[stringConverter]
}

func NewStringInterner(config internbase.Config) Interner[string] {
	return &stringInterner{
		interner: internbase.NewInternerWithBytesId[stringConverter](config),
	}
}

func (i *stringInterner) Get(str string) string {
	return i.interner.Get(newStringConverter(str))
}

func (i *stringInterner) GetStats() internbase.StatsSummary {
	return i.interner.GetStats()
}

var _ internbase.ConverterWithBytesId = stringConverter{}

type stringConverter struct {
	str string
}

func newStringConverter(str string) stringConverter {
	return stringConverter{
		str: str,
	}
}

func (c stringConverter) Identity() []byte {
	return unsafe.Slice(unsafe.StringData(c.str), len(c.str))
}
