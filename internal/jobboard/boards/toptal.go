// Package boards provides concrete job board implementations for the jobboard
// package. This file contains the Toptal freelance platform.
package boards

import (
	"context"
	"strings"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// ---------------------------------------------------------------------------
// Toptal — toptal.com
// ---------------------------------------------------------------------------

// ToptalBoard implements the Board interface for Toptal.
type ToptalBoard struct{ BaseBoard }

// NewToptalBoard creates a new Toptal board.
func NewToptalBoard() *ToptalBoard {
	return &ToptalBoard{BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
		URL:          "https://www.toptal.com",
		Name:         "Toptal",
		Slug:         "toptal",
		Category:     jobboard.CategoryFreelance,
		AuthType:     jobboard.AuthSession,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}}}
}

// Search discovers freelance listings on Toptal.
func (b *ToptalBoard) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	url := "https://www.toptal.com/freelance-jobs"
	if len(q.Keywords) > 0 {
		url += "/" + strings.Join(q.Keywords, "-")
	}

	return scrapeJobLinks(ctx, url, "/freelance-jobs/", b.BMeta.Slug, defaultMax(q.MaxResults))
}
