package intern

type StatsSummary struct {
	UsedBytes int
	Total     Stats
	Shards    []Stats
}

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
