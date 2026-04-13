package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewCryptoJobsListBoard creates a Crypto Jobs List board (ScrapeBoard).
func NewCryptoJobsListBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://cryptojobslist.com",
		Name:     "Crypto Jobs List",
		Slug:     "cryptojobslist",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://cryptojobslist.com/search", "/jobs/")
}
