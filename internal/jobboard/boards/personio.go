package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewPersonioBoard creates a Personio board (ScrapeBoard).
func NewPersonioBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.personio.com",
		Name:         "Personio",
		Slug:         "personio",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.personio.com/jobs", "/job/")
}
