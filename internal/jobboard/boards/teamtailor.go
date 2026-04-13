package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewTeamTailorBoard creates a TeamTailor board (ScrapeBoard).
func NewTeamTailorBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.teamtailor.com",
		Name:         "TeamTailor",
		Slug:         "teamtailor",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.teamtailor.com", "/jobs/")
}
