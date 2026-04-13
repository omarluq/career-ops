package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewPinpointBoard creates a Pinpoint board (ScrapeBoard).
func NewPinpointBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.pinpointhq.com",
		Name:         "Pinpoint",
		Slug:         "pinpoint",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.pinpointhq.com", "/jobs/")
}
