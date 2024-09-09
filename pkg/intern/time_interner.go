// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package intern

import (
	"time"
)

type timeInterner struct {
	interner InternerWithUint64Id[TimeConverter]
	format   string
}

func NewTimeInterner(config Config, format string) Interner[time.Time] {
	return &timeInterner{
		interner: NewInternerWithUint64Id[TimeConverter](config),
		format:   format,
	}
}

func (i *timeInterner) Get(value time.Time) string {
	return i.interner.Get(NewTimeConverter(value, i.format))
}

func (i *timeInterner) GetStats() StatsSummary {
	return i.interner.GetStats()
}

var _ ConverterWithUint64Id = TimeConverter{}

// Converter for time.Time. The int64 UnixNano() value is used to uniquely
// identify each time.Time. If time.Time values are used with different time
// zones but which have the same nanosecond values, this converter will
// consider them to be the same and may produce unexpected output.
//
// Having a converter/interner per timezone is currently the best way to handle
// this.
type TimeConverter struct {
	value  time.Time
	format string
}

func NewTimeConverter(value time.Time, format string) TimeConverter {
	return TimeConverter{
		value:  value,
		format: format,
	}
}

func (c TimeConverter) Identity() uint64 {
	return uint64(c.value.UnixNano())
}

func (c TimeConverter) String() string {
	return c.value.Format(c.format)
}
