package objectstore

import (
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
