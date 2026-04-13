package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewContraBoard creates a Contra board (ScrapeBoard).
func NewContraBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://contra.com",
		Name:         "Contra",
		Slug:         "contra",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://contra.com/search/opportunities", "/opportunity/")
}
