package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewWantedlyBoard creates a Wantedly board (ScrapeBoard).
func NewWantedlyBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.wantedly.com",
		Name:         "Wantedly",
		Slug:         "wantedly",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.wantedly.com/search", "/projects/")
}
