package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewVentureLoopBoard creates a VentureLoop board (ScrapeBoard).
func NewVentureLoopBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://www.ventureloop.com",
		Name:     "VentureLoop",
		Slug:     "ventureloop",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.ventureloop.com/ventureloop/job_search.php", "/ventureloop/job_detail.php")
}
