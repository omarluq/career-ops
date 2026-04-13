package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewRandstadBoard creates a Randstad board (ParamScrapeBoard).
func NewRandstadBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Randstad", Slug: "randstad", URL: "https://www.randstad.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.randstad.com/jobs/search", "randstad.com/jobs/", SearchParams{
		KeywordParam: "keywords", LocationParam: "location",
	})
}
