package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewRemoteLeafBoard creates a RemoteLeaf board (ScrapeBoard).
func NewRemoteLeafBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.remoteleaf.com",
		Name:     "RemoteLeaf",
		Slug:     "remoteleaf",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.remoteleaf.com/jobs", "/job/")
}
