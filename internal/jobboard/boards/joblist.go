package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewJoblistBoard creates a Joblist board (ParamScrapeBoard).
func NewJoblistBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Joblist", Slug: "joblist", URL: "https://www.joblist.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.joblist.com/search", "joblist.com/listing/", SearchParams{
		KeywordParam: "q", LocationParam: "l",
	})
}
