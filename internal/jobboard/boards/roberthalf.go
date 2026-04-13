package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewRobertHalfBoard creates a Robert Half board (ParamScrapeBoard).
func NewRobertHalfBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Robert Half", Slug: "roberthalf", URL: "https://www.roberthalf.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.roberthalf.com/us/en/jobs", "roberthalf.com/us/en/job/", SearchParams{
		KeywordParam: "query", LocationParam: "location",
	})
}
