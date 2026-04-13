package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewICIMSBoard creates an iCIMS board (ScrapeBoard).
func NewICIMSBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.icims.com",
		Name:         "iCIMS",
		Slug:         "icims",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.icims.com/jobs", "/jobs/")
}
