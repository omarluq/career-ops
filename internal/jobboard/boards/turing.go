package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewTuringBoard creates a Turing board (ScrapeBoard).
func NewTuringBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.turing.com",
		Name:         "Turing",
		Slug:         "turing",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthSession,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.turing.com/remote-developer-jobs", "/remote-developer-jobs/")
}
