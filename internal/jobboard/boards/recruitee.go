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

// RecruiteeBoard integrates with Recruitee via their public API.
type RecruiteeBoard struct{ BaseBoard }

// NewRecruiteeBoard creates a Recruitee board instance.
func NewRecruiteeBoard() *RecruiteeBoard {
	return &RecruiteeBoard{BaseBoard{BMeta: jobboard.BoardMeta{
		URL:          "https://www.recruitee.com",
		Name:         "Recruitee",
		Slug:         "recruitee",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3, CooldownOnError: 15 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapAPI},
	}}}
}

// recruiteeOffer is the JSON shape returned by the Recruitee public API.
type recruiteeOffer struct {
	Title       string `json:"title"`
	CareersURL  string `json:"careers_url"`
	City        string `json:"city"`
	Department  string `json:"department"`
	CreatedAt   string `json:"created_at"`
	CompanyName string `json:"company_name"`
}

// Search queries the Recruitee public API for job listings.
func (b *RecruiteeBoard) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	apiURL := fmt.Sprintf("https://api.recruitee.com/c/company/offers?q=%s", strings.Join(q.Keywords, "+"))

	body, err := httpGet(ctx, apiURL)
	if err != nil {
		return nil, oops.In("boards").Tags("recruitee", "search").Wrapf(err, "fetching recruitee offers")
	}

	var payload struct {
		Offers []recruiteeOffer `json:"offers"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, oops.In("boards").Tags("recruitee", "parse").Wrapf(err, "parsing recruitee response")
	}

	limit := defaultMax(q.MaxResults)
	if limit < 0 {
		limit = 0
	}

	capped := lo.Subset(payload.Offers, 0, uint(limit))

	return lo.Map(capped, func(o recruiteeOffer, _ int) jobboard.SearchResult {
		var posted time.Time
		if p, err := time.Parse(time.RFC3339, o.CreatedAt); err == nil {
			posted = p
		}

		return jobboard.SearchResult{
			URL:      o.CareersURL,
			Title:    o.Title,
			Company:  o.CompanyName,
			Location: o.City,
			Board:    b.BMeta.Slug,
			PostedAt: posted,
		}
	}), nil
}

