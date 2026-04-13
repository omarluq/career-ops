package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewRemoteCoBoard creates a Remote.co board (ScrapeBoard).
func NewRemoteCoBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://remote.co/remote-jobs",
		Name:     "Remote.co",
		Slug:     "remoteco",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://remote.co/remote-jobs/search", "/remote-jobs/")
}
