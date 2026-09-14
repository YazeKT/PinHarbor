package web

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const liveResultLimit = 6

// LiveResult is one result that has actually been flushed to the job CSV.
type LiveResult struct {
	Title   string
	Address string
	Phone   string
	Emails  string
}

// JobView is the compact, display-ready status for a scrape job.
type JobView struct {
	ID               string
	ShortID          string
	Name             string
	RunModeLabel     string
	Status           string
	CreatedLocal     string
	UpdatedLocal     string
	Elapsed          string
	ResultCount      int
	Speed            string
	ETA              string
	ProgressPercent  int
	ProgressLabel    string
	SearchProgress   string
	PlaceProgress    string
	LatestResults    []LiveResult
	LastRefreshLocal string
}

func (s *Service) jobViews(ctxJobs []Job) []JobView {
	views := make([]JobView, 0, len(ctxJobs))
	for i := range ctxJobs {
		views = append(views, s.jobView(ctxJobs[i]))
	}

	return views
}

func (s *Service) jobView(job Job) JobView {
	now := time.Now()
	resultCount, latest := s.readLiveResults(job.ID, liveResultLimit)
	view := JobView{
		ID:               job.ID,
		ShortID:          shortJobID(job.ID),
		Name:             job.Name,
		RunModeLabel:     runModeLabel(EffectiveRunMode(job.Data)),
		Status:           job.Status,
		CreatedLocal:     localDateTime(job.Date),
		UpdatedLocal:     localDateTime(job.UpdatedAt),
		ResultCount:      resultCount,
		LatestResults:    latest,
		LastRefreshLocal: now.Format("15:04:05"),
		ETA:              "Waiting to start",
		Elapsed:          "—",
		Speed:            "0/min",
		ProgressLabel:    "Queued",
		SearchProgress:   fmt.Sprintf("0/%d searches", len(job.Data.Keywords)),
		PlaceProgress:    "0 places processed",
	}

	startedAt, maxDuration, stats, running := s.runningProgress(job.ID)
	if running {
		elapsed := now.Sub(startedAt)
		view.Elapsed = compactDuration(elapsed)
		view.Speed = scrapeSpeed(resultCount, elapsed)
		view.SearchProgress = fmt.Sprintf("%d/%d searches", stats.SeedCompleted, stats.SeedCount)
		view.PlaceProgress = fmt.Sprintf("%d/%d places processed", stats.PlacesCompleted, stats.PlacesFound)

		totalWork := stats.SeedCount + stats.PlacesFound
		completedWork := stats.SeedCompleted + stats.PlacesCompleted
		if completedWork > totalWork {
			completedWork = totalWork
		}
		if totalWork > 0 {
			view.ProgressPercent = completedWork * 100 / totalWork
			if view.ProgressPercent >= 100 && job.Status == StatusWorking {
				view.ProgressPercent = 99
			}
			view.ProgressLabel = fmt.Sprintf("%d/%d tasks", completedWork, totalWork)
		}

		view.ETA = estimateCompletion(now, startedAt, maxDuration, completedWork, totalWork)
	}

	if job.Status == StatusOK {
		finishedAt := job.UpdatedAt
		if finishedAt.IsZero() {
			finishedAt = now
		}
		elapsed := finishedAt.Sub(job.Date)
		view.Elapsed = compactDuration(elapsed)
		view.Speed = scrapeSpeed(resultCount, elapsed)
		view.ProgressPercent = 100
		view.ProgressLabel = "Complete"
		view.ETA = "Done " + finishedAt.In(time.Local).Format("15:04:05")
		view.SearchProgress = fmt.Sprintf("%d/%d searches", len(job.Data.Keywords), len(job.Data.Keywords))
		view.PlaceProgress = fmt.Sprintf("%d results saved", resultCount)
	}

	if job.Status == StatusFailed || job.Status == StatusStopped {
		finishedAt := job.UpdatedAt
		if finishedAt.IsZero() {
			finishedAt = now
		}
		elapsed := finishedAt.Sub(job.Date)
		view.Elapsed = compactDuration(elapsed)
		view.Speed = scrapeSpeed(resultCount, elapsed)
		view.ETA = "Stopped"
		if job.Status == StatusStopped {
			view.ProgressLabel = "Stopped safely"
		} else {
			view.ProgressLabel = "Incomplete"
		}
		view.PlaceProgress = fmt.Sprintf("%d results saved", resultCount)
	}

	return view
}

