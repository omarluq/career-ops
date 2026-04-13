package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewAIJobsNetBoard creates an AI-Jobs.net board (ScrapeBoard).
func NewAIJobsNetBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://ai-jobs.net",
		Name:     "AI-Jobs.net",
		Slug:     "ai-jobs-net",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://ai-jobs.net/jobs", "/job/")
}
