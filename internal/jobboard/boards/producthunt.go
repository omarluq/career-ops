package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewProductHuntBoard creates a Product Hunt board (remote-aware ScrapeBoard).
func NewProductHuntBoard() *ScrapeBoard {
	return NewRemoteScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.producthunt.com/jobs",
		Name:     "Product Hunt Jobs",
		Slug:     "producthunt",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.producthunt.com/jobs", "/jobs/")
}
