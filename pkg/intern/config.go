package intern

import (
	"math/bits"
	"runtime"

	"github.com/fmstephe/offheap/offheap"
)

type Config struct {
	// Defines the maximum length of a string which can be interned.
	// Strings longer than this will be generated but not interned.
	//
	// <= 0 indicates no limit on string length.
	MaxLen int

	// Defines the maximum total number of bytes which can be interned.
	// Once this limit is reached no new strings will be interned.
	//
	// <= 0 indicates no limit on total bytes, this may risk memory
	// exhaustion.
	MaxBytes int

	// Defines the number shards used internally to determine the level of
	// available concurrency for the interner.
	//
	// Important to note that this is not an exactly configurable
	// parameter. The number of shards must always be a power of two, and
	// the value provided here may be rounded up if necessary.
	//
	// <= 0 indicates that the interner should determine the number of
	// shards automatically.
	Shards int

	// Defines the offheap store to use for allocating interned strings.
	//
	// If nil then a new store will be created internally. Only needed if
	// you want to share a single offheap store across multiple interners.
	Store *offheap.Store
}

func (c *Config) getMaxLen() int {
	return c.MaxLen
}

func (c *Config) getMaxBytes() int {
	return c.MaxBytes
}

func (c *Config) getShards() int {
	if c.Shards <= 0 {
		c.Shards = runtime.NumCPU()
	}

	return nextPowerOfTwo(c.Shards)
}

func (c *Config) getStore() *offheap.Store {
	if c.Store == nil {
		c.Store = offheap.New()
	}
	return c.Store
}

// Returns the smallest power of two >= val
func nextPowerOfTwo(val int) int {
	if val <= 1 {
		return 1
	}
	// Test if val is a power of two
	if val > 0 && val&(val-1) == 0 {
		return val
	}
	return 1 << bits.Len64(uint64(val))
}
