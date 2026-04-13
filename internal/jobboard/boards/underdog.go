package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewUnderdogBoard creates an Underdog board (ScrapeBoard).
func NewUnderdogBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://underdog.io",
		Name:     "Underdog",
		Slug:     "underdog",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://underdog.io/jobs", "/job/")
}
