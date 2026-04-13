package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewAndelaBoard creates an Andela board (ScrapeBoard).
func NewAndelaBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://andela.com",
		Name:         "Andela",
		Slug:         "andela",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthSession,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://andela.com/jobs", "/jobs/")
}
