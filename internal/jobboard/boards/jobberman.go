package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewJobbermanBoard creates a Jobberman board (ScrapeBoard).
func NewJobbermanBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.jobberman.com",
		Name:         "Jobberman",
		Slug:         "jobberman",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.jobberman.com/jobs", "/jobs/")
}
