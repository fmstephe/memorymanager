// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package intern

import "unsafe"

type stringInterner struct {
	interner InternerWithBytesId[StringConverter]
}

func NewStringInterner(config Config) Interner[string] {
	return &stringInterner{
		interner: NewInternerWithBytesId[StringConverter](config),
	}
}

func (i *stringInterner) Get(str string) string {
	return i.interner.Get(NewStringConverter(str))
}

func (i *stringInterner) GetStats() StatsSummary {
	return i.interner.GetStats()
}

var _ ConverterWithBytesId = StringConverter{}

type StringConverter struct {
	str string
}

func NewStringConverter(str string) StringConverter {
	return StringConverter{
		str: str,
	}
}

func (c StringConverter) Identity() []byte {
	return unsafe.Slice(unsafe.StringData(c.str), len(c.str))
}
