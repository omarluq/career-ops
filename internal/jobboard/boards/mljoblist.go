package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewMLJobListBoard creates an ML Job List board (ScrapeBoard).
func NewMLJobListBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://mljoblist.com",
		Name:     "ML Job List",
		Slug:     "mljoblist",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://mljoblist.com/jobs", "/job/")
}
