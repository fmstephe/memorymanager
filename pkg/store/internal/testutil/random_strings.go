package testutil

import (
	"math/rand"
	"strings"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type RandomStringMaker struct {
	r *rand.Rand
}

func NewRandomStringMaker() *RandomStringMaker {
	return &RandomStringMaker{
		r: rand.New(rand.NewSource(1)),
	}
}

func (rsm *RandomStringMaker) MakeSizedString(length int) string {
	builder := strings.Builder{}
	builder.Grow(length)
	for range length {
		builder.WriteByte(letters[rsm.r.Intn(len(letters))])
	}
	return builder.String()
}
