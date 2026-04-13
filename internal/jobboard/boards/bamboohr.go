package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewBambooHRBoard creates a BambooHR board (ScrapeBoard).
func NewBambooHRBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.bamboohr.com/careers",
		Name:         "BambooHR",
		Slug:         "bamboohr",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.bamboohr.com/careers", "/careers/")
}
