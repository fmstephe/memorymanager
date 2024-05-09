package objectstore

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that when we allocate a string, the correct value is stored and
// retrieved.
func Test_String_AllocateAndGet_Simple(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Create string
	value := "test string"

	// Allocate it
	refString, valueOutString := AllocFromStr(ss, value)

	// Assert that we can get the correct string from the Reference
	assert.Equal(t, value, valueOutString)
	assert.Equal(t, value, refString.Value())

	// Allocate it
	refBytes, valueOutBytes := AllocFromBytes(ss, []byte(value))

	// Assert that we can get the correct string from the Reference
	assert.Equal(t, value, valueOutBytes)
	assert.Equal(t, value, refBytes.Value())
}

// Test that when we allocate a string, the correct value is stored and
// retrieved.  This is a more complex version of the test above, testing a wide
// range of string sizes
func Test_String_AllocateAndGet(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	for _, length := range []int{
		0,
		1,
		2,
		3,
		4,
		(1 << 5) - 1,
		1 << 5,
		(1 << 5) + 1,
		(1 << 9) - 1,
		1 << 9,
		(1 << 9) + 1,
		(1 << 14) - 1,
		1 << 14,
		(1 << 14) + 1,
	} {
		t.Run(fmt.Sprintf("Allocate and get %d", length), func(t *testing.T) {
			// Generate a string of the desired size
			value := makeSizedString(length)

			// Allocate it using the string
			refString, valueOutString := AllocFromStr(ss, value)

			// Assert that we can get the correct string from the Reference
			assert.Equal(t, value, valueOutString)
			assert.Equal(t, value, refString.Value())

			// Allocate it using bytes
			refBytes, valueOutBytes := AllocFromBytes(ss, []byte(value))

			// Assert that we can get the correct string from the Reference
			assert.Equal(t, value, valueOutBytes)
			assert.Equal(t, value, refBytes.Value())
		})
	}
}

// Demonstrate that we can create a string then free it. If we call Value()
// on the freed RefStr call will panic
func Test_String_NewFreeGet_Panic(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Allocate and free a string value
	value := "test string"
	ref, _ := AllocFromStr(ss, value)
	FreeStr(ss, ref)

	// Assert that calling Value() now panics
	assert.Panics(t, func() { ref.Value() })
}

// Demonstrate that we can create a string then free it twice. The second Free
// call will panic.
func Test_String_NewFreeFree_Panic(t *testing.T) {
	ss := NewSized(1 << 8)
	defer func() {
		ss.Destroy()
	}()

	// Allocate and free a string value
	value := "test string"
	ref, _ := AllocFromStr(ss, value)
	FreeStr(ss, ref)

	// Assert that calling FreeStr() now panics
	assert.Panics(t, func() { FreeStr(ss, ref) })
}

// Demonstrate that when we double free a re-allocated string we still panic.
func Test_String_NewFreeAllocFree_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	value := "test string"
	r, _ := AllocFromStr(os, value)
	FreeStr(os, r)
	// This will re-allocate the just-freed string
	AllocFromStr(os, value)

	assert.Panics(t, func() { FreeStr(os, r) })
}

// Demonstrate that when we call Value() on a re-allocated RefStr we still panic
func Test_String_NewFreeAllocGet_Panic(t *testing.T) {
	os := NewSized(1 << 8)
	defer os.Destroy()

	value := "test string"
	r, _ := AllocFromStr(os, value)
	FreeStr(os, r)
	// This will re-allocate the just-freed string
	AllocFromStr(os, value)

	assert.Panics(t, func() { r.Value() })
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var strRand = rand.New(rand.NewSource(1))

func makeSizedString(length int) string {
	builder := strings.Builder{}
	builder.Grow(length)
	for range length {
		builder.WriteByte(letters[strRand.Intn(len(letters))])
	}
	return builder.String()
}
