package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewRipplingBoard creates a Rippling board (ScrapeBoard).
func NewRipplingBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://ats.rippling.com",
		Name:         "Rippling",
		Slug:         "rippling",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://ats.rippling.com/jobs", "/job/")
}
