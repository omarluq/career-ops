package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewTheMuseBoard creates a TheMuse board (ParamScrapeBoard).
func NewTheMuseBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "The Muse", Slug: "themuse", URL: "https://www.themuse.com/jobs",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.themuse.com/jobs", "themuse.com/jobs/", SearchParams{
		KeywordParam: "keyword", LocationParam: "location",
	})
}
