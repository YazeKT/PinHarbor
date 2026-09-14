//nolint:testpackage // This test needs unexported hooks to avoid running a browser.
package webrunner

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"testing"
	"time"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/web"
	"github.com/gosom/scrapemate"
)

func TestScrapeJobMarksOKBeforeClosingMate(t *testing.T) {
	t.Parallel()

	repo := &memoryJobRepo{}
	svc := web.NewService(repo, t.TempDir())
	job := web.Job{
		ID:     "job-1",
		Name:   "coffee",
		Date:   time.Now().UTC(),
		Status: web.StatusPending,
		Data: web.JobData{
			Keywords: []string{"coffee"},
			Lang:     "en",
			Zoom:     15,
			Lat:      "37.7749",
			Lon:      "-122.4194",
			FastMode: true,
			Radius:   1000,
			Depth:    10,
			MaxTime:  time.Minute,
		},
	}

	if err := svc.Create(context.Background(), &job); err != nil {
		t.Fatalf("create job: %v", err)
	}

	w := &webrunner{
		svc: svc,
		cfg: &runner.Config{DataFolder: t.TempDir(), Concurrency: 1},
		setupMate: func(_ context.Context, _ io.Writer, _ *web.Job) (mateRunner, error) {
			return fakeMate{
				onClose: func() {
					got, err := svc.Get(context.Background(), job.ID)
					if err != nil {
						t.Fatalf("get job during close: %v", err)
					}
					if got.Status != web.StatusOK {
						t.Fatalf("status during close = %q, want %q", got.Status, web.StatusOK)
					}
				},
			}, nil
		},
	}

	if err := w.scrapeJob(context.Background(), &job); err != nil {
		t.Fatalf("scrape job: %v", err)
	}
}

func TestWorkMarksInterruptedJobsFailedOnEngineStart(t *testing.T) {
	repo := &memoryJobRepo{jobs: map[string]web.Job{
		"old-job": {ID: "old-job", Status: web.StatusWorking},
	}}
	w := &webrunner{svc: web.NewService(repo, t.TempDir())}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := w.work(ctx); err != nil {
		t.Fatalf("work: %v", err)
	}
	if got := repo.jobs["old-job"].Status; got != web.StatusFailed {
		t.Fatalf("interrupted job status = %q, want %q", got, web.StatusFailed)
	}
}

func TestResilientCSVWriterSkipsNilResults(t *testing.T) {
	var output bytes.Buffer
	csvWriter := newResilientCSVWriter(csv.NewWriter(&output))
	results := make(chan scrapemate.Result, 2)
	results <- scrapemate.Result{}
	results <- scrapemate.Result{Data: csvCapableTestRow{}}
	close(results)

	if err := csvWriter.Run(context.Background(), results); err != nil {
		t.Fatalf("write results: %v", err)
	}
	if got := output.String(); got != "name\nvalue\n" {
		t.Fatalf("csv output = %q", got)
	}
}

type csvCapableTestRow struct{}

func (csvCapableTestRow) CsvHeaders() []string { return []string{"name"} }
func (csvCapableTestRow) CsvRow() []string     { return []string{"value"} }

type fakeMate struct {
	onClose func()
}

func (m fakeMate) Start(context.Context, ...scrapemate.IJob) error {
	return nil
}

func (m fakeMate) Close() error {
	if m.onClose != nil {
		m.onClose()
	}

	return nil
}

type memoryJobRepo struct {
	jobs map[string]web.Job
}

func (r *memoryJobRepo) Get(_ context.Context, id string) (web.Job, error) {
	return r.jobs[id], nil
}

func (r *memoryJobRepo) Create(_ context.Context, job *web.Job) error {
	if r.jobs == nil {
		r.jobs = make(map[string]web.Job)
	}

	r.jobs[job.ID] = *job

	return nil
}

func (r *memoryJobRepo) Delete(_ context.Context, id string) error {
	delete(r.jobs, id)
	return nil
}

func (r *memoryJobRepo) Select(_ context.Context, params web.SelectParams) ([]web.Job, error) {
	jobs := make([]web.Job, 0, len(r.jobs))

	for id := range r.jobs {
		job := r.jobs[id]
		if params.Status == "" || job.Status == params.Status {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func (r *memoryJobRepo) Update(_ context.Context, job *web.Job) error {
	r.jobs[job.ID] = *job
	return nil
}
