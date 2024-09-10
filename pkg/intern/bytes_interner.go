// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"github.com/fmstephe/memorymanager/pkg/intern/internbase"
)

type bytesInterner struct {
	interner internbase.InternerWithBytesId[bytesConverter]
}

func NewBytesInterner(config internbase.Config) Interner[[]byte] {
	return &bytesInterner{
		interner: internbase.NewInternerWithBytesId[bytesConverter](config),
	}
}

func (i *bytesInterner) Get(bytes []byte) string {
	return i.interner.Get(newBytesConverter(bytes))
}

func (i *bytesInterner) GetStats() internbase.StatsSummary {
	return i.interner.GetStats()
}

var _ internbase.ConverterWithBytesId = bytesConverter{}

type bytesConverter struct {
	bytes []byte
}

func newBytesConverter(bytes []byte) bytesConverter {
	return bytesConverter{
		bytes: bytes,
	}
}

func (c bytesConverter) Identity() []byte {
	return c.bytes
}
