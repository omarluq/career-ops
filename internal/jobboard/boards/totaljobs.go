package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewTotaljobsBoard creates a Totaljobs board (ScrapeBoard).
func NewTotaljobsBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.totaljobs.com",
		Name:         "Totaljobs",
		Slug:         "totaljobs",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.totaljobs.com/jobs", "/job/")
}
