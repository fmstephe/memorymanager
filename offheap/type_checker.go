// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

type typeChecker struct {
	typeCacheLock sync.RWMutex
	typeCache     map[reflect.Type]struct{}
}

func newTypeChecker() typeChecker {
	return typeChecker{
		typeCacheLock: sync.RWMutex{},
		typeCache:     map[reflect.Type]struct{}{},
	}
}

func (c *typeChecker) checkType(t reflect.Type) {
	if c.lookupTypeCache(t) {
		return
	}

	if err := containsNoPointers(t); err != nil {
		panic(fmt.Errorf("cannot allocate generic type containing pointers %w", err))
	}

	c.writeTypeToCache(t)
}

func (c *typeChecker) lookupTypeCache(t reflect.Type) bool {
	c.typeCacheLock.RLock()
	defer c.typeCacheLock.RUnlock()

	_, ok := c.typeCache[t]
	return ok
}

func (c *typeChecker) writeTypeToCache(t reflect.Type) {
	c.typeCacheLock.Lock()
	defer c.typeCacheLock.Unlock()

	c.typeCache[t] = struct{}{}
}

type typePaths struct {
	paths []string
}

func (p *typePaths) addPath(path string) {
	p.paths = append(p.paths, path)
}

func (p *typePaths) Len() int {
	return len(p.paths)
}

func (p *typePaths) String() string {
	if p.Len() == 0 {
		return ""
	}

	result := ""
	for _, path := range p.paths {
		result += path + ","
	}
	// Quietly strip off the trailing ,
	return result[:len(result)-1]
}

func containsNoPointers(t reflect.Type) error {
	paths := &typePaths{}
	searchForPointers(t, "", paths)
	if paths.Len() != 0 {
		return fmt.Errorf("found pointer(s): %s", paths)
	}
	return nil
}

func searchForPointers(t reflect.Type, path string, paths *typePaths) {
	switch t.Kind() {
	case reflect.Bool:

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:

	case reflect.Float32, reflect.Float64:

	case reflect.Complex64, reflect.Complex128:

	case reflect.Array:
		size := strconv.Itoa(t.Len())
		searchForPointers(t.Elem(), path+"["+size+"]", paths)

	case reflect.Chan:
		paths.addPath(path + "<" + t.String() + ">")

	case reflect.Func:
		paths.addPath(path + "<" + t.String() + ">")

	case reflect.Interface:
		paths.addPath(path + "<" + t.String() + ">")

	case reflect.Map:
		paths.addPath(path + "<" + t.String() + ">")

	case reflect.Pointer:
		paths.addPath(path + "<" + t.String() + ">")

	case reflect.Slice:
		paths.addPath(path + "<" + t.String() + ">")

	case reflect.String:
		paths.addPath(path + "<" + t.String() + ">")

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			sV := t.Field(i)
			searchForPointers(sV.Type, path+"("+t.String()+")"+sV.Name, paths)
		}

	case reflect.UnsafePointer:
		paths.addPath(path + "<" + t.String() + ">")

	default:
		paths.addPath(path + "<" + t.String() + ">")
	}
}
