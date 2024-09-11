// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

// # Usage
//
// The intern package allows users to intern string values.
//
// Importantly the interners built using this package can take types which are
// _not_ strings but which can be used to generate strings. This has the
// advantage that when the string representation of an object has already been
// interned we can skip generating the string and just return the interned
// string.
//
// An example where this would be advantageous would be in a system which
// converts a lot of integers to strings. If a lot of those integer values are
// common values then this package would avoid a lot of those string
// allocations.
//
// The basic interface of an interner is a single method which looks like
//
//	someTypeInterner.Get(someTypeValue) string
//
// The string value returned may be either a newly allocated string, or a
// previously allocated interned string from the cache. Interned strings are
// stored in an *offheap.Store. This means that there is no garbage collection
// cost associated with keeping large numbers of interned strings.
//
// This package contains a number of pre-made interners for the types int64,
// float64, time.Time, []byte and string. But this package also includes the
// tools to build custom interners for other types.
//
// Because the interned strings are manually managed, and we don't have a
// mechanism for knowing when to free interned string values, interned strings
// are retained for the life of the StringInterner instance. This means that we
// accumulate interned strings as the StringInterner is used. To prevent
// uncontrolled memory exhaustion we configure an upper limit on the total
// number of bytes which can be used to intern strings. When this limit is
// reached no new strings will be interned.
//
// It is expected that strings which are a good target for interning should
// appear for interning frequently and there should be a finite number of these
// common string values. In the case where this pattern holds true a well
// configured StringIntern cache will intern these popular strings before the
// byte limit is reached. If strings to be interned evolve over time and don't
// have a stable set of common string values, then this interning approach will
// be less effective.
//
// It should be reasonably easy to create new interners using the types found
// in the internbase package. Just following the implementation of the
// interners found in this package.
package intern
