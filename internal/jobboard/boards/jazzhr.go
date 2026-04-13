package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewJazzHRBoard creates a JazzHR board (ScrapeBoard).
func NewJazzHRBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.applytojob.com",
		Name:         "JazzHR",
		Slug:         "jazzhr",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.applytojob.com", "/apply/")
}
