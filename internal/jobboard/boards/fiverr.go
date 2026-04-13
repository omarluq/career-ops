package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewFiverrBoard creates a Fiverr board (ScrapeBoard).
func NewFiverrBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.fiverr.com",
		Name:         "Fiverr",
		Slug:         "fiverr",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.fiverr.com/search/gigs", "/gigs/")
}
