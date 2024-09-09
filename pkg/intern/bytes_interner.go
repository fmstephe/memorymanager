package intern

type bytesInterner struct {
	interner InternerWithBytesId[BytesConverter]
}

func NewBytesInterner(config Config) Interner[[]byte] {
	return &bytesInterner{
		interner: NewInternerWithBytesId[BytesConverter](config),
	}
}

func (i *bytesInterner) Get(bytes []byte) string {
	return i.interner.Get(NewBytesConverter(bytes))
}

func (i *bytesInterner) GetStats() StatsSummary {
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
