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

func (c *ByteConsumer) ConsumeBytes(size int) []byte {
	consumed := make([]byte, size)
	copy(consumed, c.bytes)

	if len(c.bytes) <= size {
		c.bytes = c.bytes[:0]
	} else {
		c.bytes = c.bytes[size:]
	}
	return consumed
}

func (c *ByteConsumer) ConsumeByte() byte {
	dest := c.ConsumeBytes(1)
	return dest[0]
}

func (c *ByteConsumer) ConsumeUint16() uint16 {
	dest := c.ConsumeBytes(2)
	return binary.LittleEndian.Uint16(dest)
}

func (c *ByteConsumer) ConsumeUint32() uint32 {
	dest := c.ConsumeBytes(4)
	return binary.LittleEndian.Uint32(dest)
}
