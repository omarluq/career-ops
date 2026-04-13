package boards

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// RemotiveBoard searches Remotive via its public JSON API.
type RemotiveBoard struct{ BaseBoard }

// NewRemotiveBoard constructs a Remotive board instance.
func NewRemotiveBoard() *RemotiveBoard {
	return &RemotiveBoard{BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
		URL:      "https://remotive.com",
		Name:     "Remotive",
		Slug:     "remotive",
		Category: jobboard.CategoryNiche,
		AuthType: jobboard.AuthNone,
		RateLimit: jobboard.RateConfig{
			RequestsPerMinute: 10,
			BurstSize:         2,
			CooldownOnError:   30 * time.Second,
		},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapAPI},
	}}}
}

// Search queries the Remotive API for matching jobs.
func (b *RemotiveBoard) Search(ctx context.Context, query jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	url := "https://remotive.com/api/remote-jobs"
	if len(query.Keywords) > 0 {
		url += "?search=" + strings.Join(query.Keywords, "+")
	}

	body, err := httpGet(ctx, url)
	if err != nil {
		return nil, oops.In("remotive").Wrapf(err, "fetching API")
	}

	var resp struct {
		Jobs []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			CompanyName string `json:"company_name"`
			Date        string `json:"publication_date"`
		} `json:"jobs"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, oops.In("remotive").Wrapf(err, "parsing API response")
	}

	limit := defaultMax(query.MaxResults)
	results := make([]jobboard.SearchResult, 0, limit)

	for i, j := range resp.Jobs {
		if i >= limit {
			break
		}

		parsed, err := time.Parse("2006-01-02T15:04:05", j.Date)

		result := jobboard.SearchResult{
			URL:     j.URL,
			Title:   j.Title,
			Company: j.CompanyName,
			Board:   b.BMeta.Slug,
			Remote:  true,
		}
		if err == nil {
			result.PostedAt = parsed
		}

		results = append(results, result)
	}

	return results, nil
}


// compile-time interface assertion.
var _ jobboard.Board = (*RemotiveBoard)(nil)
