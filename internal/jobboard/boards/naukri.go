package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewNaukriBoard creates a Naukri board (ScrapeBoard).
func NewNaukriBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.naukri.com",
		Name:         "Naukri",
		Slug:         "naukri",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.naukri.com/joblist", "/job-listings-")
}
