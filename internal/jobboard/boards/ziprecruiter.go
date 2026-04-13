package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewZipRecruiterBoard creates a ZipRecruiter board (ParamScrapeBoard).
func NewZipRecruiterBoard() *ParamScrapeBoard {
	return NewParamScrapeBoard(jobboard.BoardMeta{
		Name: "ZipRecruiter", Slug: "ziprecruiter", URL: "https://www.ziprecruiter.com",
		Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.ziprecruiter.com/jobs-search", "ziprecruiter.com/c/", SearchParams{
		KeywordParam: "search", LocationParam: "location",
		RemoteParam: "refine_by_location_type", RemoteValue: "only_remote",
	})
}
