package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/samber/oops"
	"github.com/vingarcia/ksql"
)

// UserProfile maps to the user_profile table.
type UserProfile struct {
	FullName        string `ksql:"full_name"`
	Email           string `ksql:"email"`
	Phone           string `ksql:"phone"`
	Location        string `ksql:"location"`
	Timezone        string `ksql:"timezone"`
	LinkedInURL     string `ksql:"linkedin_url"`
	GitHubURL       string `ksql:"github_url"`
	PortfolioURL    string `ksql:"portfolio_url"`
	TwitterURL      string `ksql:"twitter_url"`
	Headline        string `ksql:"headline"`
	ExitStory       string `ksql:"exit_story"`
	Superpowers     string `ksql:"superpowers"`       // JSON array
	ProofPoints     string `ksql:"proof_points"`      // JSON array
	TargetRoles     string `ksql:"target_roles"`      // JSON array
	Archetypes      string `ksql:"archetypes"`        // JSON array
	CompTarget      string `ksql:"comp_target"`
	CompCurrency    string `ksql:"comp_currency"`
	CompMinimum     string `ksql:"comp_minimum"`
	CompLocationFlex string `ksql:"comp_location_flex"`
	Country         string `ksql:"country"`
	City            string `ksql:"city"`
	VisaStatus      string `ksql:"visa_status"`
	DealBreakers    string `ksql:"deal_breakers"`     // JSON array
	Preferences     string `ksql:"preferences"`       // JSON object
	CreatedAt       string `ksql:"created_at"`
	UpdatedAt       string `ksql:"updated_at"`
	ksql.Table      `ksql:"user_profile"`
	ID              int    `ksql:"id"`
}

// ProofPoint represents a single proof point entry.
type ProofPoint struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	HeroMetric string `json:"hero_metric"`
}

// Archetype represents a target role archetype.
type Archetype struct {
	Name  string `json:"name"`
	Level string `json:"level"`
	Fit   string `json:"fit"` // primary, secondary, adjacent
}

// GetSuperpowers returns the superpowers as a string slice.
func (p *UserProfile) GetSuperpowers() []string {
	return parseJSONStringArray(p.Superpowers)
}

// SetSuperpowers serializes a string slice into the superpowers field.
func (p *UserProfile) SetSuperpowers(s []string) {
	p.Superpowers = toJSONStringArray(s)
}

// GetTargetRoles returns the target roles as a string slice.
func (p *UserProfile) GetTargetRoles() []string {
	return parseJSONStringArray(p.TargetRoles)
}

// SetTargetRoles serializes a string slice into the target_roles field.
func (p *UserProfile) SetTargetRoles(s []string) {
	p.TargetRoles = toJSONStringArray(s)
}

// GetDealBreakers returns the deal breakers as a string slice.
func (p *UserProfile) GetDealBreakers() []string {
	return parseJSONStringArray(p.DealBreakers)
}

// SetDealBreakers serializes a string slice into the deal_breakers field.
func (p *UserProfile) SetDealBreakers(s []string) {
	p.DealBreakers = toJSONStringArray(s)
}

// GetProofPoints returns the proof points parsed from JSON.
func (p *UserProfile) GetProofPoints() []ProofPoint {
	var points []ProofPoint
	if err := json.Unmarshal([]byte(p.ProofPoints), &points); err != nil {
		return nil
	}
	return points
}

// SetProofPoints serializes proof points to JSON.
func (p *UserProfile) SetProofPoints(points []ProofPoint) {
	data, err := json.Marshal(points)
	if err != nil {
		p.ProofPoints = "[]"
		return
	}
	p.ProofPoints = string(data)
}

// GetArchetypes returns the archetypes parsed from JSON.
func (p *UserProfile) GetArchetypes() []Archetype {
	var arcs []Archetype
	if err := json.Unmarshal([]byte(p.Archetypes), &arcs); err != nil {
		return nil
	}
	return arcs
}

// SetArchetypes serializes archetypes to JSON.
func (p *UserProfile) SetArchetypes(arcs []Archetype) {
	data, err := json.Marshal(arcs)
	if err != nil {
		p.Archetypes = "[]"
		return
	}
	p.Archetypes = string(data)
}

