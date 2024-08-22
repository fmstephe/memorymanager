package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomStringMaker_MakeSizedString(t *testing.T) {
	rsm := NewRandomStringMaker()

	for i := 0; i < 1000; i++ {
		str := rsm.MakeSizedString(i)
		assert.Equal(t, i, len(str))
	}
}

func TestRandomStringMaker_MakeSizedBytes(t *testing.T) {
	rsm := NewRandomStringMaker()

	for i := 0; i < 1000; i++ {
		bytes := rsm.MakeSizedBytes(i)
		assert.Equal(t, i, len(bytes))
	}
}
