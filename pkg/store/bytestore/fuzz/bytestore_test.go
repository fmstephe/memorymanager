package fuzz

import (
	"math/rand"
	"testing"
)

func FuzzObjectStore(f *testing.F) {
	r := rand.New(rand.NewSource(1))
	testCases := [][]byte{
		randomBytes(r, 0),
		randomBytes(r, 10),
		randomBytes(r, 100),
		randomBytes(r, 1000),
		randomBytes(r, 10000),
		randomBytes(r, 100000),
	}
	for _, tc := range testCases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, bytes []byte) {
		tr := NewTestRun(bytes)
		tr.Run()
	})
}

func randomBytes(r *rand.Rand, size int) []byte {
	bytes := make([]byte, size)
	r.Read(bytes)
	return bytes
}