// GetProfile returns the user profile, creating a default one if none exists.
func (d *DB) GetProfile(ctx context.Context) (*UserProfile, error) {
	var p UserProfile
	err := d.ksql.QueryOne(ctx, &p, `
		SELECT id, full_name, email, phone, location, timezone,
		       linkedin_url, github_url, portfolio_url, twitter_url,
		       headline, exit_story, superpowers, proof_points,
		       target_roles, archetypes, comp_target, comp_currency,
		       comp_minimum, comp_location_flex, country, city,
		       visa_status, deal_breakers, preferences,
		       created_at, updated_at
		FROM user_profile LIMIT 1`)
	if errors.Is(err, ksql.ErrRecordNotFound) {
		// No profile yet -- return empty default.
		return &UserProfile{
			Superpowers:  "[]",
			ProofPoints:  "[]",
			TargetRoles:  "[]",
			Archetypes:   "[]",
			DealBreakers: "[]",
			Preferences:  "{}",
			CompCurrency: "USD",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SaveProfile upserts the user profile (only one row allowed).
func (d *DB) SaveProfile(ctx context.Context, p *UserProfile) error {
	if p.ID == 0 {
		// Insert new profile.
		_, err := d.sql.ExecContext(ctx, `
			INSERT INTO user_profile (
				full_name, email, phone, location, timezone,
				linkedin_url, github_url, portfolio_url, twitter_url,
				headline, exit_story, superpowers, proof_points,
				target_roles, archetypes, comp_target, comp_currency,
				comp_minimum, comp_location_flex, country, city,
				visa_status, deal_breakers, preferences
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.FullName, p.Email, p.Phone, p.Location, p.Timezone,
			p.LinkedInURL, p.GitHubURL, p.PortfolioURL, p.TwitterURL,
			p.Headline, p.ExitStory, p.Superpowers, p.ProofPoints,
			p.TargetRoles, p.Archetypes, p.CompTarget, p.CompCurrency,
			p.CompMinimum, p.CompLocationFlex, p.Country, p.City,
			p.VisaStatus, p.DealBreakers, p.Preferences,
		)
		return oops.Wrapf(err, "inserting user profile")
	}

	// Update existing profile.
	_, err := d.sql.ExecContext(ctx, `
		UPDATE user_profile SET
			full_name = ?, email = ?, phone = ?, location = ?, timezone = ?,
			linkedin_url = ?, github_url = ?, portfolio_url = ?, twitter_url = ?,
			headline = ?, exit_story = ?, superpowers = ?, proof_points = ?,
			target_roles = ?, archetypes = ?, comp_target = ?, comp_currency = ?,
			comp_minimum = ?, comp_location_flex = ?, country = ?, city = ?,
			visa_status = ?, deal_breakers = ?, preferences = ?,
			updated_at = datetime('now')
		WHERE id = ?`,
		p.FullName, p.Email, p.Phone, p.Location, p.Timezone,
		p.LinkedInURL, p.GitHubURL, p.PortfolioURL, p.TwitterURL,
		p.Headline, p.ExitStory, p.Superpowers, p.ProofPoints,
		p.TargetRoles, p.Archetypes, p.CompTarget, p.CompCurrency,
		p.CompMinimum, p.CompLocationFlex, p.Country, p.City,
		p.VisaStatus, p.DealBreakers, p.Preferences,
		p.ID,
	)
	return oops.Wrapf(err, "updating user profile id=%d", p.ID)
}

// profileFieldQueries maps each updatable field name to its prepared UPDATE query.
// Using pre-built queries avoids SQL string concatenation (gosec G202).
var profileFieldQueries = func() map[string]string {
	fields := []string{
		"full_name", "email", "phone", "location", "timezone",
		"linkedin_url", "github_url", "portfolio_url", "twitter_url",
		"headline", "exit_story", "superpowers", "proof_points",
		"target_roles", "archetypes", "comp_target", "comp_currency",
		"comp_minimum", "comp_location_flex", "country", "city",
		"visa_status", "deal_breakers", "preferences",
	}
	m := make(map[string]string, len(fields))
	for _, f := range fields {
		m[f] = "UPDATE user_profile SET " + f +
			" = ?, updated_at = datetime('now') WHERE id = (SELECT id FROM user_profile LIMIT 1)"
	}
	return m
}()

// UpdateProfileField updates a single field on the profile by name.
// Useful for incremental AI enrichment.
func (d *DB) UpdateProfileField(ctx context.Context, field, value string) error {
	query, ok := profileFieldQueries[field]
	if !ok {
		return oops.Errorf("field %q is not updatable", field)
	}

	_, err := d.sql.ExecContext(ctx, query, value)
	return oops.Wrapf(err, "updating profile field %s", field)
}

// parseJSONStringArray parses a JSON array of strings.
func parseJSONStringArray(s string) []string {
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil
	}
	return result
}

// toJSONStringArray serializes a string slice to JSON.
func toJSONStringArray(s []string) string {
	data, err := json.Marshal(s)
	if err != nil {
		return "[]"
	}
	return string(data)
}
