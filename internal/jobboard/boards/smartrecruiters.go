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

// SmartRecruitersBoard integrates with SmartRecruiters via their public API.
type SmartRecruitersBoard struct{ BaseBoard }

// NewSmartRecruitersBoard creates a SmartRecruiters board instance.
func NewSmartRecruitersBoard() *SmartRecruitersBoard {
	return &SmartRecruitersBoard{BaseBoard{BMeta: jobboard.BoardMeta{
		URL:          "https://jobs.smartrecruiters.com",
		Name:         "SmartRecruiters",
		Slug:         "smartrecruiters",
		Category:     jobboard.CategoryATS,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 30, BurstSize: 5, CooldownOnError: 10 * time.Second},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapAPI},
	}}}
}

// smartRecruitersJob is the JSON shape returned by the SmartRecruiters API.
type smartRecruitersJob struct {
	Name    string                    `json:"name"`
	RefURL  string                    `json:"ref"`
	Company smartRecruitersJobCompany `json:"company"`
	Loc     smartRecruitersJobLoc     `json:"location"`
	RelDate string                    `json:"releasedDate"`
}

// smartRecruitersJobCompany holds the company name from the SmartRecruiters API.
type smartRecruitersJobCompany struct {
	Name string `json:"name"`
}

// smartRecruitersJobLoc holds the location city from the SmartRecruiters API.
type smartRecruitersJobLoc struct {
	City string `json:"city"`
}

// Search queries the SmartRecruiters public API for job listings.
func (b *SmartRecruitersBoard) Search(ctx context.Context, q jobboard.SearchQuery) ([]jobboard.SearchResult, error) {
	apiURL := fmt.Sprintf("https://api.smartrecruiters.com/v1/companies/jobs?q=%s&limit=%d",
		strings.Join(q.Keywords, "+"), defaultMax(q.MaxResults))

	body, err := httpGet(ctx, apiURL)
	if err != nil {
		return nil, oops.In("boards").Tags("smartrecruiters", "search").Wrapf(err, "fetching smartrecruiters jobs")
	}

	var payload struct {
		Content []smartRecruitersJob `json:"content"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, oops.In("boards").Tags("smartrecruiters", "parse").Wrapf(err, "parsing smartrecruiters response")
	}

	return lo.Map(payload.Content, func(j smartRecruitersJob, _ int) jobboard.SearchResult {
		var posted time.Time
		if p, err := time.Parse(time.RFC3339, j.RelDate); err == nil {
			posted = p
		}

		return jobboard.SearchResult{
			URL:      j.RefURL,
			Title:    j.Name,
			Company:  j.Company.Name,
			Location: j.Loc.City,
			Board:    b.BMeta.Slug,
			PostedAt: posted,
		}
	}), nil
}

