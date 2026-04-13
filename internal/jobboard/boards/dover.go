package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewDoverBoard creates a Dover board (ScrapeBoard).
func NewDoverBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://app.dover.com",
		Name:         "Dover",
		Slug:         "dover",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://app.dover.com/jobs", "/apply/")
}
