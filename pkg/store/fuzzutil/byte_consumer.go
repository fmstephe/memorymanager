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

func (c *ByteConsumer) Byte() byte {
	dest := c.Bytes(1)
	return dest[0]
}

func (c *ByteConsumer) Uint16() uint16 {
	dest := c.Bytes(2)
	return binary.LittleEndian.Uint16(dest)
}

func (c *ByteConsumer) Uint32() uint32 {
	dest := c.Bytes(4)
	return binary.LittleEndian.Uint32(dest)
}
