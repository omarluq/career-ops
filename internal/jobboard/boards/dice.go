package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewDiceBoard creates a Dice board (ParamScrapeBoard).
func NewDiceBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Dice", Slug: "dice", URL: "https://www.dice.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.dice.com/jobs", "dice.com/job-detail/", SearchParams{
		KeywordParam: "q", LocationParam: "location",
		RemoteParam: "filters.isRemote", RemoteValue: "true",
	})
}
