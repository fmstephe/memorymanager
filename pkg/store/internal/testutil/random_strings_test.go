package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomStringMaker(t *testing.T) {
	rsm := NewRandomStringMaker()

	for i := 0; i < 1000; i++ {
		str := rsm.MakeSizedString(i)
		assert.Equal(t, i, len(str))
	}
}
