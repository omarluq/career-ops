package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewBaytBoard creates a Bayt board (ScrapeBoard).
func NewBaytBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.bayt.com",
		Name:         "Bayt",
		Slug:         "bayt",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.bayt.com/en/jobs/", "/en/job/")
}
