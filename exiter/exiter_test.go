package exiter

import "testing"

func TestSnapshotReturnsCurrentCounters(t *testing.T) {
	t.Parallel()

	monitor := New()
	monitor.SetSeedCount(3)
	monitor.IncrSeedCompleted(2)
	monitor.IncrPlacesFound(7)
	monitor.IncrPlacesCompleted(4)

	stats := monitor.Snapshot()
	if stats.SeedCount != 3 || stats.SeedCompleted != 2 || stats.PlacesFound != 7 || stats.PlacesCompleted != 4 {
		t.Fatalf("unexpected snapshot: %+v", stats)
	}
}
