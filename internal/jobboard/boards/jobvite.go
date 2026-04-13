package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewJobviteBoard creates a Jobvite board (ScrapeBoard).
func NewJobviteBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://jobs.jobvite.com",
		Name:         "Jobvite",
		Slug:         "jobvite",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://jobs.jobvite.com", "/job/")
}
