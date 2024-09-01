package intern

import (
	"strconv"
)

type Int64Interner struct {
	interner InternerWithUint64Id[Int64Converter]
	base     int
}

func NewInt64Interner(config Config, base int) *Int64Interner {
	return &Int64Interner{
		interner: NewInternerWithUint64Id[Int64Converter](config),
		base:     base,
	}
}

func (i *Int64Interner) Get(value int64) string {
	return i.interner.Get(NewInt64Converter(value, i.base))
}

func (i *Int64Interner) GetStats() StatsSummary {
	return i.interner.GetStats()
}

var _ ConverterWithUint64Id = Int64Converter{}

// A converter for int64 values. Here the identity is just the value itself.
type Int64Converter struct {
	value int64
	base  int
}

func NewInt64Converter(value int64, base int) Int64Converter {
	return Int64Converter{
		value: value,
		base:  base,
	}
}

func (c Int64Converter) Identity() uint64 {
	return uint64(c.value)
}

func (c Int64Converter) String() string {
	return strconv.FormatInt(c.value, c.base)
}