// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.
package pointerstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlabIntegrity(t *testing.T) {
	for _, objectSize := range []uint64{
		0,
		1,
		2,
		3,
		(1 << 3) - 1,
		1 << 3,
		(1 << 3) + 1,
		(1 << 4) - 1,
		1 << 4,
		(1 << 4) + 1,
		(1 << 5) - 1,
		1 << 5,
		(1 << 5) + 1,
		(1 << 6) - 1,
		1 << 6,
		(1 << 6) + 1,
		(1 << 7) - 1,
		1 << 7,
		(1 << 7) + 1,
		(1 << 8) - 1,
		1 << 8,
		(1 << 8) + 1,
		(1 << 9) - 1,
		1 << 9,
		(1 << 9) + 1,
		(1 << 10) - 1,
		1 << 10,
		(1 << 10) + 1,
		(1 << 11) - 1,
		1 << 11,
		(1 << 11) + 1,
		(1 << 12) - 1,
		1 << 12,
		(1 << 12) + 1,
		(1 << 13) - 1,
		1 << 13,
		(1 << 13) + 1,
		(1 << 14) - 1,
		1 << 14,
		(1 << 14) + 1,
		(1 << 15) - 1,
		1 << 15,
		(1 << 15) + 1,
	} {
		t.Run(fmt.Sprintf("Test allocation integrity for %d", objectSize), func(t *testing.T) {
			conf := NewAllocConfigBySize(objectSize, 1<<16)
			store := New(conf)
			defer store.Destroy()

			// Force 3 slabs to be created for this object size
			// Test that the allocations for each slab are correct
			for range 3 {
				refs := []RefPointer{}
				for range conf.ObjectsPerSlab {
					refs = append(refs, store.Alloc())
				}

				baseSlabData := refs[0].DataPtr()
				baseSlabMetadata := refs[0].metadataPtr()

				// Check that the metadata is allocated immediately _after_ the data
				assert.Equal(t, baseSlabMetadata, baseSlabData+uintptr(conf.TotalObjectSize))

				// Check all the allocations for their integrity
				for i, ref := range refs {
					// Check that the data allocations are spaced out appropriately
					dataPtr := ref.DataPtr()
					expectedDataOffset := uintptr(conf.ObjectSize) * uintptr(i)
					assert.Equal(t, baseSlabData+expectedDataOffset, dataPtr)

					// Check that the metadata allocations are spaced out appropriately
					metaPtr := ref.metadataPtr()
					expectedMetaOffset := uintptr(conf.MetadataSize) * uintptr(i)
					assert.Equal(t, baseSlabMetadata+expectedMetaOffset, metaPtr)
				}
			}
		})
	}
}
