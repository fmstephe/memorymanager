// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

// # Usage
//
// The offheap package allows us to allocate, free and retrieve Go objects
// manually. The objects allocated in this way are not visible to the garbage
// collector, allowing us to build very large datastructures like trees and
// hashmaps without impacting garbage collection efficiency.
//
// Each Store instance can allocate/free a wide variety of object types.
// Broadly we can allocate _any_ Go type so long as no part of the type
// contains pointers. We can allocate slices of pointerless types with a
// dedicated slice type. And finally we can allocate strings through a
// dedicated string type.
//
// Each allocation has a corresponding Reference, named RefObject, RefSlice and
// RefString respectively, which acts like a conventional pointer to retrieve
// the allocated object via Reference.Value() e.g.
//
//	var store *offheap.Store = offheap.New()
//
//	var ref offheap.RefObject[int] = offheap.AllocObject[int](store)
//	var i1 *int = ref.Value()
//
// When you know that an allocation will never be used again it's memory
// can be released back to the Store using one of the Free*() functions e.g.
//
//	var store *offheap.Store = offheap.New()
//
//	var ref offheap.RefObject[int] = offheap.AllocObject[int](store)
//
//	offheap.FreeObject(store, ref)
//	// You must never use ref again
//
// A best effort has been made to panic if an object is freed twice or if a
// freed object is accessed using Reference.Value(). However, it isn't
// guaranteed that these calls will panic.
//
// References can be kept and stored in arbitrary datastructures, which can
// themselves be managed by a Store e.g.
//
//	type Node struct {
//		left  offheap.RefObject[Node]
//		right offheap.RefObject[Node]
//	}
//
//	var store *offheap.Store = offheap.New()
//
//	var refParent offheap.RefObject[Node] = offheap.AllocObject[Node]((store))
//
// The Reference types contain no conventional Go pointers which are recognised
// by the garbage collector.
//
// It is important to note that the objects managed by a Store do not exist on
// the managed Go heap. They live in a series of manually mapped memory regions
// which are managed separately by the Store. This means that the amount of
// memory used by the Store has no impact on the frequency of garbage
// collection runs.
//
// Not all pointers in Go types are obvious. Here are a few examples of types
// which can't be managed in a Store.
//
//	type BadStruct1 struct {
//	  stringsHavePointers string
//	}
//
//	type BadStruct2 struct {
//	  mapsHavePointers map[int]int
//	}
//
//	type BadStruct3 struct {
//	  slicesHavePointers []int
//	}
//
//	type BadStruct4 struct {
//	  pointersHavePointers *int
//	}
//
//	type BadStruct5 struct {
//	  storesHavePointers *Store
//	}
//
// Trying to allocate an object or slice with a generic type which contains
// pointers will panic.
//
// Memory Model Constraints:
//
// A Store has a moderate degree of concurrency safety, but users must still be
// careful how they access and modify data allocated by a Store instance. A
// shorter version of the guarantees described below would be to say that
// allocating and retrieving objects managed by a Store has the same guarantees
// and limitations that conventionally allocated Go objects have.
//
// # Concurrency Guarantees
//
// 1: Independent Alloc/Free Safety
//
// It is safe for multiple goroutines using a shared Store instance to call
// Alloc() and Free() generating independent sets of objects/References. They
// can safely read the objects they have allocated without any additional
// concurrency protection.
//
// 2: Safe Data Publication
//
// It is safe to create objects using Alloc() and then share those objects with
// other goroutines. We must establish the usual happens-before relationships
// when sharing objects/References with other goroutines.
//
// For example it is safe to Alloc() new objects and publish References to
// those objects on a channel and have other goroutines read from that channel
// and call Reference.Value() on those References.
//
// 3: Independent Read Safety
//
// For a given set of live objects, previously allocated with a happens-before
// barrier between the allocator and readers, all objects can be read freely.
// Calling Reference.Value() and performing arbitrary reads of the retrieved
// objects from multiple goroutines with no other concurrency control code will
// work without data races.
//
// 4: Safe Object Reads And Writes
//
// It is not safe for multiple goroutines to freely write to the object, nor to
// have multiple goroutines freely perform a mixture of read/write operations
// on the object. You can however perform concurrent reads and writes to a
// shared object if you use appropriate concurrency controls such as
// sync.Mutex.
//
// 5: Free Safety
//
// If we call Free() on the same Reference (a Reference pointing to the same
// allocation) concurrently from two or more goroutines this will be a data
// race. The behaviour is unpredictable in this case. This is also a bug, but
// potentially one with stranger behaviour than just calling Free() twice from
// a single goroutine.
//
// If we call Free() on a Reference while another goroutine is calling
// Reference.Value() this is a data race. This will have unpredictable
// behaviour, and is never safe.
package offheap
