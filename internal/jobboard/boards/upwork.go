package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewUpworkBoard creates an Upwork board (ScrapeBoard).
func NewUpworkBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.upwork.com",
		Name:         "Upwork",
		Slug:         "upwork",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthSession,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.upwork.com/nx/search/jobs/", "/jobs/")
}
