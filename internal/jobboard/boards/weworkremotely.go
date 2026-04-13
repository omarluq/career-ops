package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewWeWorkRemotelyBoard creates a We Work Remotely board (ScrapeBoard).
func NewWeWorkRemotelyBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://weworkremotely.com",
		Name:     "We Work Remotely",
		Slug:     "weworkremotely",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://weworkremotely.com/remote-jobs/search", "/remote-jobs/")
}
