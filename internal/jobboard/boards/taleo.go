package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewTaleoBoard creates a Taleo board (ScrapeBoard).
func NewTaleoBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.taleo.net",
		Name:         "Taleo",
		Slug:         "taleo",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 10, BurstSize: 2, CooldownOnError: 30 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.taleo.net/careersection/jobsearch.ftl", "/requisition/")
}
