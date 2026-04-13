package boards

import (
	"context"
	"time"

	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// Wellfound searches and applies on Wellfound (formerly AngelList Talent).
type Wellfound struct{ BaseBoard }

// NewWellfoundBoard returns a configured Wellfound board.
func NewWellfoundBoard() *Wellfound {
	return &Wellfound{BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
		URL:      "https://wellfound.com",
		Name:     "Wellfound",
		Slug:     "wellfound",
		Category: jobboard.CategoryStartup,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 20,
			BurstSize:         3,
			CooldownOnError:   15 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapApply, jobboard.CapScrape},
	}}}
}

// Search discovers job listings on wellfound.com.
func (b *Wellfound) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	url := buildSearchURL(b.BMeta.URL+"/jobs", q.Keywords)
	if q.Remote {
		url += remoteQueryParam
	}

	results, err := scrapeJobLinks(ctx, url, "/jobs/", b.BMeta.Slug, defaultMax(q.MaxResults))
	if err != nil {
		return nil, oops.In("boards").Tags("search", b.BMeta.Slug).Wrap(err)
	}

	return results, nil
}

// Apply submits an application on Wellfound via browser automation.
func (b *Wellfound) Apply(_ context.Context, _ jobboard.Application) (jobboard.ApplyResult, error) {
	// TODO: implement Wellfound apply via chromedp form fill
	return jobboard.ApplyResult{
		Status:       jobboard.ApplyPending,
		ErrorMessage: "wellfound apply not yet implemented",
	}, nil
}
