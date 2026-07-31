package fuzzhelper

type fillSliceStruct[T any] struct {
	allowableTypes []T `fuzz-ignore`
	Result         []T `fuzz-interface-method:"InterfaceOptions"`
}

func newFillSliceStruct[T any](allowableTypes []T) *fillSliceStruct[T] {
	return &fillSliceStruct[T]{
		allowableTypes: allowableTypes,
	}
}

func (s *fillSliceStruct[T]) InterfaceOptions() []T {
	return s.allowableTypes
}

func MakeSliceOf[T any](allowableTypes []T, bytes []byte) []T {
	fss := newFillSliceStruct[T](allowableTypes)
	Fill(fss, bytes)
	return fss.Result
}
