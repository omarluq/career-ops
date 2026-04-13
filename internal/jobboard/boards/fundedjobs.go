package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewFundedJobsBoard creates a Funded Club board (ScrapeBoard).
func NewFundedJobsBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://funded.club",
		Name:     "Funded Club",
		Slug:     "funded-club",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://funded.club/jobs", "/job/")
}
