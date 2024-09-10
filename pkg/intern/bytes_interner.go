// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"github.com/fmstephe/memorymanager/pkg/intern/internbase"
)

type bytesInterner struct {
	interner internbase.InternerWithBytesId[BytesConverter]
}

func NewBytesInterner(config internbase.Config) Interner[[]byte] {
	return &bytesInterner{
		interner: internbase.NewInternerWithBytesId[BytesConverter](config),
	}
}

func (i *bytesInterner) Get(bytes []byte) string {
	return i.interner.Get(NewBytesConverter(bytes))
}

func (i *bytesInterner) GetStats() internbase.StatsSummary {
	return i.interner.GetStats()
}

var _ internbase.ConverterWithBytesId = BytesConverter{}

type BytesConverter struct {
	bytes []byte
}

func NewBytesConverter(bytes []byte) BytesConverter {
	return BytesConverter{
		bytes: bytes,
	}
}

func (c BytesConverter) Identity() []byte {
	return c.bytes
}
