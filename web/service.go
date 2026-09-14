package web

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Service struct {
	repo       JobRepository
	dataFolder string
	runningMu  sync.Mutex
	running    map[string]*runningJob
}

type runningJob struct {
	cancel      context.CancelFunc
	done        chan struct{}
	once        sync.Once
	startedAt   time.Time
	maxDuration time.Duration
	snapshot    func() WorkSnapshot
}

// WorkSnapshot contains exact counters reported by the active scraper.
type WorkSnapshot struct {
	SeedCount       int
	SeedCompleted   int
	PlacesFound     int
	PlacesCompleted int
}

func NewService(repo JobRepository, dataFolder string) *Service {
	return &Service{
		repo:       repo,
		dataFolder: dataFolder,
		running:    make(map[string]*runningJob),
	}
}

// RegisterRunning links a job ID to its live scraper context. The returned
// function must be deferred by the runner after files and browser resources
// have been released.
func (s *Service) RegisterRunning(id string, cancel context.CancelFunc) func() {
	control := &runningJob{cancel: cancel, done: make(chan struct{}), startedAt: time.Now()}

	s.runningMu.Lock()
	s.running[id] = control
	s.runningMu.Unlock()

	return func() {
		s.runningMu.Lock()
		if s.running[id] == control {
			delete(s.running, id)
		}
		s.runningMu.Unlock()
		control.once.Do(func() { close(control.done) })
	}
}

// SetRunningProgress attaches the scraper's live work counters to a registered
// job. The source is read only while rendering status and must be thread-safe.
func (s *Service) SetRunningProgress(id string, maxDuration time.Duration, snapshot func() WorkSnapshot) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if control := s.running[id]; control != nil {
		control.maxDuration = maxDuration
		control.snapshot = snapshot
	}
}

func (s *Service) runningProgress(id string) (time.Time, time.Duration, WorkSnapshot, bool) {
	s.runningMu.Lock()
	control := s.running[id]
	if control == nil {
		s.runningMu.Unlock()
		return time.Time{}, 0, WorkSnapshot{}, false
	}
	startedAt := control.startedAt
	maxDuration := control.maxDuration
	snapshot := control.snapshot
	s.runningMu.Unlock()

	var stats WorkSnapshot
	if snapshot != nil {
		stats = snapshot()
	}

	return startedAt, maxDuration, stats, true
}

func (s *Service) stopRunning(ctx context.Context, id string) error {
	s.runningMu.Lock()
	control := s.running[id]
	s.runningMu.Unlock()
	if control == nil {
		return nil
	}

	control.cancel()
	select {
	case <-control.done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timed out stopping active scrape %s: %w", id, ctx.Err())
	}
}

func (s *Service) Create(ctx context.Context, job *Job) error {
	return s.repo.Create(ctx, job)
}

func (s *Service) All(ctx context.Context) ([]Job, error) {
	return s.repo.Select(ctx, SelectParams{})
}

func (s *Service) Get(ctx context.Context, id string) (Job, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	stopCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := s.stopRunning(stopCtx, id); err != nil {
		return err
	}

	datapath, err := s.csvPath(id)
	if err != nil {
		return err
	}

	if _, err := os.Stat(datapath); err == nil {
		if err := os.Remove(datapath); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return s.repo.Delete(ctx, id)
}

// Stop cancels an active collection and preserves its job row and partial CSV.
func (s *Service) Stop(ctx context.Context, id string) error {
	job, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if job.Status == StatusOK || job.Status == StatusFailed || job.Status == StatusStopped {
		return nil
	}

	stopCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := s.stopRunning(stopCtx, id); err != nil {
		return err
	}

	job, err = s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	job.Status = StatusStopped
	return s.repo.Update(ctx, &job)
}

func (s *Service) Update(ctx context.Context, job *Job) error {
	return s.repo.Update(ctx, job)
}

func (s *Service) SelectPending(ctx context.Context) ([]Job, error) {
	return s.repo.Select(ctx, SelectParams{Status: StatusPending, Limit: 1})
}

// MarkInterruptedJobsFailed ensures a newly started scraper engine never
// presents work from an earlier process as still running. A running job cannot
// be resumed safely after its browser and worker context have exited.
func (s *Service) MarkInterruptedJobsFailed(ctx context.Context) error {
	jobs, err := s.repo.Select(ctx, SelectParams{Status: StatusWorking})
	if err != nil {
		return err
	}

	for i := range jobs {
		jobs[i].Status = StatusFailed
		if err := s.repo.Update(ctx, &jobs[i]); err != nil {
			return err
		}
	}

	return nil
}

// csvPath returns the on-disk path of a job's CSV output, rejecting ids that
// could escape the data folder.
func (s *Service) csvPath(id string) (string, error) {
	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "..") {
		return "", fmt.Errorf("invalid file name")
	}

	return filepath.Join(s.dataFolder, id+".csv"), nil
}

func (s *Service) GetCSV(_ context.Context, id string) (string, error) {
	datapath, err := s.csvPath(id)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(datapath); os.IsNotExist(err) {
		return "", fmt.Errorf("csv file not found for job %s", id)
	}

	return datapath, nil
}
