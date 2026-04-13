package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewCakeResumeBoard creates a CakeResume board (ScrapeBoard).
func NewCakeResumeBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.cakeresume.com/jobs",
		Name:         "CakeResume",
		Slug:         "cakeresume",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.cakeresume.com/jobs", "/jobs/")
}
