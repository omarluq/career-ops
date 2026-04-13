package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewAdzunaBoard creates an Adzuna board (ParamScrapeBoard).
func NewAdzunaBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Adzuna", Slug: "adzuna", URL: "https://www.adzuna.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthAPIKey,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 5, CooldownOnError: 15 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapAPI},
	}, "https://api.adzuna.com/v1/api/jobs/us/search/1", "adzuna.com/details/", SearchParams{
		KeywordParam: "what", LocationParam: "where",
	})
}
