package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewWorkableBoard creates a Workable board (ApplyScrapeBoard).
func NewWorkableBoard() *ApplyScrapeBoard {
	return NewApplyScrapeBoard(jobboard.BoardMeta{
		URL:          "https://apply.workable.com",
		Name:         "Workable",
		Slug:         "workable",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3, CooldownOnError: 15 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapApply, jobboard.CapScrape},
	}, "https://apply.workable.com", "/apply/")
}
