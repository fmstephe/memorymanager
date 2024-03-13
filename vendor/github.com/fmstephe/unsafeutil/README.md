# Unsafe Conversions Between String <-> []byte

This repo is intended to be a clear implementation of the solution suggested by Keith Randal here

https://groups.google.com/d/msg/golang-nuts/Zsfk-VMd_fU/WXPjfZwPBAAJ

This avoids the problems associated with using a `uintptr` as well as failing tests if the runtime representation of string or []byte changes.
