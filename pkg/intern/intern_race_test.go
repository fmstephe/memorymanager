package intern

/*
import (
	"math/rand"
	"testing"
	"time"

	"github.com/fmstephe/offheap/testpkg/testutil"
)

func TestInterner_Race(t *testing.T) {
	interner := New(-1, -1)

	bytes, ints, floats := buildTestData(100_000, 1000)
	const goroutines = 100
	const iterations = 100_000_000

	for range goroutines {
		go internStrings(interner, bytes, iterations)
		go internInts(interner, ints, iterations)
		go internFloats(interner, floats, iterations)
	}
}

func internStrings(interner *StringInterner, bytes [][]byte, iterations int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		interner.GetFromBytes(bytes[r.Intn(len(bytes))])
	}
}

func internInts(interner *StringInterner, ints []int64, iterations int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		interner.GetFromInt64(ints[r.Intn(len(ints))])
	}
}

func internFloats(interner *StringInterner, ints []float64, iterations int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		interner.GetFromFloat64(ints[r.Intn(len(ints))])
	}
}

func buildTestData(size, bytesLimit int) ([][]byte, []int64, []float64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	bytes := make([][]byte, 0, size)
	rsm := testutil.NewRandomStringMaker()
	for range size {
		bytes = append(bytes, rsm.MakeSizedBytes(r.Intn(bytesLimit)))
	}

	ints := make([]int64, 0, size)
	for range size {
		ints = append(ints, int64(r.Uint64()))
	}

	floats := make([]float64, 0, size)
	for range size {
		floats = append(floats, r.NormFloat64())
	}

	return bytes, ints, floats
}
*/
