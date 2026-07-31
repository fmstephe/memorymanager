package fuzzhelper

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

const defaultLengthMin = 0
const defaultLengthMax = 20

type fuzzTags struct {
	// Debugging field containing the fieldName of the struct field the tag was
	// taken from
	fieldName string

	intRange    intTagRange
	uintRange   uintTagRange
	floatRange  floatTagRange
	sliceRange  lengthTagRange
	stringRange lengthTagRange
	mapRange    lengthTagRange

	intValues       methodTag[[]int64]
	uintValues      methodTag[[]uint64]
	floatValues     methodTag[[]float64]
	stringValues    methodTag[[]string]
	interfaceValues methodTag[[]any]
}

func newFuzzTags(structVal reflect.Value, field reflect.StructField) fuzzTags {
	t := newEmptyFuzzTags()

	t.fieldName = field.Name

	t.intRange = newIntTagRange(field, "fuzz-int-range")
	t.uintRange = newUintTagRange(field, "fuzz-uint-range")
	t.floatRange = newFloatTagRange(field, "fuzz-float-range")
	t.stringRange = newLengthTagRangeWithDefault(field, "fuzz-string-range", defaultLengthMin, defaultLengthMax)
	t.sliceRange = newLengthTagRange(field, "fuzz-slice-range")
	t.mapRange = newLengthTagRangeWithDefault(field, "fuzz-map-range", defaultLengthMin, defaultLengthMax)

	t.intValues = newMethodTag[int64](structVal, field, "fuzz-int-method")
	t.uintValues = newMethodTag[uint64](structVal, field, "fuzz-uint-method")
	t.floatValues = newMethodTag[float64](structVal, field, "fuzz-float-method")
	t.stringValues = newMethodTag[string](structVal, field, "fuzz-string-method")
	t.interfaceValues = newMethodTag[any](structVal, field, "fuzz-interface-method")

	return t
}

func newEmptyFuzzTags() fuzzTags {
	return fuzzTags{}
}

type intTagRange struct {
	wasSet bool
	intMin int64
	intMax int64
}

func newIntTagRange(field reflect.StructField, tag string) intTagRange {
	//println(field.Tag)

	valStr, ok := field.Tag.Lookup(tag)
	if !ok {
		//println("no tag found: ", tag, field.Name)
		return intTagRange{}
	}

	parts := strings.Split(valStr, ",")
	if len(parts) != 2 {
		//println("bad min max tag", valStr)
		return intTagRange{}
	}

	minVal, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		//println("bad min tag value", valStr)
		return intTagRange{}
	}

	maxVal, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		//println("bad max tag value", valStr)
		return intTagRange{}
	}

	//println("int64 min max", tag, minVal, maxVal)
	return intTagRange{
		wasSet: true,
		intMin: minVal,
		intMax: maxVal,
	}
}

func (r *intTagRange) fit(val int64) int64 {
	if !r.wasSet {
		return val
	}

	if r.intMax == r.intMin {
		// If min/max are the same the value is clamped to that value
		return r.intMax
	}

	if r.intMax <= r.intMin {
		// Our min/max values are incorrectly set up
		return val
	}
	spread := (r.intMax - r.intMin) + 1

	fitted := (absInt(val) % spread) + r.intMin
	//println("int val fitted", val, r.intMin, r.intMax, fitted)

	return fitted
}

func absInt(val int64) int64 {
	if val == math.MinInt64 {
		// taking -math.MinInt64 produces math.MinInt64
		// So we need to special case this value
		return math.MaxInt64
	}
	if val < 0 {
		return -val
	}
	return val
}

type uintTagRange struct {
	wasSet  bool
	uintMin uint64
	uintMax uint64
}

func newUintTagRange(field reflect.StructField, tag string) uintTagRange {
	//println(field.Tag)

	valStr, ok := field.Tag.Lookup(tag)
	if !ok {
		//println("no tag found: ", tag, field.Name)
		return uintTagRange{}
	}

	parts := strings.Split(valStr, ",")
	if len(parts) != 2 {
		//println("bad min max tag", valStr)
		return uintTagRange{}
	}

	minVal, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		//println("bad min tag value", valStr)
		return uintTagRange{}
	}

	maxVal, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		//println("bad max tag value", valStr)
		return uintTagRange{}
	}

	//println("uint64 min max", tag, minVal, maxVal)
	return uintTagRange{
		wasSet:  true,
		uintMin: minVal,
		uintMax: maxVal,
	}
}

