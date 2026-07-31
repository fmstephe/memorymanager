package fuzzhelper

import (
	"fmt"
	"reflect"
	"strings"
)

var _ valueVisitor = &fillVisitor{}

type fillVisitor struct {
}

func Fill(root any, bytes []byte) {
	visitRoot(&fillVisitor{}, root, newByteConsumer(bytes))
}

func (v *fillVisitor) visitBool(value reflect.Value, c *byteConsumer, _ fuzzTags, path valuePath) {
	//print(leftPad(len(path.names)))
	//print("bool")
	if !value.CanSet() {
		return
	}

	val := c.consumeBool()
	value.SetBool(val)
}

func (v *fillVisitor) visitInt(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	//print(leftPad(len(path.names)))
	//print("int")
	if !value.CanSet() {
		return
	}

	// First check there is a list of valid int values
	if tags.intValues.wasSet {
		val := c.consumeUint64(bytesForNative)
		intVal := tags.intValues.value[val%uint64(len(tags.intValues.value))]

		value.SetInt(intVal)
		return
	}

	val := c.consumeInt64(value.Type().Size())
	fittedVal := tags.intRange.fit(val)
	value.SetInt(fittedVal)
}

func (v *fillVisitor) visitUint(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	//print(leftPad(len(path.names)))
	//print("uint")
	if !value.CanSet() {
		return
	}

	// First check there is a list of valid uint values
	if tags.uintValues.wasSet {
		val := c.consumeUint64(bytesForNative)
		uintVal := tags.uintValues.value[val%uint64(len(tags.uintValues.value))]

		value.SetUint(uintVal)
		return
	}

	val := c.consumeUint64(value.Type().Size())
	fittedVal := tags.uintRange.fit(val)
	value.SetUint(fittedVal)
}

func (v *fillVisitor) visitUintptr(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	//print(leftPad(len(path.names)))
	//println("uintptr: ignored")
}

func (v *fillVisitor) visitFloat(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	//print(leftPad(len(path.names)))
	//print("float")
	if !value.CanSet() {
		return
	}

	// First check there is a list of valid uint values
	if tags.floatValues.wasSet {
		val := c.consumeUint64(bytesForNative)
		floatVal := tags.floatValues.value[val%uint64(len(tags.floatValues.value))]
		value.SetFloat(floatVal)
		return
	}

	val := c.consumeFloat64(value.Type().Size())
	fittedVal := tags.floatRange.fit(val)
	value.SetFloat(fittedVal)
}

func (v *fillVisitor) visitComplex(value reflect.Value, tags fuzzTags, path valuePath) {
	// Do nothing - complex numbers are simply not supported
	// we still visit them so we can _describe_ that we don't support them
	// We can add it, I just don't use them so I didn't bother at first
}

func (v *fillVisitor) visitArray(value reflect.Value, tags fuzzTags, path valuePath) {
	// Do nothing - the array is fixed in size and we do nothing here
	// Each of it's elements will be visited and we will fill those
}

func (v *fillVisitor) visitPointer(value reflect.Value, c *byteConsumer, _ fuzzTags, path valuePath) {
	//print(leftPad(len(path.names)))
	//print("pointer")
	if !value.CanSet() {
		return
	}

	// If the value is nil - allocate a value for it to point to
	pType := value.Type()
	vType := pType.Elem()
	newVal := reflect.New(vType)
	value.Set(newVal)
}

func (v *fillVisitor) visitSlice(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) (from, to int) {
	//print(leftPad(len(path.names)))
	if !value.CanSet() {
		return 0, 0
	}

	initialLen := value.Len()

	appendSize := 1
	if tags.sliceRange.uintRange.wasSet {
		// If we have a size range, use that to determine appendSize
		val := int(c.consumeInt64(bytesForNative))
		appendSize = tags.sliceRange.fit(val)
	}

	toAppend := reflect.MakeSlice(value.Type(), appendSize, appendSize)
	value.Set(reflect.AppendSlice(value, toAppend))

	//print("slice ", sliceLen)
	return initialLen, value.Len()
}

func (v *fillVisitor) visitMap(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) int {
	//print(leftPad(len(path.names)))
	if !value.CanSet() {
		return 0
	}

	val := int(c.consumeInt64(bytesForNative))
	mapLen := tags.mapRange.fit(val)

	//print("map ", mapLen)
	mapType := value.Type()
	newMap := reflect.MakeMapWithSize(mapType, mapLen)
	value.Set(newMap)

	return mapLen
}

func (v *fillVisitor) visitChan(value reflect.Value, tags fuzzTags, path valuePath) {
	// Do nothing - channels are simply not supported
	// we still visit them so we can _describe_ that we don't support them
}

func (v *fillVisitor) visitFunc(value reflect.Value, tags fuzzTags, path valuePath) {
	// Do nothing - functions are simply not supported
	// we still visit them so we can _describe_ that we don't support them
}

func (v *fillVisitor) visitInterface(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) bool {
	if !value.CanSet() {
		return false
	}

	if tags.interfaceValues.wasSet {
		// Choose one of the values from the interfaceValues slice
		val := c.consumeUint64(bytesForNative)
		options := tags.interfaceValues.value
		chosen := options[val%uint64(len(options))]

		// Use the chosen value (an actual object we use only as an
		// example) and extract its type so we can build a new instance
		// of it
		ptrType := reflect.ValueOf(chosen).Type()

		// NB: This check should be redundant as we validate the values
		// when processing the tags, but we leave it here to sleep
		// better.
		if ptrType.Kind() != reflect.Pointer {
			panic(fmt.Errorf("interface values (at %s) can only be satisfied by pointer types, found %s", path.pathString(reflect.New(ptrType)), ptrType))
		}

		// Construct new instance of value type
		newValue := reflect.New(ptrType.Elem())

		if !ptrType.AssignableTo(value.Type()) {
			panic(fmt.Errorf("interface value at %s cannot be satisfied with type %s", path.pathString(newValue), ptrType))
		}

		value.Set(newValue)
		return true
	}

	return false
}

func (v *fillVisitor) visitString(value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) {
	if !value.CanSet() {
		return
	}

	// First check if there is a list of valid string values
	if tags.stringValues.wasSet {
		val := c.consumeUint64(bytesForNative)
		str := tags.stringValues.value[val%uint64(len(tags.stringValues.value))]

		value.SetString(str)
		return
	}

	lengthVal := int(c.consumeInt64(bytesForNative))
	strLength := tags.stringRange.fit(lengthVal)

	val := c.String(strLength)
	value.SetString(val)
}

func (v *fillVisitor) visitStruct(value reflect.Value, tags fuzzTags, path valuePath) bool {
	// Do nothing - the struct is fixed in size and we do nothing here
	// Each of it's fields will be visited and we will fill those
	return true
}

func (v *fillVisitor) visitUnsafePointer(value reflect.Value, tags fuzzTags, path valuePath) {
	// Do nothing - unsafe pointers are simply not supported
	// we still visit them so we can _describe_ that we don't support them
}

//lint:ignore U1000 This method is used if we uncomment the debug printlns
func leftPad(l int) string {
	return strings.Repeat(" ", l)
}
