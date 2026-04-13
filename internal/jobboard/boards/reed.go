package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewReedBoard creates a Reed board (ParamScrapeBoard).
func NewReedBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Reed", Slug: "reed", URL: "https://www.reed.co.uk",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.reed.co.uk/jobs", "reed.co.uk/jobs/", SearchParams{
		KeywordParam: "keywords", LocationParam: "location",
	})
}
