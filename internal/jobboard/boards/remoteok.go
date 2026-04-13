package boards

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// RemoteOKBoard searches RemoteOK via its public JSON API.
type RemoteOKBoard struct{ BaseBoard }

// NewRemoteOKBoard constructs a RemoteOK board instance.
func NewRemoteOKBoard() *RemoteOKBoard {
	return &RemoteOKBoard{BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
		URL:      "https://remoteok.com",
		Name:     "RemoteOK",
		Slug:     "remoteok",
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

// Search queries the RemoteOK JSON API for matching jobs.
func (b *RemoteOKBoard) Search(ctx context.Context, query jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	url := "https://remoteok.com/api"
	if len(query.Keywords) > 0 {
		url += "?tag=" + strings.Join(query.Keywords, ",")
	}

	body, err := httpGet(ctx, url)
	if err != nil {
		return nil, oops.In("remoteok").Wrapf(err, "fetching API")
	}

	var raw []struct {
		URL     string `json:"url"`
		Title   string `json:"position"`
		Company string `json:"company"`
		Date    string `json:"date"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, oops.In("remoteok").Wrapf(err, "parsing API response")
	}

	limit := defaultMax(query.MaxResults)
	results := make([]jobboard.SearchResult, 0, limit)

	for i, r := range raw {
		if r.URL == "" || i >= limit {
			break
		}

		parsed, err := time.Parse(time.RFC3339, r.Date)

		result := jobboard.SearchResult{
			URL:     r.URL,
			Title:   r.Title,
			Company: r.Company,
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
var _ jobboard.Board = (*RemoteOKBoard)(nil)
