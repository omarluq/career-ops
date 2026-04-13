package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewFountainBoard creates a Fountain board (ScrapeBoard).
func NewFountainBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.fountain.com",
		Name:         "Fountain",
		Slug:         "fountain",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.fountain.com/jobs", "/apply/")
}
