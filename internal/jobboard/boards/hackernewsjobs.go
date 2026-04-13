package boards

import (
	"context"
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// HackerNewsJobsBoard scrapes Hacker News jobs for listings.
type HackerNewsJobsBoard struct{ BaseBoard }

// NewHackerNewsJobsBoard constructs a Hacker News Jobs board instance.
func NewHackerNewsJobsBoard() *HackerNewsJobsBoard {
	return &HackerNewsJobsBoard{BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
		URL:      "https://news.ycombinator.com/jobs",
		Name:     "Hacker News Jobs",
		Slug:     "hackernews-jobs",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 5,
			BurstSize:         1,
			CooldownOnError:   60 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}}}
}

// Search scrapes Hacker News jobs page for listings matching the query.
func (b *HackerNewsJobsBoard) Search(ctx context.Context, query jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	url := "https://news.ycombinator.com/jobs"

	return scrapeJobLinks(ctx, url, "item?id=", b.BMeta.Slug, defaultMax(query.MaxResults))
}

// compile-time interface assertion.
var _ jobboard.Board = (*HackerNewsJobsBoard)(nil)
