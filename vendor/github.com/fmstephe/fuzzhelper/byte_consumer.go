// Copyright 2025 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package fuzzhelper

import (
	"encoding/binary"
	"fmt"
	"math"
	"unicode/utf8"
	"unsafe"
)

var nativeInt = 0

const (
	bytesForNative = uintptr(unsafe.Sizeof(nativeInt))
	bytesFor64     = 8
	bytesFor32     = 4
	bytesFor16     = 2
	bytesFor8      = 1
)

type byteConsumer struct {
	bytes []byte
}

func newByteConsumer(bytes []byte) *byteConsumer {
	return &byteConsumer{
		bytes: bytes,
	}
}

func (c *byteConsumer) getRawBytes() []byte {
	return c.bytes
}

func (c *byteConsumer) len() int {
	return len(c.bytes)
}

func (c *byteConsumer) consume(size int) []byte {
	consumed := make([]byte, size)
	copy(consumed, c.bytes)

	if len(c.bytes) <= size {
		c.bytes = c.bytes[:0]
	} else {
		c.bytes = c.bytes[size:]
	}
	return consumed
}

// Test only
func (c *byteConsumer) pushBytes(bytes []byte) {
	c.bytes = append(c.bytes, bytes...)
}

func (c *byteConsumer) singleByte() byte {
	dest := c.consume(1)
	return dest[0]
}

// Test only
func (c *byteConsumer) pushByte(b byte) {
	c.pushBytes([]byte{b})
}

func (c *byteConsumer) consumeUint64(bytes uintptr) uint64 {
	switch bytes {
	case 8:
		dest := c.consume(8)
		return binary.LittleEndian.Uint64(dest)
	case 4:
		dest := c.consume(4)
		return uint64(binary.LittleEndian.Uint32(dest))
	case 2:
		dest := c.consume(2)
		return uint64(binary.LittleEndian.Uint16(dest))
	case 1:
		dest := c.consume(1)
		return uint64(dest[0])
	default:
		panic(fmt.Sprintf("Must provided either 8, 4, 2, or 1 as bytes argument. %d found.", bytes))
	}
}

func (c *byteConsumer) consumeInt64(bytes uintptr) int64 {
	switch bytes {
	case 8:
		dest := c.consume(8)
		return int64(binary.LittleEndian.Uint64(dest))
	case 4:
		dest := c.consume(4)
		return int64(int32((binary.LittleEndian.Uint32(dest))))
	case 2:
		dest := c.consume(2)
		return int64(int16((binary.LittleEndian.Uint16(dest))))
	case 1:
		dest := c.consume(1)
		return int64(int8((dest[0])))
	default:
		panic(fmt.Sprintf("Must provided either 8, 4, 2, or 1 as bytes argument. %d found.", bytes))
	}
}

func (c *byteConsumer) consumeFloat64(bytes uintptr) float64 {
	switch bytes {
	case 8:
		dest := c.consume(8)
		return math.Float64frombits(binary.LittleEndian.Uint64(dest))
	case 4:
		dest := c.consume(4)
		return float64(math.Float32frombits(binary.LittleEndian.Uint32(dest)))
	default:
		panic(fmt.Sprintf("Must provided either 8 or 4 as bytes argument. %d found.", bytes))
	}
}

// Returns a valid UTF8 string, which is at _most_ as long as length in runes
func (c *byteConsumer) String(length int) string {
	bytes := c.consume(length)

	// Extract the valid runes from the bytes
	//
	// The way this is implemented right now will create strings which are
	// randomly shorter than asked for. We may want to correct this.
	validRunes := []rune{}
	for len(bytes) > 0 {
		r, l := utf8.DecodeRune(bytes)
		bytes = bytes[l:]
		if r != utf8.RuneError {
			validRunes = append(validRunes, r)
		}
	}

	return string(validRunes)
}

func (c *byteConsumer) consumeBool() bool {
	bytes := c.consume(1)
	return bytes[0]%2 == 1
}

// test only
func (c *byteConsumer) pushUint64(value uint64, bytes uintptr) {
	switch bytes {
	case 8:
		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, value)
		c.pushBytes(bytes)
	case 4:
		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, uint32(value))
		c.pushBytes(bytes)
	case 2:
		bytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(bytes, uint16(value))
		c.pushBytes(bytes)
	case 1:
		bytes := make([]byte, 1)
		bytes[0] = byte(value)
		c.pushBytes(bytes)
	default:
		panic(fmt.Sprintf("Must provided either 8, 4, 2, or 1 as bytes argument. %d found.", bytes))
	}
}

// test only
func (c *byteConsumer) pushInt64(value int64, bytes uintptr) {
	switch bytes {
	case 8:
		c.pushUint64(uint64(value), bytes)
	case 4:
		c.pushUint64(uint64(int32(value)), bytes)
	case 2:
		c.pushUint64(uint64(int16(value)), bytes)
	case 1:
		c.pushUint64(uint64(int8(value)), bytes)
	default:
		panic(fmt.Sprintf("Must provided either 8, 4, 2, or 1 as bytes argument. %d found.", bytes))
	}
}

func (c *byteConsumer) pushFloat64(value float64, bytes uintptr) {
	switch bytes {
	case 8:
		floatBits := math.Float64bits(value)
		c.pushUint64(floatBits, 8)
	case 4:
		floatBits := uint64(math.Float32bits(float32(value)))
		c.pushUint64(floatBits, 4)
	default:
		panic(fmt.Sprintf("Must provided either 8 or 4 as bytes argument. %d found.", bytes))
	}
}

// test only
func (c *byteConsumer) pushString(str string) {
	c.pushInt64(int64(len(str)), bytesForNative)
	c.pushBytes([]byte(str))
}

func (c *byteConsumer) pushBool(value bool) {
	if value {
		c.pushBytes([]byte{1})
	} else {
		c.pushBytes([]byte{0})
	}
}
