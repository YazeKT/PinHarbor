//nolint:testpackage // shares the internal web test package with web_test.go
package web

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type cancelAwareRepo struct {
	cancelled <-chan struct{}
	deleted   bool
}

type stopRepo struct {
	job       Job
	cancelled <-chan struct{}
}

func (r *stopRepo) Get(context.Context, string) (Job, error) { return r.job, nil }
func (r *stopRepo) Create(context.Context, *Job) error       { return nil }
func (r *stopRepo) Delete(context.Context, string) error     { return nil }
func (r *stopRepo) Select(context.Context, SelectParams) ([]Job, error) {
	return []Job{r.job}, nil
}
func (r *stopRepo) Update(_ context.Context, job *Job) error {
	select {
	case <-r.cancelled:
		r.job = *job
		return nil
	default:
		return errors.New("job updated before its worker was cancelled")
	}
}

func (r *cancelAwareRepo) Get(context.Context, string) (Job, error) {
	return Job{}, errors.New("not implemented")
}
func (r *cancelAwareRepo) Create(context.Context, *Job) error { return errors.New("not implemented") }
func (r *cancelAwareRepo) Select(context.Context, SelectParams) ([]Job, error) {
	return nil, errors.New("not implemented")
}
func (r *cancelAwareRepo) Update(context.Context, *Job) error { return errors.New("not implemented") }
func (r *cancelAwareRepo) Delete(context.Context, string) error {
	select {
	case <-r.cancelled:
		r.deleted = true
		return nil
	default:
		return errors.New("job deleted before its worker was cancelled")
	}
}

func writeCSV(t *testing.T, dir, id, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, id+".csv"), []byte(content), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}
}

func TestGetPlacesParsesCSV(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(nil, dir)

	csv := "title,address,latitude,longitude,link,category,phone,website,review_rating\n" +
		"Coffee Place,1 Main St,37.7749,-122.4194,http://maps/1,cafe,555,http://web,4.5\n"
	writeCSV(t, dir, "job-1", csv)

	places, err := svc.GetPlaces(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("GetPlaces: %v", err)
	}

	if len(places) != 1 {
		t.Fatalf("expected 1 place, got %d", len(places))
	}

	p := places[0]
	if p.Title != "Coffee Place" || p.Latitude != 37.7749 || p.Longitude != -122.4194 {
		t.Fatalf("unexpected place: %+v", p)
	}

	if p.ReviewRating != 4.5 {
		t.Fatalf("unexpected rating: %v", p.ReviewRating)
	}
}

func TestGetPlacesSkipsRowsWithoutCoords(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(nil, dir)

	csv := "title,latitude,longitude\n" +
		"No Coords,,\n" +
		"Zero,0,0\n" +
		"Bad,abc,def\n" +
		"Good,1.5,2.5\n"
	writeCSV(t, dir, "job-2", csv)

	places, err := svc.GetPlaces(context.Background(), "job-2")
	if err != nil {
		t.Fatalf("GetPlaces: %v", err)
	}

	if len(places) != 1 {
		t.Fatalf("expected 1 place, got %d", len(places))
	}

	if places[0].Title != "Good" {
		t.Fatalf("unexpected place: %+v", places[0])
	}
}

func TestGetPlacesSkipsNonFiniteAndOutOfRangeCoords(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(nil, dir)

	csv := "title,latitude,longitude,review_rating\n" +
		"NaN,NaN,2.5,3\n" +
		"Inf,Inf,2.5,3\n" +
		"OutOfRange,91,200,3\n" +
		"BadRating,1.5,2.5,NaN\n" +
		"Good,1.5,2.5,4.5\n"
	writeCSV(t, dir, "job-nf", csv)

	places, err := svc.GetPlaces(context.Background(), "job-nf")
	if err != nil {
		t.Fatalf("GetPlaces: %v", err)
	}

	if len(places) != 2 {
		t.Fatalf("expected 2 places (BadRating + Good), got %d: %+v", len(places), places)
	}

	for _, p := range places {
		if p.Title == "BadRating" && p.ReviewRating != 0 {
			t.Fatalf("non-finite rating should be sanitized to 0, got %v", p.ReviewRating)
		}
	}
}

func TestGetPlacesMissingCSV(t *testing.T) {
	svc := NewService(nil, t.TempDir())

	if _, err := svc.GetPlaces(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for missing csv")
	}
}

func TestGetPlacesRejectsTraversal(t *testing.T) {
	svc := NewService(nil, t.TempDir())

	if _, err := svc.GetPlaces(context.Background(), "../etc/passwd"); err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestDeleteCancelsAndWaitsForRunningJob(t *testing.T) {
	cancelled := make(chan struct{})
	repo := &cancelAwareRepo{cancelled: cancelled}
	svc := NewService(repo, t.TempDir())
	finish := svc.RegisterRunning("job-live", func() { close(cancelled) })

	go func() {
		<-cancelled
		time.Sleep(25 * time.Millisecond)
		finish()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := svc.Delete(ctx, "job-live"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !repo.deleted {
		t.Fatal("expected repository row to be deleted after cancellation")
	}
}

func TestStopCancelsWorkerAndPreservesPartialCSV(t *testing.T) {
	t.Parallel()

	cancelled := make(chan struct{})
	repo := &stopRepo{
		job:       Job{ID: "job-live", Status: StatusWorking},
		cancelled: cancelled,
	}
	dir := t.TempDir()
	svc := NewService(repo, dir)
	writeCSV(t, dir, "job-live", "title\nSaved result\n")
	finish := svc.RegisterRunning("job-live", func() { close(cancelled) })

	go func() {
		<-cancelled
		time.Sleep(25 * time.Millisecond)
		finish()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := svc.Stop(ctx, "job-live"); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if repo.job.Status != StatusStopped {
		t.Fatalf("status = %q, want stopped", repo.job.Status)
	}
	if _, err := os.Stat(filepath.Join(dir, "job-live.csv")); err != nil {
		t.Fatalf("partial CSV was not preserved: %v", err)
	}
}
