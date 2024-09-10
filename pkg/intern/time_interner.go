// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package intern

import (
	"time"

	"github.com/fmstephe/memorymanager/pkg/intern/internbase"
)

type timeInterner struct {
	interner internbase.InternerWithUint64Id[timeConverter]
	format   string
}

func NewTimeInterner(config internbase.Config, format string) Interner[time.Time] {
	return &timeInterner{
		interner: internbase.NewInternerWithUint64Id[timeConverter](config),
		format:   format,
	}
}

func (i *timeInterner) Get(value time.Time) string {
	return i.interner.Get(newTimeConverter(value, i.format))
}

func (i *timeInterner) GetStats() internbase.StatsSummary {
	return i.interner.GetStats()
}

var _ internbase.ConverterWithUint64Id = timeConverter{}

// Converter for time.Time. The int64 UnixNano() value is used to uniquely
// identify each time.Time. If time.Time values are used with different time
// zones but which have the same nanosecond values, this converter will
// consider them to be the same and may produce unexpected output.
//
// Having a converter/interner per timezone is currently the best way to handle
// this.
type timeConverter struct {
	value  time.Time
	format string
}

func newTimeConverter(value time.Time, format string) timeConverter {
	return timeConverter{
		value:  value,
		format: format,
	}
}

func (c timeConverter) Identity() uint64 {
	return uint64(c.value.UnixNano())
}

func (c timeConverter) String() string {
	return c.value.Format(c.format)
}
