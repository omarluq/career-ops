package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewAshbyBoard creates an Ashby board (ApplyScrapeBoard).
func NewAshbyBoard() *ApplyScrapeBoard {
	return NewApplyScrapeBoard(jobboard.BoardMeta{
		URL:          "https://jobs.ashbyhq.com",
		Name:         "Ashby",
		Slug:         "ashby",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3, CooldownOnError: 15 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapApply, jobboard.CapScrape},
	}, "https://jobs.ashbyhq.com", "/jobs/")
}
