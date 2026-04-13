package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewUSAJobsBoard creates a USAJobs board (ParamScrapeBoard).
func NewUSAJobsBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "USAJobs", Slug: "usajobs", URL: "https://www.usajobs.gov",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthAPIKey,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 5, CooldownOnError: 15 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapAPI},
	}, "https://data.usajobs.gov/api/Search", "usajobs.gov/job/", SearchParams{
		KeywordParam: "Keyword", LocationParam: "LocationName",
	})
}
