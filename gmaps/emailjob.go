package gmaps

import (
	"context"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/gosom/scrapemate"
	"github.com/mcnijman/go-emailaddress"

	"github.com/gosom/google-maps-scraper/exiter"
)

type EmailExtractJobOptions func(*EmailExtractJob)

type EmailExtractJob struct {
	scrapemate.Job

	Entry                   *Entry
	ExitMonitor             exiter.Exiter
	WriterManagedCompletion bool
	FollowupDepth           int
}

func NewEmailJob(parentID string, entry *Entry, opts ...EmailExtractJobOptions) *EmailExtractJob {
	const (
		defaultPrio       = scrapemate.PriorityHigh
		defaultMaxRetries = 2
	)

	job := EmailExtractJob{
		Job: scrapemate.Job{
			ID:         uuid.New().String(),
			ParentID:   parentID,
			Method:     "GET",
			URL:        normalizeGoogleURL(entry.WebSite),
			MaxRetries: defaultMaxRetries,
			Priority:   defaultPrio,
		},
	}

	job.Entry = entry

	for _, opt := range opts {
		opt(&job)
	}

	return &job
}

func WithEmailJobExitMonitor(exitMonitor exiter.Exiter) EmailExtractJobOptions {
	return func(j *EmailExtractJob) {
		j.ExitMonitor = exitMonitor
	}
}

func WithEmailJobWriterManagedCompletion() EmailExtractJobOptions {
	return func(j *EmailExtractJob) {
		j.WriterManagedCompletion = true
	}
}

func withEmailJobURL(targetURL string, depth int) EmailExtractJobOptions {
	return func(j *EmailExtractJob) {
		j.URL = targetURL
		j.FollowupDepth = depth
	}
}

func (j *EmailExtractJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	complete := true
	defer func() {
		resp.Document = nil
		resp.Body = nil
	}()

	defer func() {
		if complete && j.ExitMonitor != nil && !j.WriterManagedCompletion {
			j.ExitMonitor.IncrPlacesCompleted(1)
		}
	}()

	log := scrapemate.GetLoggerFromContext(ctx)

	log.Info("Processing email job", "url", j.URL)

	// if html fetch failed just return
	if resp.Error != nil {
		return j.Entry, nil, nil
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return j.Entry, nil, nil
	}

	emails := docEmailExtractor(doc)
	if len(emails) == 0 {
		emails = regexEmailExtractor(resp.Body)
	}
	if len(emails) == 0 && j.FollowupDepth == 0 {
		if followupURL := findEmailFollowupURL(resp.URL, doc); followupURL != "" {
			opts := []EmailExtractJobOptions{withEmailJobURL(followupURL, 1)}
			if j.ExitMonitor != nil {
				opts = append(opts, WithEmailJobExitMonitor(j.ExitMonitor))
			}
			if j.WriterManagedCompletion {
				opts = append(opts, WithEmailJobWriterManagedCompletion())
			}
			complete = false

			return nil, []scrapemate.IJob{NewEmailJob(j.ParentID, j.Entry, opts...)}, nil
		}
	}

	j.Entry.Emails = emails

	return j.Entry, nil, nil
}

func findEmailFollowupURL(rawBaseURL string, doc *goquery.Document) string {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return ""
	}

	var best string
	doc.Find("a[href]").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		href, ok := selection.Attr("href")
		if !ok {
			return true
		}

		label := strings.ToLower(strings.TrimSpace(selection.Text()) + " " + href)
		if !strings.Contains(label, "contact") && !strings.Contains(label, "get-in-touch") &&
			!strings.Contains(label, "about") && !strings.Contains(label, "team") {
			return true
		}

		candidate, err := url.Parse(href)
		if err != nil || candidate.Scheme == "mailto" || candidate.Scheme == "tel" {
			return true
		}

		candidate = baseURL.ResolveReference(candidate)
		if candidate.Host != baseURL.Host {
			return true
		}

		best = candidate.String()

		return false
	})

	return best
}

func (j *EmailExtractJob) ProcessOnFetchError() bool {
	return true
}

func docEmailExtractor(doc *goquery.Document) []string {
	seen := map[string]bool{}

	var emails []string

	doc.Find("a[href^='mailto:']").Each(func(_ int, s *goquery.Selection) {
		mailto, exists := s.Attr("href")
		if exists {
			value := strings.TrimPrefix(mailto, "mailto:")
			if email, err := getValidEmail(value); err == nil {
				if !seen[email] {
					emails = append(emails, email)
					seen[email] = true
				}
			}
		}
	})

	return emails
}

func regexEmailExtractor(body []byte) []string {
	seen := map[string]bool{}

	var emails []string

	addresses := emailaddress.Find(body, false)
	for i := range addresses {
		if !seen[addresses[i].String()] {
			emails = append(emails, addresses[i].String())
			seen[addresses[i].String()] = true
		}
	}

	return emails
}

func getValidEmail(s string) (string, error) {
	email, err := emailaddress.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}

	return email.String(), nil
}

// normalizeGoogleURL extracts the actual target URL from Google redirect URLs.
// Google Maps sometimes returns URLs like "/url?q=http://example.com/&opi=..."
// for external website links.
func normalizeGoogleURL(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}

	if strings.HasPrefix(rawURL, "/url?q=") {
		fullURL := "https://www.google.com" + rawURL

		parsed, err := url.Parse(fullURL)
		if err != nil {
			return rawURL
		}

		if target := parsed.Query().Get("q"); target != "" {
			return target
		}
	}

	if strings.HasPrefix(rawURL, "/") {
		return "https://www.google.com" + rawURL
	}

	return rawURL
}
