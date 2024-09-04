package intern

type Interner[T any] interface {
	Get(t T) string
	GetStats() StatsSummary
}
