package boards

import (
	"context"
	"strings"
	"time"

	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// StartupJobs searches the startup.jobs board.
type StartupJobs struct{ BaseBoard }

// NewStartupJobsBoard returns a configured StartupJobs board.
func NewStartupJobsBoard() *StartupJobs {
	return &StartupJobs{BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
		URL:      "https://startup.jobs",
		Name:     "Startup Jobs",
		Slug:     "startup-jobs",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 30,
			BurstSize:         5,
			CooldownOnError:   10 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}}}
}

// Search discovers job listings on startup.jobs.
func (b *StartupJobs) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	url := b.BMeta.URL + "?q=" + strings.Join(q.Keywords, "+")
	if q.Remote {
		url += remoteQueryParam
	}

	results, err := scrapeJobLinks(ctx, url, "/jobs/", b.BMeta.Slug, defaultMax(q.MaxResults))
	if err != nil {
		return nil, oops.In("boards").Tags("search", b.BMeta.Slug).Wrap(err)
	}

	return results, nil
}
