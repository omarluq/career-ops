package boards

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// --- LinkedIn ---

// LinkedInBoard searches and applies to jobs on linkedin.com.
type LinkedInBoard struct {
	BaseBoard
}

// NewLinkedInBoard creates a LinkedIn board instance.
func NewLinkedInBoard() *LinkedInBoard {
	return &LinkedInBoard{
		BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
			Name: "LinkedIn", Slug: "linkedin", URL: "https://www.linkedin.com/jobs",
			Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthSession,
			RateLimit: jobboard.RateConfig{RequestsPerMinute: 10, BurstSize: 2, CooldownOnError: 30 * time.Second},
			Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapApply, jobboard.CapScrape},
		}},
	}
}

// Search discovers job listings on LinkedIn.
func (b *LinkedInBoard) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	params := url.Values{"keywords": {strings.Join(q.Keywords, " ")}}
	if q.Location != "" {
		params.Set("location", q.Location)
	}
	if q.Remote {
		params.Set("f_WT", "2")
	}

	searchURL := "https://www.linkedin.com/jobs/search/?" + params.Encode()

	return scrapeJobLinks(ctx, searchURL, "linkedin.com/jobs/view/", b.BMeta.Slug, defaultMax(q.MaxResults))
}

// Apply submits an application via LinkedIn Easy Apply (stub).
func (b *LinkedInBoard) Apply(_ context.Context, _ jobboard.Application) (jobboard.ApplyResult, error) {
	return jobboard.ApplyResult{Status: jobboard.ApplySkipped, ErrorMessage: "linkedin apply not yet implemented"}, nil
}
