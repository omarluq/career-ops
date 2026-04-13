package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewTrueUpBoard creates a TrueUp board (remote-aware ScrapeBoard).
func NewTrueUpBoard() *ScrapeBoard {
	return NewRemoteScrapeBoard(jobboard.BoardMeta{
		URL:      "https://trueup.io/jobs",
		Name:     "TrueUp",
		Slug:     "trueup",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://trueup.io/jobs", "/job/")
}
