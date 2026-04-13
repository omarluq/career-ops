package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewCareerBuilderBoard creates a CareerBuilder board (ParamScrapeBoard).
func NewCareerBuilderBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "CareerBuilder", Slug: "careerbuilder", URL: "https://www.careerbuilder.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.careerbuilder.com/jobs", "careerbuilder.com/job/", SearchParams{
		KeywordParam: "keywords", LocationParam: "location",
	})
}
