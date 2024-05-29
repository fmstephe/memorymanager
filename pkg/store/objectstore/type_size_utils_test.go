package objectstore

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxAllocationBits(t *testing.T) {
	assert.Equal(t, 31, maxAllocationBitsInternal(32))
	assert.Equal(t, 48, maxAllocationBitsInternal(64))
	assert.Panics(t, func() { maxAllocationBitsInternal(3) })
	assert.Panics(t, func() { maxAllocationBitsInternal(8) })
	assert.Panics(t, func() { maxAllocationBitsInternal(16) })
}

func TestMaxAllocationSize(t *testing.T) {
	assert.Equal(t, 1<<30, maxAllocationSizeInternal(32))
	assert.Equal(t, 1<<47, maxAllocationSizeInternal(64))
	assert.Panics(t, func() { maxAllocationSizeInternal(3) })
	assert.Panics(t, func() { maxAllocationSizeInternal(8) })
	assert.Panics(t, func() { maxAllocationSizeInternal(16) })
}

func TestResidentObjectSize_64BitArch(t *testing.T) {
	// Restore maxAllocSize after this test is complete
	oldMaxAllocSize := maxAllocSize
	defer func() {
		maxAllocSize = oldMaxAllocSize
	}()

	maxAllocSize = maxAllocationSizeInternal(64)

	// For all the power of two values which are small enough to allocate
	// the resident objects size is the same as the input
	for i := range 48 {
		requestedSize := 1 << i
		assert.Equal(t, requestedSize, residentObjectSize(requestedSize))
	}

	// For non-power of two values we round up to the nearest power of two
	// Requests for 0 sized allocations have resident size 1
	for i := range 48 {
		requestedSize := (1 << i) - 1
		switch requestedSize {
		case 0:
			assert.Equal(t, 1, residentObjectSize(requestedSize))
		case 1:
			assert.Equal(t, 1, residentObjectSize(requestedSize))
		default:
			assert.Equal(t, 1<<i, residentObjectSize(requestedSize))
		}
	}
	// For non-power of two values we round up to the nearest power of two
	// Requests for 0 sized allocations have resident size 1
	for i := range 47 {
		requestedSize := (1 << i) + 1
		assert.Equal(t, 1<<(i+1), residentObjectSize(requestedSize))
	}

	// For negative values residentObjectSize panics
	assert.Panics(t, func() { residentObjectSize(-1) })
	assert.Panics(t, func() { residentObjectSize(math.MinInt) })

	// For too large allocation values residentObjectSize panics
	for i := 49; i <= 63; i++ {
		assert.Panics(t, func() { residentObjectSize(1 << i) })
		assert.Panics(t, func() { residentObjectSize((1 << i) + 1) })
		assert.Panics(t, func() { residentObjectSize((1 << i) + 2) })
	}
}

func TestResidentObjectSize_32BitArch(t *testing.T) {
	// Restore maxAllocSize after this test is complete
	oldMaxAllocSize := maxAllocSize
	defer func() {
		maxAllocSize = oldMaxAllocSize
	}()

	maxAllocSize = maxAllocationSizeInternal(32)

	// For all the power of two values which are small enough to allocate
	// the resident objects size is the same as the input
	for i := range 31 {
		requestedSize := 1 << i
		assert.Equal(t, requestedSize, residentObjectSize(requestedSize))
	}

	// For non-power of two values we round up to the nearest power of two
	// Requests for 0 sized allocations have resident size 1
	for i := range 31 {
		requestedSize := (1 << i) - 1
		switch requestedSize {
		case 0:
			assert.Equal(t, 1, residentObjectSize(requestedSize))
		case 1:
			assert.Equal(t, 1, residentObjectSize(requestedSize))
		default:
			assert.Equal(t, 1<<i, residentObjectSize(requestedSize))
		}
	}
	// For non-power of two values we round up to the nearest power of two
	// Requests for 0 sized allocations have resident size 1
	for i := range 30 {
		requestedSize := (1 << i) + 1
		assert.Equal(t, 1<<(i+1), residentObjectSize(requestedSize))
	}

	// For negative values residentObjectSize panics
	assert.Panics(t, func() { residentObjectSize(-1) })
	assert.Panics(t, func() { residentObjectSize(math.MinInt32) })

	// For too large allocation values residentObjectSize panics
	assert.Panics(t, func() { residentObjectSize(1 << 31) })
	assert.Panics(t, func() { residentObjectSize((1 << 31) + 1) })
	assert.Panics(t, func() { residentObjectSize((1 << 31) + 2) })
}
