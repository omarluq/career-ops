package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewWorkdayBoard creates a Workday board (ScrapeBoard).
func NewWorkdayBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.myworkdayjobs.com",
		Name:         "Workday",
		Slug:         "workday",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 10, BurstSize: 2, CooldownOnError: 30 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.myworkdayjobs.com", "/job/")
}
