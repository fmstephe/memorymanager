# Low GC Quadtree

This project is a small experiment in building a quadtree datastructure with very low GC impact.

This project contains a standard pointer based quadtreee implementation to allow for easy comparisons between the two approaches.

The main novel component used here is the 'store' package which allows for the creation of objects from large contiguous slices of these objects. A pointer-like type is provided to store references to objects. It's worth noting that the creation of this ObjectStore only became easy and convenient with the introduction of generics. I think would have avoided this kind of project if I needed to copy/paste code for multiple types - or use code generation.

There is a significant limitation in the code right now, that there is no way to delete a created object, objects are created and then live forever. It's likely not a difficult to add a mechanism to allow for object deletion and reuse. This will be added in the future (probably).

There is also a ByteStore, which works very similarly to the ObjectStore but is custom built to allocate variably sized byte slices.
