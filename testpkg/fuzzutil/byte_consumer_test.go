// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package fuzzutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteConsumer_Bytes(t *testing.T) {
	consumer := NewByteConsumer([]byte{})
	consumer.pushBytes([]byte{1, 2, 3, 4, 5, 6})
	consumer.pushByte(7)
	assert.Equal(t, 7, consumer.Len())

	// Consume the available bytes
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6}, consumer.Bytes(6))
	assert.Equal(t, 1, consumer.Len())

	// Consume bytes, but not enoough available - get remaining bytes and zeroes
	assert.Equal(t, []byte{7, 0, 0, 0, 0, 0}, consumer.Bytes(6))
	assert.Equal(t, 0, consumer.Len())

	// Consume bytes, but none available - get zeroes
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, consumer.Bytes(6))
	assert.Equal(t, 0, consumer.Len())
}

func TestByteConsumer_Byte(t *testing.T) {
	consumer := NewByteConsumer([]byte{})
	consumer.pushByte(12)
	assert.Equal(t, 1, consumer.Len())

	// Consume the available bytes
	assert.Equal(t, byte(12), consumer.Byte())
	assert.Equal(t, 0, consumer.Len())

	// Consume bytes, but none available - get zeroes
	assert.Equal(t, byte(0), consumer.Byte())
	assert.Equal(t, 0, consumer.Len())
}

func TestByteConsumer_Uint16(t *testing.T) {
	consumer := NewByteConsumer([]byte{})
	consumer.pushUint16(10_000)
	consumer.pushByte(7)
	assert.Equal(t, 3, consumer.Len())

	// Consume the available bytes
	assert.Equal(t, uint16(10_000), consumer.Uint16())
	assert.Equal(t, 1, consumer.Len())

	// Consume bytes, but not enoough available - get remaining bytes and zeroes
	assert.Equal(t, uint16(7), consumer.Uint16())
	assert.Equal(t, 0, consumer.Len())

	// Consume bytes, but none available - get zeroes
	assert.Equal(t, uint16(0), consumer.Uint16())
	assert.Equal(t, 0, consumer.Len())
}

func TestByteConsumer_Uint32(t *testing.T) {
	consumer := NewByteConsumer([]byte{})
	consumer.pushUint32(100_000)
	consumer.pushByte(7)
	assert.Equal(t, 5, consumer.Len())

	// Consume the available bytes
	assert.Equal(t, uint32(100_000), consumer.Uint32())
	assert.Equal(t, 1, consumer.Len())

	// Consume bytes, but not enoough available - get remaining bytes and zeroes
	assert.Equal(t, uint32(7), consumer.Uint32())
	assert.Equal(t, 0, consumer.Len())

	// Consume bytes, but none available - get zeroes
	assert.Equal(t, uint32(0), consumer.Uint32())
	assert.Equal(t, 0, consumer.Len())
}

func TestByteConsumer_Combined(t *testing.T) {
	consumer := NewByteConsumer([]byte{})
	consumer.pushBytes([]byte{1, 2, 3, 4, 5, 6})
	consumer.pushByte(12)
	consumer.pushUint16(10_000)
	consumer.pushUint32(100_000)
	assert.Equal(t, 13, consumer.Len())

	// Consume the available bytes
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6}, consumer.Bytes(6))
	assert.Equal(t, 7, consumer.Len())

	assert.Equal(t, byte(12), consumer.Byte())
	assert.Equal(t, 6, consumer.Len())

	assert.Equal(t, uint16(10_000), consumer.Uint16())
	assert.Equal(t, 4, consumer.Len())

	assert.Equal(t, uint32(100_000), consumer.Uint32())
	assert.Equal(t, 0, consumer.Len())
}
