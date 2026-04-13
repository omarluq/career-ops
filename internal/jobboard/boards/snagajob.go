package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewSnagajobBoard creates a Snagajob board (ParamScrapeBoard).
func NewSnagajobBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Snagajob", Slug: "snagajob", URL: "https://www.snagajob.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.snagajob.com/jobs", "snagajob.com/jobs/", SearchParams{
		KeywordParam: "q", LocationParam: "w",
	})
}