func runModeLabel(value string) string {
	switch value {
	case RunModeQuick:
		return "Quick"
	case RunModeBalanced:
		return "Balanced"
	case RunModeDeep:
		return "Deep"
	case RunModeCustom:
		return "Custom"
	default:
		return "Auto complete"
	}
}

func (s *Service) readLiveResults(id string, limit int) (int, []LiveResult) {
	path, err := s.csvPath(id)
	if err != nil {
		return 0, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	headers, err := reader.Read()
	if err != nil {
		return 0, nil
	}

	indexes := make(map[string]int, len(headers))
	for i, header := range headers {
		indexes[strings.ToLower(strings.TrimSpace(header))] = i
	}

	count := 0
	recent := make([]LiveResult, 0, limit)
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) || errors.Is(readErr, io.ErrUnexpectedEOF) {
			break
		}
		if readErr != nil {
			continue
		}

		count++
		item := LiveResult{
			Title:   csvColumn(row, indexes, "title"),
			Address: csvColumn(row, indexes, "address"),
			Phone:   csvColumn(row, indexes, "phone"),
			Emails:  csvColumn(row, indexes, "emails"),
		}
		if len(recent) == limit {
			copy(recent, recent[1:])
			recent[len(recent)-1] = item
		} else {
			recent = append(recent, item)
		}
	}

	for left, right := 0, len(recent)-1; left < right; left, right = left+1, right-1 {
		recent[left], recent[right] = recent[right], recent[left]
	}

	return count, recent
}

func csvColumn(row []string, indexes map[string]int, name string) string {
	index, ok := indexes[name]
	if !ok {
		return ""
	}
	if index < 0 || index >= len(row) {
		return ""
	}

	return row[index]
}

func shortJobID(id string) string {
	if len(id) <= 8 {
		return id
	}

	return id[:8]
}

func localDateTime(value time.Time) string {
	if value.IsZero() {
		return "—"
	}

	return value.In(time.Local).Format("2006-01-02 15:04:05")
}

func compactDuration(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	seconds := int(value.Round(time.Second).Seconds())
	hours := seconds / 3600
	minutes := seconds % 3600 / 60
	seconds %= 60
	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %02ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}

func scrapeSpeed(results int, elapsed time.Duration) string {
	if results == 0 || elapsed <= 0 {
		return "0/min"
	}
	perMinute := float64(results) / elapsed.Minutes()
	precision := 1
	if perMinute < 1 {
		precision = 2
	}

	return strconv.FormatFloat(perMinute, 'f', precision, 64) + "/min"
}

func estimateCompletion(now, startedAt time.Time, maxDuration time.Duration, completed, total int) string {
	if completed > 0 && total > completed {
		elapsed := now.Sub(startedAt)
		remaining := time.Duration(float64(elapsed) * float64(total-completed) / float64(completed))
		estimatedAt := now.Add(remaining)
		if maxDuration > 0 && estimatedAt.After(startedAt.Add(maxDuration)) {
			return "By limit " + startedAt.Add(maxDuration).Format("15:04")
		}

		return "~" + estimatedAt.Format("15:04") + " (" + compactDuration(remaining) + ")"
	}
	if total > 0 && completed >= total {
		return "Finalising…"
	}
	if maxDuration > 0 {
		return "Calculating · limit " + startedAt.Add(maxDuration).Format("15:04")
	}

	return "Calculating…"
}
