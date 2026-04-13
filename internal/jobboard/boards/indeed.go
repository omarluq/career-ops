package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewIndeedBoard creates an Indeed board (ParamScrapeBoard).
func NewIndeedBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Indeed", Slug: "indeed", URL: "https://www.indeed.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.indeed.com/jobs", "indeed.com/viewjob", SearchParams{
		KeywordParam: "q", LocationParam: "l",
		RemoteParam: "remotejob", RemoteValue: "1",
	})
}