func (r *uintTagRange) fit(val uint64) uint64 {
	if !r.wasSet {
		return val
	}

	if r.uintMax == r.uintMin {
		// If min/max are the same the value is clamped to that value
		return r.uintMax
	}

	if r.uintMax <= r.uintMin {
		// Our min/max values are incorrectly set up
		return val
	}
	spread := (r.uintMax - r.uintMin) + 1

	fitted := (val % spread) + r.uintMin
	//println("uint val fitted", val, r.uintMin, r.uintMax, fitted)

	return fitted
}

type lengthTagRange struct {
	uintRange uintTagRange
}

func newLengthTagRangeWithDefault(field reflect.StructField, tag string, defaultMin, defaultMax uint64) lengthTagRange {
	r := newLengthTagRange(field, tag)
	if !r.uintRange.wasSet {
		r.uintRange = uintTagRange{
			wasSet:  true,
			uintMin: defaultMin,
			uintMax: defaultMax,
		}
	}

	return r
}

func newLengthTagRange(field reflect.StructField, tag string) lengthTagRange {
	return lengthTagRange{
		uintRange: newUintTagRange(field, tag),
	}
}

func (r *lengthTagRange) fit(val int) int {
	if val < 0 {
		return int(r.uintRange.uintMin)
	}

	return int(r.uintRange.fit(uint64(val)))
}

type floatTagRange struct {
	wasSet   bool
	floatMin float64
	floatMax float64
}

func newFloatTagRange(field reflect.StructField, tag string) floatTagRange {
	//println(field.Tag)

	valStr, ok := field.Tag.Lookup(tag)
	if !ok {
		//println("no tag found: ", tag, field.Name)
		return floatTagRange{}
	}

	parts := strings.Split(valStr, ",")
	if len(parts) != 2 {
		//println("bad min max tag", valStr)
		return floatTagRange{}
	}

	minVal, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		//println("bad min tag value", valStr)
		return floatTagRange{}
	}

	maxVal, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		//println("bad max tag value", valStr)
		return floatTagRange{}
	}

	//println("float64 min max", tag, minVal, maxVal)
	return floatTagRange{
		wasSet:   true,
		floatMin: minVal,
		floatMax: maxVal,
	}
}

func (r *floatTagRange) fit(val float64) float64 {
	if !r.wasSet {
		return val
	}

	if r.floatMax == r.floatMin {
		// If min/max are the same the value is clamped to that value
		return r.floatMax
	}

	if r.floatMax <= r.floatMin {
		// Our min/max values are incorrectly set up
		return val
	}
	spread := (r.floatMax - r.floatMin)

	// If val is not-a-number then just take the mid-point between min and max
	if math.IsNaN(val) {
		return r.floatMin + (spread / 2)
	}

	// If val is positive infinity then take max
	if math.IsInf(val, 1) {
		return r.floatMax
	}

	// If val is negative infinity then take min
	if math.IsInf(val, -1) {
		return r.floatMin
	}

	fitted := math.Mod(math.Abs(val), spread) + r.floatMin
	//println("float val fitted", val, r.floatMin, r.floatMax, fitted)

	return fitted
}

type methodTag[T any] struct {
	wasSet     bool
	methodName string
	value      T
}

