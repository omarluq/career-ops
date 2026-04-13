package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewArcBoard creates an Arc board (ScrapeBoard).
func NewArcBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://arc.dev",
		Name:         "Arc",
		Slug:         "arc",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://arc.dev/remote-jobs", "/remote-jobs/")
}
