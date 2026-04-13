package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewAIJobBoardBoard creates an AI Job Board board (ScrapeBoard).
func NewAIJobBoardBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://aijobs.ai",
		Name:     "AI Job Board",
		Slug:     "aijobs-ai",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://aijobs.ai/jobs", "/job/")
}
