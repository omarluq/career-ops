package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewBuiltInBoard creates a BuiltIn board (remote-aware ScrapeBoard).
func NewBuiltInBoard() *ScrapeBoard {
	return NewRemoteScrapeBoard(jobboard.BoardMeta{
		URL:      "https://builtin.com/jobs",
		Name:     "BuiltIn",
		Slug:     "builtin",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://builtin.com/jobs", "/job/")
}
