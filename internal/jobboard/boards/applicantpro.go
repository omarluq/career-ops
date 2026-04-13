package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewApplicantProBoard creates an ApplicantPro board (ScrapeBoard).
func NewApplicantProBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.applicantpro.com",
		Name:         "ApplicantPro",
		Slug:         "applicantpro",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.applicantpro.com/jobs", "/jobs/")
}
