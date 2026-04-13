package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewMonsterBoard creates a Monster board (ParamScrapeBoard).
func NewMonsterBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "Monster", Slug: "monster", URL: "https://www.monster.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.monster.com/jobs/search", "monster.com/job-openings/", SearchParams{
		KeywordParam: "q", LocationParam: "where",
	})
}
