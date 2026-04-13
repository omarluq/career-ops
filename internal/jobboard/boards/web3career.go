package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewWeb3CareerBoard creates a Web3.Career board (ScrapeBoard).
func NewWeb3CareerBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://web3.career",
		Name:     "Web3.Career",
		Slug:     "web3career",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://web3.career", "/job/")
}
