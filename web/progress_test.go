package web

import (
	"strings"
	"testing"
	"time"
)

func TestJobViewUsesLiveCountersAndSavedRows(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	svc := NewService(nil, dir)
	id := "11111111-1111-1111-1111-111111111111"
	writeCSV(t, dir, id, "title,address,phone,emails\nFirst,One,1,a@example.com\nSecond,Two,2,b@example.com\n")
	finish := svc.RegisterRunning(id, func() {})
	defer finish()
	svc.SetRunningProgress(id, time.Hour, func() WorkSnapshot {
		return WorkSnapshot{SeedCount: 2, SeedCompleted: 1, PlacesFound: 4, PlacesCompleted: 2}
	})

	view := svc.jobView(Job{
		ID:     id,
		Name:   "test",
		Date:   time.Now().Add(-time.Minute),
		Status: StatusWorking,
		Data:   JobData{Keywords: []string{"one", "two"}},
	})

	if view.ResultCount != 2 {
		t.Fatalf("result count = %d, want 2", view.ResultCount)
	}
	if view.ProgressPercent != 50 {
		t.Fatalf("progress = %d, want 50", view.ProgressPercent)
	}
	if view.SearchProgress != "1/2 searches" || view.PlaceProgress != "2/4 places processed" {
		t.Fatalf("unexpected work labels: %q, %q", view.SearchProgress, view.PlaceProgress)
	}
	if len(view.LatestResults) != 2 || view.LatestResults[0].Title != "Second" {
		t.Fatalf("latest results are not newest first: %+v", view.LatestResults)
	}
	if !strings.Contains(view.ETA, "~") {
		t.Fatalf("ETA = %q, want a rate-based estimate", view.ETA)
	}
}

func TestReadLiveResultsDoesNotUseMissingColumns(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	svc := NewService(nil, dir)
	writeCSV(t, dir, "job", "title\nOnly title\n")

	count, results := svc.readLiveResults("job", 6)
	if count != 1 || len(results) != 1 {
		t.Fatalf("count = %d, results = %+v", count, results)
	}
	if results[0].Address != "" || results[0].Phone != "" || results[0].Emails != "" {
		t.Fatalf("missing columns reused another value: %+v", results[0])
	}
}

func TestCompletedJobShowsOneHundredPercent(t *testing.T) {
	t.Parallel()

	now := time.Now()
	view := NewService(nil, t.TempDir()).jobView(Job{
		ID:        "done",
		Date:      now.Add(-2 * time.Minute),
		UpdatedAt: now,
		Status:    StatusOK,
		Data:      JobData{Keywords: []string{"coffee"}},
	})
	if view.ProgressPercent != 100 || view.ProgressLabel != "Complete" {
		t.Fatalf("completed progress = %d %q", view.ProgressPercent, view.ProgressLabel)
	}
}
