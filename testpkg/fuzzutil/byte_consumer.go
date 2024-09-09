// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package fuzzutil

import (
	"encoding/binary"
)

type ByteConsumer struct {
	bytes []byte
}

func NewByteConsumer(bytes []byte) *ByteConsumer {
	return &ByteConsumer{
		bytes: bytes,
	}
}

func (c *ByteConsumer) Len() int {
	return len(c.bytes)
}

func (c *ByteConsumer) Bytes(size int) []byte {
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
func (c *ByteConsumer) pushBytes(bytes []byte) {
	c.bytes = append(c.bytes, bytes...)
}

func (c *ByteConsumer) Byte() byte {
	dest := c.Bytes(1)
	return dest[0]
}

// Test only
func (c *ByteConsumer) pushByte(b byte) {
	c.pushBytes([]byte{b})
}

func (c *ByteConsumer) Uint16() uint16 {
	dest := c.Bytes(2)
	return binary.LittleEndian.Uint16(dest)
}

// Test only
func (c *ByteConsumer) pushUint16(value uint16) {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, value)
	c.pushBytes(bytes)
}

func (c *ByteConsumer) Uint32() uint32 {
	dest := c.Bytes(4)
	return binary.LittleEndian.Uint32(dest)
}

// Test only
func (c *ByteConsumer) pushUint32(value uint32) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, value)
	c.pushBytes(bytes)
}
