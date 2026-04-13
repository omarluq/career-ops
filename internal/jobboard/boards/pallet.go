package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewPalletBoard creates a Pallet board (ScrapeBoard).
func NewPalletBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://pallet.com",
		Name:     "Pallet",
		Slug:     "pallet",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://pallet.com/jobs", "/jobs/")
}
