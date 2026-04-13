package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewLemonBoard creates a Lemon.io board (ScrapeBoard).
func NewLemonBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://lemon.io",
		Name:     "Lemon.io",
		Slug:     "lemon",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthSession,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 15,
			BurstSize:         3,
			CooldownOnError:   20 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://lemon.io/jobs", "/jobs/")
}
