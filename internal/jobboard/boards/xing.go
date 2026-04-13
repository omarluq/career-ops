package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewXingBoard creates a Xing board (ScrapeBoard).
func NewXingBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.xing.com/jobs",
		Name:         "Xing",
		Slug:         "xing",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.xing.com/jobs/search", "/jobs/")
}
