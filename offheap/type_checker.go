// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package offheap

import (
	"fmt"
	"reflect"
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
