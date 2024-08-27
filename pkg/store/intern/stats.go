package intern

// A summary of the stats for a specific type of interned converter.
//
// UsedBytes stat is global across all converters.
//
// Total is sum across all shards of the fields in Stats.
//
// Shards holds the individual shard Stats.
type StatsSummary struct {
	UsedBytes int
	Total     Stats
	Shards    []Stats
}

// The statistics capturing the runtime behaviour of the interner.
//
// Returned indicates the number of previously interned strings that have
// been returned.
//
// Interned indicates the number of strings which have been interned.
//
// MaxLenExceeded indicates the number of strings not interned because they
// were too long.
//
// UsedBytesExceeded indicates the number of strings not interned because the
// global usedBytes limit was exceeded.
//
// HashCollision indicates the number of strings not interned because of a hash
// collision.
type Stats struct {
	Returned          int
	Interned          int
	MaxLenExceeded    int
	UsedBytesExceeded int
	HashCollision     int
}

func makeSummary(shards []Stats, controller *internController) StatsSummary {
	total := Stats{}

	for i := range shards {
		total.Returned += shards[i].Returned
		total.Interned += shards[i].Interned
		total.MaxLenExceeded += shards[i].MaxLenExceeded
		total.UsedBytesExceeded += shards[i].UsedBytesExceeded
		total.HashCollision += shards[i].HashCollision
	}

	return StatsSummary{
		UsedBytes: controller.getUsedBytes(),
		Total:     total,
		Shards:    shards,
	}
}
