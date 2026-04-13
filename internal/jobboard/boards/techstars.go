package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewTechStarsBoard creates a Techstars board (ScrapeBoard).
func NewTechStarsBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.techstars.com/jobs",
		Name:     "Techstars",
		Slug:     "techstars",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.techstars.com/jobs", "/jobs/")
}
