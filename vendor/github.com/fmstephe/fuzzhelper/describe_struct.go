package fuzzhelper

import (
	"fmt"
	"os"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var _ valueVisitor = &describeVisitor{}

type describeVisitor struct {
}

func Describe(root any) {
	visitRoot(&describeVisitor{}, root, newByteConsumer([]byte{1, 2, 3}))
}

func shortenString(s string) string {
	const limit = 20
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}

func methodValuesString[T any](v []T) string {
	vStr := ""
	if len(v) == 0 {
		return vStr
	}
	return shortenString(fmt.Sprintf("%v", v))
}

func isExported(name string) bool {
	if name == "" {
		return false
	}

	firstRune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(firstRune)
}

func introDescription(value reflect.Value, tags fuzzTags, path valuePath) {
	fmt.Fprintf(os.Stdout, "%s\n", path.pathString(value))

	// Value is settable
	if value.CanSet() {
		return
	}

	// Field is not settable, lets find out why

	// If this is a field we want to make a very explicit message describing that fact
	if !isExported(tags.fieldName) {
		fmt.Fprintf(os.Stdout, "\tnot exported, will ignore\n")
		return
	}

	// If we reached here then the value cannot be set
	fmt.Fprintf(os.Stdout, "\tcan't set\n")
}

func (v *describeVisitor) visitBool(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	introDescription(value, tags, path)
}

func (v *describeVisitor) visitInt(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	introDescription(value, tags, path)

	if !value.CanSet() {
		// If we can't set this value don't provide any other details about it
		return
	}

	// First check if there is a list of valid string values
	if tags.intValues.wasSet {
		fmt.Fprintf(os.Stdout, "\tmethod (%s): %s\n", tags.intValues.methodName, methodValuesString(tags.intValues.value))
		return
	}

	fmt.Fprintf(os.Stdout, "\trange min: %d max: %d\n", tags.intRange.intMin, tags.intRange.intMax)

}

func (v *describeVisitor) visitUint(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	introDescription(value, tags, path)

	if !value.CanSet() {
		// If we can't set this value don't provide any other details about it
		return
	}

	// First check if there is a list of valid uint values
	if tags.uintValues.wasSet {
		fmt.Fprintf(os.Stdout, "\tmethod (%s): %s\n", tags.uintValues.methodName, methodValuesString(tags.uintValues.value))
		return
	}

	fmt.Fprintf(os.Stdout, "\trange min: %d max: %d\n", tags.uintRange.uintMin, tags.uintRange.uintMax)
}

func (v *describeVisitor) visitUintptr(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	notSupported(value, path)
}

func (v *describeVisitor) visitFloat(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	introDescription(value, tags, path)

	if !value.CanSet() {
		// If we can't set this value don't provide any other details about it
		return
	}

	// First check if there is a list of valid float values
	if tags.floatValues.wasSet {
		fmt.Fprintf(os.Stdout, "\tmethod (%s): %s\n", tags.floatValues.methodName, methodValuesString(tags.floatValues.value))
		return
	}

	fmt.Fprintf(os.Stdout, "\trange min: %g max: %g\n", tags.floatRange.floatMin, tags.floatRange.floatMax)
}

func (v *describeVisitor) visitComplex(value reflect.Value, tags fuzzTags, path valuePath) {
	// if this upsets you we can probably add it
	notSupported(value, path)
}

func (v *describeVisitor) visitArray(value reflect.Value, tags fuzzTags, path valuePath) {
	introDescription(value, tags, path)
}

func (v *describeVisitor) visitPointer(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	//introDescription(value, tags, path)

	if !value.CanSet() {
		return
	}

	// allocate a value for value to point to
	pType := value.Type()
	vType := pType.Elem()
	newVal := reflect.New(vType)
	value.Set(newVal)
}

func (v *describeVisitor) visitSlice(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) (from, to int) {
	if value.Len() != 0 {
		// This slice has already been set, we are building it
		// recursively with unbounded size. In Describe we stop after
		// the first iteration.
		return 0, 0
	}

	introDescription(value, tags, path)

	fmt.Fprintf(os.Stdout, "\trange min: %d max: %d\n", tags.sliceRange.uintRange.uintMin, tags.sliceRange.uintRange.uintMax)

	if !value.CanSet() {
		return 0, 0
	}

	newSlice := reflect.MakeSlice(value.Type(), 1, 1)
	value.Set(newSlice)

	return 0, 1
}

func (v *describeVisitor) visitMap(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) int {
	introDescription(value, tags, path)

	fmt.Fprintf(os.Stdout, "\trange min: %d max: %d\n", tags.mapRange.uintRange.uintMin, tags.mapRange.uintRange.uintMax)

	if !value.CanSet() {
		return 0
	}

	mapLen := 1

	mapType := value.Type()
	newMap := reflect.MakeMapWithSize(mapType, mapLen)
	value.Set(newMap)

	return mapLen
}

func (v *describeVisitor) visitChan(value reflect.Value, tags fuzzTags, path valuePath) {
	notSupported(value, path)
}

func (v *describeVisitor) visitFunc(value reflect.Value, tags fuzzTags, path valuePath) {
	notSupported(value, path)
}

func (v *describeVisitor) visitInterface(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) bool {
	notSupported(value, path)
	return false
}

func (v *describeVisitor) visitString(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	introDescription(value, tags, path)

	if !value.CanSet() {
		// If we can't set this value don't provide any other details about it
		return
	}

	// First check if there is a list of valid string values
	if tags.stringValues.wasSet {
		fmt.Fprintf(os.Stdout, "\tmethod (%s): %s\n", tags.stringValues.methodName, methodValuesString(tags.stringValues.value))
		return
	}

	fmt.Fprintf(os.Stdout, "\trange min: %d max: %d\n", tags.stringRange.uintRange.uintMin, tags.stringRange.uintRange.uintMax)
}

func (v *describeVisitor) visitStruct(value reflect.Value, tags fuzzTags, path valuePath) bool {
	recursion := path.containsType(value.Type())

	if !value.CanSet() || recursion {
		// We only describe a struct if we can't set it, or if it is recursive
		// If it can be set then it will be described via its fields
		introDescription(value, tags, path)
	}

	if recursion {
		fmt.Fprintf(os.Stdout, "\tRecursion...\n")
	}

	// If we've already visited (and described) a struct
	// we don't want to visit it again - so we return false
	return !recursion
}

func (v *describeVisitor) visitUnsafePointer(value reflect.Value, tags fuzzTags, path valuePath) {
	notSupported(value, path)
}

func notSupported(value reflect.Value, path valuePath) {
	fmt.Fprintf(os.Stdout, "%s\n", path.pathString(value))
	fmt.Fprintln(os.Stdout, "\tnot supported, will ignore")
}
