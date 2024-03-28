package objectstore

import (
	"fmt"
	"reflect"
)

func containsNoPointers[O any]() error {
	t := reflect.TypeFor[O]()
	return searchForPointers(t)
}

func searchForPointers(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Bool:
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return nil

	case reflect.Float32, reflect.Float64:
		return nil

	case reflect.Complex64, reflect.Complex128:
		return nil

	case reflect.Array:
		return searchForPointers(t.Elem())

	case reflect.Chan:
		return fmt.Errorf("channel found in %s", t)

	case reflect.Func:
		return fmt.Errorf("func found in %s", t)

	case reflect.Interface:
		return fmt.Errorf("interface found in %s", t)

	case reflect.Map:
		return fmt.Errorf("map found in %s", t)

	case reflect.Pointer:
		return fmt.Errorf("pointer found in %s", t)

	case reflect.Slice:
		return fmt.Errorf("slice found in %s", t)

	case reflect.String:
		return fmt.Errorf("string found in %s", t)

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			sV := t.Field(i)
			err := searchForPointers(sV.Type)
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.UnsafePointer:
		return fmt.Errorf("unsafe pointer found in %s", t)

	default:
		return fmt.Errorf("unknown kind found in %s", t)
	}
}
