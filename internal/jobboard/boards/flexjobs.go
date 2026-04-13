package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewFlexJobsBoard creates a FlexJobs board (ScrapeBoard).
func NewFlexJobsBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.flexjobs.com",
		Name:     "FlexJobs",
		Slug:     "flexjobs",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthSession,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 5,
			BurstSize:         1,
			CooldownOnError:   60 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.flexjobs.com/search", "/job/")
}
