package intern

type BytesInterner struct {
	interner InternerWithBytesId[BytesConverter]
}

func NewBytesInterner(config Config) *BytesInterner {
	return &BytesInterner{
		interner: NewInternerWithBytesId[BytesConverter](config),
	}
}

func (i *BytesInterner) Get(bytes []byte) string {
	return i.interner.Get(NewBytesConverter(bytes))
}

func (i *BytesInterner) GetStats() StatsSummary {
	return i.interner.GetStats()
}

var _ ConverterWithBytesId = BytesConverter{}

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
