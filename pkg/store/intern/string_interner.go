package intern

import "unsafe"

type StringInterner struct {
	interner InternerWithBytesId[StringConverter]
}

func NewStringInterner(maxLen, maxString int) *StringInterner {
	return &StringInterner{
		interner: NewInternerWithBytesId[StringConverter](maxLen, maxString),
	}
}

func (i *StringInterner) Get(str string) string {
	return i.interner.Get(NewStringConverter(str))
}

func (i *StringInterner) GetStats() StatsSummary {
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
