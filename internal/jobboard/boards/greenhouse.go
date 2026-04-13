package boards

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// GreenhouseBoard integrates with Greenhouse's public boards API.
type GreenhouseBoard struct {
	companySlug string
	BaseBoard
}

// NewGreenhouseBoard creates a Greenhouse board configured for API access.
func NewGreenhouseBoard() *GreenhouseBoard {
	return &GreenhouseBoard{
		BaseBoard: BaseBoard{BMeta: jobboard.BoardMeta{
			URL:          "https://boards-api.greenhouse.io",
			Name:         "Greenhouse",
			Slug:         "greenhouse",
			Category:     jobboard.CategoryATS,
			AuthType:     jobboard.AuthNone,
			RateLimit:    jobboard.RateConfig{RequestsPerMinute: 30, BurstSize: 5, CooldownOnError: 10 * time.Second},
			Capabilities: []jobboard.Capability{jobboard.CapAPI, jobboard.CapSearch, jobboard.CapApply},
		}},
	}
}

// greenhouseJob is the JSON shape returned by the Greenhouse boards API.
type greenhouseJob struct {
	Title     string                `json:"title"`
	AbsURL    string                `json:"absolute_url"`
	Location  struct{ Name string } `json:"location"`
	UpdatedAt string                `json:"updated_at"`
}

// Search queries the Greenhouse public boards API for job listings.
func (b *GreenhouseBoard) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	apiURL := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs", b.companySlug)

	body, err := httpGet(ctx, apiURL)
	if err != nil {
		return nil, oops.In("boards").Tags("greenhouse", "search").Wrapf(err, "fetching greenhouse jobs")
	}

	var payload struct {
		Jobs []greenhouseJob `json:"jobs"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, oops.In("boards").Tags("greenhouse", "parse").Wrapf(err, "parsing greenhouse response")
	}

	needle := strings.ToLower(strings.Join(q.Keywords, " "))
	limit := defaultMax(q.MaxResults)

	filtered := lo.Filter(payload.Jobs, func(j greenhouseJob, _ int) bool {
		return needle == "" || strings.Contains(strings.ToLower(j.Title), needle)
	})

	if limit < 0 {
		limit = 0
	}

	capped := lo.Subset(filtered, 0, uint(limit))

	return lo.Map(capped, func(j greenhouseJob, _ int) jobboard.SearchResult {
		var posted time.Time
		if p, err := time.Parse(time.RFC3339, j.UpdatedAt); err == nil {
			posted = p
		}

		return jobboard.SearchResult{
			URL:      j.AbsURL,
			Title:    j.Title,
			Location: j.Location.Name,
			Board:    b.BMeta.Slug,
			PostedAt: posted,
		}
	}), nil
}

// Apply stubs application submission for Greenhouse.
func (b *GreenhouseBoard) Apply(_ context.Context, _ jobboard.Application) (jobboard.ApplyResult, error) {
	return jobboard.ApplyResult{
		Status:      jobboard.ApplySkipped,
		SubmittedAt: time.Now(),
	}, nil
}
