package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewFreelancerBoard creates a Freelancer board (ScrapeBoard).
func NewFreelancerBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.freelancer.com",
		Name:         "Freelancer",
		Slug:         "freelancer",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthSession,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.freelancer.com/jobs/", "/projects/")
}
