package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewDailyRemoteBoard creates a DailyRemote board (ScrapeBoard).
func NewDailyRemoteBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://dailyremote.com",
		Name:     "DailyRemote",
		Slug:     "dailyremote",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://dailyremote.com/remote-jobs", "/remote-job/")
}
