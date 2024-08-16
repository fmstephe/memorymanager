package intern

type StatsSummary struct {
	UsedBytes int
	Total     Stats
	Shards    []Stats
}

type Stats struct {
	returned          int
	interned          int
	maxLenExceeded    int
	usedBytesExceeded int
	hashCollision     int
}

func makeSummary(shards []Stats, controller *internController) StatsSummary {
	total := Stats{}

	for i := range shards {
		total.returned += shards[i].returned
		total.interned += shards[i].interned
		total.maxLenExceeded += shards[i].maxLenExceeded
		total.usedBytesExceeded += shards[i].usedBytesExceeded
		total.hashCollision += shards[i].hashCollision
	}

	return StatsSummary{
		UsedBytes: controller.getUsedBytes(),
		Total:     total,
		Shards:    shards,
	}
}
