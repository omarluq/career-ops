package boards

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// --- Jooble ---

// JoobleBoard searches jobs on jooble.org, a meta-aggregator.
type JoobleBoard struct {
	BaseBoard
}

// NewJoobleBoard creates a Jooble board instance.
func NewJoobleBoard() *JoobleBoard {
	return &JoobleBoard{
		BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
			Name: "Jooble", Slug: "jooble", URL: "https://jooble.org",
			Category: jobboard.CategoryAggregator, AuthType: jobboard.AuthNone,
			RateLimit: jobboard.RateConfig{RequestsPerMinute: 15, BurstSize: 3, CooldownOnError: 20 * time.Second},
			Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
		}},
	}
}

// Search discovers job listings on Jooble.
func (b *JoobleBoard) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	params := url.Values{"ukw": {strings.Join(q.Keywords, " ")}}
	if q.Location != "" {
		params.Set("ukw", strings.Join(q.Keywords, " ")+" "+q.Location)
	}

	searchURL := "https://jooble.org/SearchResult?" + params.Encode()

	return scrapeJobLinks(ctx, searchURL, "jooble.org/desc/", b.BMeta.Slug, defaultMax(q.MaxResults))
}
