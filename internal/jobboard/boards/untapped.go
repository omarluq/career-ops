package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewUntappedBoard creates an Untapped board (remote-aware ScrapeBoard).
func NewUntappedBoard() *ScrapeBoard {
	return NewRemoteScrapeBoard(jobboard.BoardMeta{
		URL:      "https://untapped.io",
		Name:     "Untapped",
		Slug:     "untapped",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://untapped.io/jobs", "/job/")
}
