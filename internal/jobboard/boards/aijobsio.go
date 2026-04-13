package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewAIJobsIOBoard creates an AI-Jobs.io board (ScrapeBoard).
func NewAIJobsIOBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://ai-jobs.io",
		Name:     "AI-Jobs.io",
		Slug:     "ai-jobs-io",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://ai-jobs.io/jobs", "/job/")
}
