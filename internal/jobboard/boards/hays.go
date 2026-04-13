package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewHaysBoard creates a Hays board (ScrapeBoard).
func NewHaysBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.hays.com",
		Name:         "Hays",
		Slug:         "hays",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.hays.com/en/job-search", "/job/")
}
