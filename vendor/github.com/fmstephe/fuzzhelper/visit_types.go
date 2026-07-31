package fuzzhelper

import (
	"fmt"
	"reflect"
)

type visitFunc func() []visitFunc

type valueVisitor interface {
	visitBool(reflect.Value, *byteConsumer, fuzzTags, valuePath)
	visitInt(reflect.Value, *byteConsumer, fuzzTags, valuePath)
	visitUint(reflect.Value, *byteConsumer, fuzzTags, valuePath)
	visitUintptr(reflect.Value, *byteConsumer, fuzzTags, valuePath)
	visitFloat(reflect.Value, *byteConsumer, fuzzTags, valuePath)
	visitComplex(reflect.Value, fuzzTags, valuePath)
	visitArray(reflect.Value, fuzzTags, valuePath)
	visitChan(reflect.Value, fuzzTags, valuePath)
	visitFunc(reflect.Value, fuzzTags, valuePath)
	visitInterface(reflect.Value, *byteConsumer, fuzzTags, valuePath) bool
	visitMap(reflect.Value, *byteConsumer, fuzzTags, valuePath) int
	visitPointer(reflect.Value, *byteConsumer, fuzzTags, valuePath)
	visitSlice(reflect.Value, *byteConsumer, fuzzTags, valuePath) (from, to int)
	visitString(reflect.Value, *byteConsumer, fuzzTags, valuePath)
	visitStruct(reflect.Value, fuzzTags, valuePath) bool
	visitUnsafePointer(reflect.Value, fuzzTags, valuePath)
}

func newVisitFunc(callback valueVisitor, value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) visitFunc {
	return func() []visitFunc {
		//println(fmt.Sprintf("before %#v\n", value.Interface()))
		ffs := visitValue(callback, value, c, tags, path)
		//println(fmt.Sprintf("after %#v\n", value.Interface()))
		return ffs
	}
}

func visitRoot(callback valueVisitor, root any, c *byteConsumer) {
	rootVal := reflect.ValueOf(root)

	path := valuePath{}
	visitBreadthFirst(callback, rootVal, c, path)

	//println("")
}

func visitBreadthFirst(callback valueVisitor, value reflect.Value, c *byteConsumer, path valuePath) {
	values := newDeque[visitFunc]()

	visitFuncs := visitValue(callback, value, c, newEmptyFuzzTags(), path)
	values.addMany(visitFuncs)

	for values.len() != 0 {
		ff := values.popFirst()
		visitFuncs := ff()
		values.addMany(visitFuncs)
	}
}

func visitValue(callback valueVisitor, value reflect.Value, c *byteConsumer, tags fuzzTags, path valuePath) []visitFunc {
	if c.len() == 0 {
		// There are no more bytes to use to visit data
		return []visitFunc{}
	}

	switch value.Kind() {
	case reflect.Bool:
		callback.visitBool(value, c, tags, path)
		return []visitFunc{}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		callback.visitInt(value, c, tags, path)
		return []visitFunc{}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		callback.visitUint(value, c, tags, path)
		return []visitFunc{}

	case reflect.Uintptr:
		callback.visitUintptr(value, c, tags, path)
		return []visitFunc{}

	case reflect.Float32, reflect.Float64:
		callback.visitFloat(value, c, tags, path)
		return []visitFunc{}

	case reflect.Complex64, reflect.Complex128:
		callback.visitComplex(value, tags, path)
		return []visitFunc{}

	case reflect.Array:
		callback.visitArray(value, tags, path)

		newValues := []visitFunc{}
		for i := 0; i < value.Len(); i++ {
			pathVal := fmt.Sprintf("[%d]", i)
			newValues = append(newValues, visitValue(callback, value.Index(i), c, tags, path.add(value, pathVal))...)
		}
		return newValues

	case reflect.Chan:
		callback.visitChan(value, tags, path)
		return []visitFunc{}

	case reflect.Func:
		callback.visitFunc(value, tags, path)
		return []visitFunc{}

	case reflect.Interface:
		if callback.visitInterface(value, c, tags, path) {
			return []visitFunc{
				newVisitFunc(callback, value.Elem(), c, newEmptyFuzzTags(), path.add(value, "(ifc)")),
			}
		} else {
			return []visitFunc{}
		}

	case reflect.Map:
		mapLen := callback.visitMap(value, c, tags, path)

		mapType := value.Type()
		keyType := mapType.Key()
		valType := mapType.Elem()

		newValues := []visitFunc{}
		for range mapLen {
			// Create the key
			mapKeyP := reflect.New(keyType)
			mapKey := mapKeyP.Elem()
			// Note here that the tags used to create this map are also
			// used to create the key
			newValues = append(newValues, visitValue(callback, mapKey, c, tags, path.add(value, "[key]"))...)

			// Create the value
			mapValP := reflect.New(valType)
			mapVal := mapValP.Elem()
			// Note here that the tags used to create this map are also
			// used to create the value
			newValues = append(newValues, visitValue(callback, mapVal, c, tags, path.add(value, "[value]"))...)

			// Add key/val to map
			//println("setting map element")
			value.SetMapIndex(mapKey, mapVal)
		}

		return newValues

	case reflect.Pointer:
		callback.visitPointer(value, c, tags, path)
		return []visitFunc{
			newVisitFunc(callback, value.Elem(), c, newEmptyFuzzTags(), path.add(value, "*")),
		}

	case reflect.Slice:
		from, to := callback.visitSlice(value, c, tags, path)

		// Add a single element to the slice (which should be non-nil now)
		newValues := []visitFunc{}

		if !value.CanSet() {
			return newValues
		}

		// Fill in all elements.
		for i := from; i < to; i++ {
			pathVal := fmt.Sprintf("[%d]", i)
			newValues = append(newValues, visitValue(callback, value.Index(i), c, tags, path.add(value, pathVal))...)
		}

		if !tags.sliceRange.uintRange.wasSet && from != to {
			// This slice has an unbounded size.  Create a
			// recursive callback to this slice, to allow more
			// elements to be appended to the slice if there is
			// enough data
			newValues = append(newValues, newVisitFunc(callback, value, c, tags, path))
		}

		return newValues

	case reflect.String:
		callback.visitString(value, c, tags, path)
		return []visitFunc{}

	case reflect.Struct:
		if !callback.visitStruct(value, tags, path) {
			// We allow the visitor to elect not to process a struct.
			// This was introduced to allow the Describe() function
			// to avoid infinite recursion
			return []visitFunc{}
		}

		if !value.CanSet() {
			// Can't set struct - ignore the struct and ignore its fields
			return []visitFunc{}
		}

		newValues := []visitFunc{}
		vType := value.Type()
		path = path.add(value, "("+vType.Name()+")")
		for i := 0; i < vType.NumField(); i++ {
			vField := value.Field(i)
			tField := vType.Field(i)
			tags := newFuzzTags(value, tField)
			newValues = append(newValues, visitValue(callback, vField, c, tags, path.add(value, tField.Name))...)
		}

		return newValues

	case reflect.UnsafePointer:
		callback.visitUnsafePointer(value, tags, path)
		return []visitFunc{}

	default:
		panic(fmt.Errorf("unsupported kind %s", value.Kind()))
	}

	panic("unreachable")
}
