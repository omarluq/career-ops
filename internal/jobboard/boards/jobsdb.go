package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewJobsDBBoard creates a JobsDB board (ScrapeBoard).
func NewJobsDBBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.jobsdb.com",
		Name:         "JobsDB",
		Slug:         "jobsdb",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.jobsdb.com/jobs", "/job/")
}
