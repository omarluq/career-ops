package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewSeekBoard creates a Seek board (ParamScrapeBoard).
func NewSeekBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Seek", Slug: "seek", URL: "https://www.seek.com.au",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.seek.com.au/jobs", "seek.com.au/job/", SearchParams{
		KeywordParam: "keywords", LocationParam: "where",
	})
}
