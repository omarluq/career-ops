package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewYCombinatorBoard creates a Y Combinator board (remote-aware ScrapeBoard).
func NewYCombinatorBoard() *ScrapeBoard {
	return NewRemoteScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.workatastartup.com",
		Name:     "Y Combinator (Work at a Startup)",
		Slug:     "ycombinator",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.workatastartup.com/companies", "/jobs/")
}
