package objectstore

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Demonstrate that the zero value of a reference is a nil reference
func TestIsNil(t *testing.T) {
	p := Reference[string]{}
	assert.True(t, p.IsNil())
}

// Demonstrate that a pointer with any non-0 field is not nil
func TestIsNotNil(t *testing.T) {
	// Show that a range of basic values are not nil
	for i := uint64(0); i < 10; i++ {
		r := newReference[string](i)
		assert.False(t, r.IsNil())
	}
	{
		// Show that some very large values are also not-nil
		r := Reference[string]{
			allocIdx: math.MaxUint64,
		}
		assert.False(t, r.IsNil())
	}

	{
		r := Reference[string]{
			allocIdx: math.MaxUint64 - 1,
		}
		assert.False(t, r.IsNil())
	}
}

func TestReferenceChunkAndOffset(t *testing.T) {
	// Choose a range of power-of-two chunk sizes
	for chunkSize := uint64(1); chunkSize < 1<<12; chunkSize = chunkSize << 1 {
		// Show that for a range of allocation locations the chunkIdx, offsetIdx are correct
		for i := uint64(0); i < chunkSize*4; i++ {
			r := newReference[string](i)
			chunkIdx, offsetIdx := r.chunkAndOffset(chunkSize)
			assert.Less(t, offsetIdx, chunkSize)
			assert.Equal(t, chunkIdx*chunkSize+offsetIdx, i)
		}
	}
}
