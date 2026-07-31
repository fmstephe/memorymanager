package fuzzhelper

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type valuePath struct {
	values []reflect.Value
	names  []string
}

func (p valuePath) add(value reflect.Value, name string) valuePath {
	return valuePath{
		values: append(p.values, value),
		names:  append(p.names, name),
	}
}

func (p valuePath) containsType(typ reflect.Type) bool {
	for _, val := range p.values {
		if val.Type() == typ {
			return true
		}
	}
	return false
}

var pointerRegex = regexp.MustCompile(`\.(\**)\(`)

func (p valuePath) pathString(value reflect.Value) string {
	// If our current value is preceeded by any number of pointers, we
	// include those pointers into the value's type description for
	// readibility
	names := p.names
	for i := len(p.values) - 1; i >= 0; i-- {
		if p.values[i].Kind() == reflect.Pointer {
			value = p.values[i]
			names = names[:i]
		} else {
			break
		}
	}

	pStr := strings.Join(names, ".")
	pStr = strings.ReplaceAll(pStr, "*.", "*")
	pStr = strings.ReplaceAll(pStr, ".[", "[")
	pStr = strings.ReplaceAll(pStr, ".(", "(")
	// This complex regex replace pulls leading pointer '*' inside the (typeName) parenthesis
	// This all feels a bit convoluted - will think about it a bit more
	pStr = pointerRegex.ReplaceAllString(pStr, "($1")

	if len(pStr) > 0 {
		pStr += " "
	}

	pStr = pStr + "(" + typeString(value.Type()) + ")"

	return pStr
}

func typeString(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Pointer:
		return "*" + typeString(typ.Elem())
	case reflect.Slice:
		return "[]" + typeString(typ.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", typ.Len(), typeString(typ.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", typeString(typ.Key()), typeString(typ.Elem()))
	case reflect.Func:
		// We don't bother to capture the actual function signature
		return "func"
	case reflect.Chan:
		return "chan " + typeString(typ.Elem())
	case reflect.UnsafePointer:
		return "unsafe.Pointer"
	case reflect.Interface:
		// We don't bother to capture the actual interface type
		return "interface"
	default:
		if name := typ.Name(); name != "" {
			return name
		}
		return typ.String()
	}
}
