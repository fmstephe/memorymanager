# Memory Manager

This project contains a number of packages built to allow for manual memory management in Go.

This project was motivated by my experiences of the very high garbage collection costs associated with building in memory caches in Go. These pools of relatively long lived memory allocations can often cause very long garbage collection runs with associated high CPU usage. I envision that the offheap package can be used to build large long lived, and relatively stable, datastructures such as in-memory caches. It's definitely not expected to replace or even reach parity with conventionally managed Go allocations, but to be applied in specialised situations or in specialised services.

The primary package is the [offheap](offheap/docs.go) package sitting at the root of the project. This project allows us to allocate and free _reasonably_ normal Go data types. Convention Go pointers cannot be stored in these offheap allocations, but a number of pointer-like Reference types can be used for this purpose and clearly identify memory which must be managed manually.

The pkg/ directory contains utilities packages built using the offheap package. Most interestingly (to me) is the [intern](pkg/intern/docs.go) package which allows for the interning of very large numbers of strings with near zero garbage collection impact.

(Also, unrelated to such serious minded things as garbage collection or CPU usage, this project has been so much fun)
