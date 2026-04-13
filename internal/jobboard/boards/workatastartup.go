package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewWorkAtAStartup2Board creates a YC Jobs (Direct) board (ScrapeBoard).
func NewWorkAtAStartup2Board() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.ycombinator.com/jobs",
		Name:     "YC Jobs (Direct)",
		Slug:     "yc-jobs-direct",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.ycombinator.com/jobs", "/companies/")
}
