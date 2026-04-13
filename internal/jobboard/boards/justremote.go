package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewJustRemoteBoard creates a JustRemote board (ScrapeBoard).
func NewJustRemoteBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://justremote.co",
		Name:     "JustRemote",
		Slug:     "justremote",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://justremote.co/remote-jobs", "/remote-jobs/")
}
