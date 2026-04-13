package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewF6SBoard creates an F6S board (ScrapeBoard).
func NewF6SBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.f6s.com/jobs",
		Name:     "F6S",
		Slug:     "f6s",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.f6s.com/jobs", "/jobs/")
}
