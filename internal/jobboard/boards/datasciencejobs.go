package boards

import (
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// NewDataScienceJobsBoard creates a DataScienceJobs board (ScrapeBoard).
func NewDataScienceJobsBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:      "https://datasciencejobs.com",
		Name:     "DataScienceJobs",
		Slug:     "datasciencejobs",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://datasciencejobs.com/jobs", "/job/")
}
