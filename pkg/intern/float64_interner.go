package intern

import (
	"math"
	"strconv"
)

type Float64Interner struct {
	interner InternerWithUint64Id[Float64Converter]
	fmt      byte
	prec     int
	bitSize  int
}

func NewFloat64Interner(config Config, fmt byte, prec, bitSize int) *Float64Interner {
	return &Float64Interner{
		interner: NewInternerWithUint64Id[Float64Converter](config),
		fmt:      fmt,
		prec:     prec,
		bitSize:  bitSize,
	}
}

func (i *Float64Interner) Get(value float64) string {
	return i.interner.Get(NewFloat64Converter(value, i.fmt, i.prec, i.bitSize))
}

func (i *Float64Interner) GetStats() StatsSummary {
	return i.interner.GetStats()
}

var _ ConverterWithUint64Id = Float64Converter{}

// A flexible converter for float64 values. Here the identity is generated by a
// call to math.Float64bits(...) and we convert the value into a string using
// strconv.FormatFloat(...)
type Float64Converter struct {
	value   float64
	fmt     byte
	prec    int
	bitSize int
}

func NewFloat64Converter(value float64, fmt byte, prec, bitSize int) Float64Converter {
	return Float64Converter{
		value:   value,
		fmt:     fmt,
		prec:    prec,
		bitSize: bitSize,
	}
}

func (c Float64Converter) Identity() uint64 {
	return math.Float64bits(c.value)
}

func (c Float64Converter) String() string {
	return strconv.FormatFloat(c.value, c.fmt, c.prec, c.bitSize)
}