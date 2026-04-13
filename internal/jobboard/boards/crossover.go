package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewCrossoverBoard creates a Crossover board (ScrapeBoard).
func NewCrossoverBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.crossover.com",
		Name:         "Crossover",
		Slug:         "crossover",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.crossover.com/jobs", "/jobs/")
}
