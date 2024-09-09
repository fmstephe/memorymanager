package intern

import "sync/atomic"

type internController struct {
	// Immutable fields
	maxLen   int   // set to <= 0 for unlimited string length
	maxBytes int64 // set to <= 0 for unlimited bytes
	// Mutable fields
	usedBytes atomic.Int64
}

func newController(maxLen, maxBytes int) *internController {
	return &internController{
		maxLen:    maxLen,
		maxBytes:  int64(maxBytes),
		usedBytes: atomic.Int64{},
	}
}

func (c *internController) getUsedBytes() int {
	return int(c.usedBytes.Load())
}

func (c *internController) canInternMaxLen(str string) bool {
	if c.maxLen <= 0 {
		// No length limit
		return true
	}

	// There is a limit to how long an interned string can be
	return len(str) <= c.maxLen
}

func (c *internController) canInternUsedBytes(str string) bool {
	for {
		usedBytes := c.usedBytes.Load()
		nextUsedBytes := usedBytes + int64(len(str))

		if (c.maxBytes > 0) && (nextUsedBytes > c.maxBytes) {
			// There is a limit to the total number of bytes that
			// can be interned and interning str would cause us to
			// exceed that limit
			return false
		}

		// Cas the new value into usedBytes
		//
		// if this fails then someone else has probably
		// interned a string - check usedBytes again
		if c.usedBytes.CompareAndSwap(usedBytes, nextUsedBytes) {
			break
		}
	}

	// str can be interned, and the additional bytes have been accounted for
	return true
}
