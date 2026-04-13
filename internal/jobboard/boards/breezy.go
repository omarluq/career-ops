package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewBreezyBoard creates a Breezy HR board (ScrapeBoard).
func NewBreezyBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.breezy.hr",
		Name:         "Breezy",
		Slug:         "breezy",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.breezy.hr", "/p/")
}
