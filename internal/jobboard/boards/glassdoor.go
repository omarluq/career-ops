package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewGlassdoorBoard creates a Glassdoor board (ParamScrapeBoard).
func NewGlassdoorBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Glassdoor", Slug: "glassdoor", URL: "https://www.glassdoor.com/Job",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthSession,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 10, BurstSize: 2, CooldownOnError: 30 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.glassdoor.com/Job/jobs.htm", "glassdoor.com/job-listing/", SearchParams{
		KeywordParam: "sc.keyword", LocationParam: "locT",
	})
}
