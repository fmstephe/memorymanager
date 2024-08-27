// The intern package allows users to intern string values. These strings can
// be converted from either a []byte, int64 or Float64 value. The intention of
// this package is to allow the reuse of common string values and avoid
// allocating strings unnecessarily.
//
// The string value returned may be either a newly allocated string, or a
// previously allocated interned string from the cache. Intered strings are
// stored in an offheap manually managed store. This means that there is no
// garbage collection cost associated with keeping interned strings.
//
// It is expected that shorter strings are typically better targets for
// interning than longer strings (although the degree to which this is true
// will depend on the data being processed). To reflect this intuition we allow
// a per-string length limit to be set for a StringInterner instance. Strings
// which are longer than this value will not be interned.
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
// Right now we have a very inflexible system for interning []byte, int64 and
// float64. It seems to me that there will likely be a desire to provide
// interning of string conversions for other types, time.Time comes to mind
// immediately. It seems unlikely that we can or should attempt to anticipate
// every internable conversion point. We must find a way to allow user
// implemented string conversions with interning. Probably not easy, but
// potentially valuable.
package intern
