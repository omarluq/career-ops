package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewClimateBaseBoard creates a Climatebase board (ScrapeBoard).
func NewClimateBaseBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://climatebase.org/jobs",
		Name:     "Climatebase",
		Slug:     "climatebase",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://climatebase.org/jobs", "/jobs/")
}