func newMethodTag[T any](structVal reflect.Value, field reflect.StructField, tag string) methodTag[[]T] {
	methodName, ok := field.Tag.Lookup(tag)
	if !ok {
		//println("no tag found: ", tag, field.Name)
		return methodTag[[]T]{
			wasSet:     false,
			methodName: methodName,
		}
	}

	if !isExported(methodName) {
		//println("method is not exported, can't be called: ", methodName, field.Name, structVal.Type().String())
		panic(fmt.Errorf("%s.%s() is not exported and can't be called", structVal.Type(), methodName))
	}

	// Try to get the method from the struct
	// We look for pointer receiver method first, then value receivers
	// We it in this order under the assumption that people usually use pointer receivers
	method := structVal.Addr().MethodByName(methodName)
	if !method.IsValid() {
		method = structVal.MethodByName(methodName)
		if !method.IsValid() {
			panic(fmt.Errorf("%s.%s() could not be found and can't be called", structVal.Type(), methodName))
		}
	}

	methodType := method.Type()
	if methodType.NumIn() != 0 {
		panic(fmt.Errorf("%s.%s() has arguments (found %d), must have no arguments", structVal.Type(), methodName, methodType.NumIn()))
	}

	if methodType.NumOut() != 1 {
		panic(fmt.Errorf("%s.%s() does not return a single value (found %d), must return a single value", structVal.Type(), methodName, methodType.NumOut()))
	}

	// Get the results from the method call
	result := method.Call([]reflect.Value{})

	// Convert to a slice typed []T - and ensure that every value in the slice can be assigned to the target field
	typedSlice, err := copyToTypedSlice[T](result[0], field.Type)
	if err != nil {
		panic(fmt.Errorf("%s.%s cannot be assigned by every value returned by %s.%s(), %w", structVal.Type(), field.Name, structVal.Type(), methodName, err))
	}

	if len(typedSlice) == 0 {
		panic(fmt.Errorf("%s.%s has options method %s.%s(), but it returns an empty slice", structVal.Type(), field.Name, structVal.Type(), methodName))
	}

	return methodTag[[]T]{
		wasSet:     true,
		methodName: methodName,
		value:      typedSlice,
	}
}

func copyToTypedSlice[T any](srcSlice reflect.Value, assignType reflect.Type) ([]T, error) {
	if srcSlice.Kind() != reflect.Slice {
		return nil, fmt.Errorf("method must return a slice, but it returns %s", srcSlice.Kind().String())
	}

	result := make([]T, srcSlice.Len())
	resultVal := reflect.ValueOf(result)

	for i := range result {
		elem := srcSlice.Index(i)
		if err := isAssignable(elem, assignType); err != nil {
			return nil, err
		}

		assign(elem, resultVal.Index(i))
	}

	return result, nil
}

func isAssignable(value reflect.Value, assignType reflect.Type) error {
	if assignType.Kind() == reflect.Slice {
		// If we are assigning to a slice, we want to know if we can
		// append value to that slice. Get the type of the slice
		// elements.
		// TODO, we probably need to do this for maps as well
		assignType = assignType.Elem()
	}

	if assignType.Kind() == reflect.Interface && value.Kind() == reflect.Interface && value.Elem().Kind() != reflect.Pointer {
		// We restrict all interface assignments to pointer types only.
		// This is done _purely_ to simplify the process of determining
		// whether a type satisfies an interface
		return fmt.Errorf("value of type %s must be a pointer (not a value type) to assign to interface type %s", value.Elem().Type(), assignType)
	}

	if !value.Type().AssignableTo(assignType) {
		return fmt.Errorf("value of type %s cannot be assigned to %s", value.Type(), assignType)
	}

	return nil
}

// Here do type conversion for storing in the tag options slices
// The tag slices are of type int64, uint64, float64, string or any
//
// Many of the values we are assigning here actually have narrower specific
// types. For example the options might be providing a range of int32 values or
// an interface more specific than any.  In the function above we checked that
// the value was assignable to its destination field, here we assign this
// narrow type to one of the tag slice types. So we need to perform some type
// conversion, like widing a numeric type or converting a specific interface
// type to any. Strings are fine as is and don't need any type conversion.
//
// This works in practice because our conversions always preserve the original
// type and so the wider values can be used to set the field without loss of
// information.
func assign(value, dest reflect.Value) {
	switch dest.Kind() {
	case reflect.Int64:
		dest.SetInt(value.Int())
	case reflect.Uint64:
		dest.SetUint(value.Uint())
	case reflect.Float64:
		dest.SetFloat(value.Float())
	case reflect.String:
		dest.Set(value)
	case reflect.Interface:
		dest.Set(reflect.ValueOf(value.Interface()))
	default:
		panic(fmt.Errorf("unsupported assignment found: %s", dest.Kind()))
	}
}
