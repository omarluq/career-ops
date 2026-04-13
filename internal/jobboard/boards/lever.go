package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewLeverBoard creates a Lever board (ApplyScrapeBoard).
func NewLeverBoard() *ApplyScrapeBoard {
	return NewApplyScrapeBoard(jobboard.BoardMeta{
		URL:          "https://jobs.lever.co",
		Name:         "Lever",
		Slug:         "lever",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3, CooldownOnError: 15 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapApply, jobboard.CapScrape},
	}, "https://jobs.lever.co", "/jobs/")
}
