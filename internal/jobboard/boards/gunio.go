package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewGunIOBoard creates a Gun.io board (ScrapeBoard).
func NewGunIOBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://gun.io",
		Name:         "Gun.io",
		Slug:         "gunio",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://gun.io/find-work/", "/find-work/")
}
