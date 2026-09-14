//nolint:testpackage // tests unexported handlers (viewJob, requestWithID, securityHeaders) directly
package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type captureCreateRepo struct {
	job Job
}

func (r *captureCreateRepo) Get(context.Context, string) (Job, error) { return Job{}, nil }
func (r *captureCreateRepo) Create(_ context.Context, job *Job) error {
	r.job = *job
	return nil
}
func (r *captureCreateRepo) Delete(context.Context, string) error { return nil }
func (r *captureCreateRepo) Select(context.Context, SelectParams) ([]Job, error) {
	return nil, nil
}
func (r *captureCreateRepo) Update(context.Context, *Job) error { return nil }

func newTestServer(t *testing.T, dir string) *Server {
	t.Helper()

	srv, err := New(NewService(nil, dir), ":0")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return srv
}

func TestViewJobRendersPlaces(t *testing.T) {
	dir := t.TempDir()
	id := "11111111-1111-1111-1111-111111111111"

	csv := "title,latitude,longitude\nPlace,1.5,2.5\n"
	if err := os.WriteFile(filepath.Join(dir, id+".csv"), []byte(csv), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	srv := newTestServer(t, dir)

	req := requestWithID(httptest.NewRequest(http.MethodGet, "/view?id="+id, http.NoBody))
	rec := httptest.NewRecorder()
	srv.viewJob(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	for _, want := range []string{`id="map-modal"`, `initJobMap()`, `"title":"Place"`, `"latitude":1.5`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q:\n%s", want, body)
		}
	}
}

func TestViewJobEmptyState(t *testing.T) {
	srv := newTestServer(t, t.TempDir())

	id := "22222222-2222-2222-2222-222222222222"
	req := requestWithID(httptest.NewRequest(http.MethodGet, "/view?id="+id, http.NoBody))
	rec := httptest.NewRecorder()
	srv.viewJob(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "var places = [];") {
		t.Fatalf("expected empty places array, got:\n%s", body)
	}
}

func TestViewJobInvalidID(t *testing.T) {
	srv := newTestServer(t, t.TempDir())

	req := requestWithID(httptest.NewRequest(http.MethodGet, "/view?id=not-a-uuid", http.NoBody))
	rec := httptest.NewRecorder()
	srv.viewJob(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestSecurityHeadersAllowMapResources(t *testing.T) {
	handler := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	for _, want := range []string{"tile.openstreetmap.org", "cdnjs.cloudflare.com"} {
		if !strings.Contains(csp, want) {
			t.Fatalf("CSP missing %q: %s", want, csp)
		}
	}
}

func TestScrapeDefaultsToAutoModeWithoutTimedFields(t *testing.T) {
	t.Parallel()

	repo := &captureCreateRepo{}
	srv, err := New(NewService(repo, t.TempDir()), ":0")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	form := "name=Auto+job&run_mode=auto&keywords=coffee+in+East+London&lang=en&zoom=15&radius=10000&latitude=0&longitude=0"
	req := httptest.NewRequest(http.MethodPost, "/scrape", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.scrape(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if repo.job.Data.RunMode != RunModeAuto || repo.job.Data.MaxTime != 0 || repo.job.Data.Depth != 0 {
		t.Fatalf("unexpected auto job: %+v", repo.job.Data)
	}
	if !repo.job.Data.Email {
		t.Fatal("auto mode must enable email discovery")
	}
}
