package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewStepStoneBoard creates a StepStone board (ParamScrapeBoard).
func NewStepStoneBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "StepStone", Slug: "stepstone", URL: "https://www.stepstone.de",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.stepstone.de/jobs", "stepstone.de/stellenangebote/", SearchParams{
		KeywordParam: "what", LocationParam: "where",
	})
}
