package intern

import "sync/atomic"

type internController struct {
	// Immutable fields
	maxLen   int64 // set of -1 for unlimited string length
	maxBytes int64 // set to -1 for unlimited bytes
	// Mutable fields
	usedBytes atomic.Int64
}

func newController(maxLen, maxBytes int) *internController {
	return &internController{
		maxLen:    int64(maxLen),
		maxBytes:  int64(maxBytes),
		usedBytes: atomic.Int64{},
	}
}

func (c *internController) canInternMaxLen(str string) bool {
	if c.maxLen != -1 {
		// There is a limit to how long an interned string can be

		if c.maxLen < int64(len(str)) {
			// str is too long, cannot intern str
			return false
		}
	}

	return true
}

func (c *internController) canInternUsedBytes(str string) bool {
	if c.maxBytes != -1 {
		// There is a limit to the total number of bytes that can be
		// interned

		for {
			usedBytes := c.usedBytes.Load()
			nextUsedBytes := usedBytes + int64(len(str))

			if nextUsedBytes > c.maxBytes {
				// Interning str would cause us to exceed
				// usedBytes limit, cannot intern str
				return false
			}

			// Cas the new value into usedBytes
			//
			// if this fails then someone else has probably
			// interned a string - better check usedBytes again
			if c.usedBytes.CompareAndSwap(usedBytes, nextUsedBytes) {
				break
			}

		}
	}

	// str can be interned, and the additional bytes have been accounted for
	return true
}
