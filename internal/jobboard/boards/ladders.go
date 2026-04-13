package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewLaddersBoard creates a Ladders board (ParamScrapeBoard).
func NewLaddersBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Ladders", Slug: "ladders", URL: "https://www.theladders.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthSession,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 10, BurstSize: 2, CooldownOnError: 30 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.theladders.com/jobs/search-jobs", "theladders.com/job/", SearchParams{
		KeywordParam: "keywords", LocationParam: "location",
	})
}
