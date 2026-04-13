package db

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/model"
)

// GetProfile returns the user profile, creating a default if none exists.
func (r *sqliteRepo) GetProfile(ctx context.Context) (*model.UserProfile, error) {
	dbProf, err := r.d.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	return dbProfileToModel(dbProf), nil
}

// SaveProfile upserts the user profile.
func (r *sqliteRepo) SaveProfile(ctx context.Context, p *model.UserProfile) error {
	dbProf := modelToDBProfile(p)
	return r.d.SaveProfile(ctx, dbProf)
}

// UpdateProfileField updates a single field on the profile by name.
func (r *sqliteRepo) UpdateProfileField(ctx context.Context, field, value string) error {
	return r.d.UpdateProfileField(ctx, field, value)
}

// RecordEnrichment stores a profile enrichment event.
func (r *sqliteRepo) RecordEnrichment(ctx context.Context, e *model.ProfileEnrichment) error {
	dbE := &ProfileEnrichment{
		SourceType:    e.SourceType,
		SourceURL:     e.SourceURL,
		SourceTitle:   e.SourceTitle,
		ExtractedData: e.ExtractedData,
		AppliedFields: e.AppliedFields,
		Confidence:    e.Confidence,
		Applied:       e.Applied,
	}
	return r.d.InsertEnrichment(ctx, dbE)
}

// ListPendingEnrichments returns enrichments not yet applied.
func (r *sqliteRepo) ListPendingEnrichments(ctx context.Context) ([]model.ProfileEnrichment, error) {
	dbEnrichments, err := r.d.ListPendingEnrichments(ctx)
	if err != nil {
		return nil, err
	}
	return lo.Map(dbEnrichments, func(e ProfileEnrichment, _ int) model.ProfileEnrichment {
		return model.ProfileEnrichment{
			ID:            e.ID,
			SourceType:    e.SourceType,
			SourceURL:     e.SourceURL,
			SourceTitle:   e.SourceTitle,
			ExtractedData: e.ExtractedData,
			AppliedFields: e.AppliedFields,
			Confidence:    e.Confidence,
			Applied:       e.Applied,
			AppliedAt:     e.AppliedAt,
			CreatedAt:     e.CreatedAt,
		}
	}), nil
}

// ApplyEnrichment marks an enrichment as applied.
func (r *sqliteRepo) ApplyEnrichment(ctx context.Context, id int, appliedFields string) error {
	return r.d.MarkEnrichmentApplied(ctx, id, appliedFields)
}

// dbProfileToModel converts a db.UserProfile to model.UserProfile.
func dbProfileToModel(p *UserProfile) *model.UserProfile {
	return &model.UserProfile{
		ID:               p.ID,
		FullName:         p.FullName,
		Email:            p.Email,
		Phone:            p.Phone,
		Location:         p.Location,
		Timezone:         p.Timezone,
		LinkedInURL:      p.LinkedInURL,
		GitHubURL:        p.GitHubURL,
		PortfolioURL:     p.PortfolioURL,
		TwitterURL:       p.TwitterURL,
		Headline:         p.Headline,
		ExitStory:        p.ExitStory,
		Superpowers:      p.GetSuperpowers(),
		ProofPoints:      dbProofPointsToModel(p.GetProofPoints()),
		TargetRoles:      p.GetTargetRoles(),
		Archetypes:       dbArchetypesToModel(p.GetArchetypes()),
		CompTarget:       p.CompTarget,
		CompCurrency:     p.CompCurrency,
		CompMinimum:      p.CompMinimum,
		CompLocationFlex: p.CompLocationFlex,
		Country:          p.Country,
		City:             p.City,
		VisaStatus:       p.VisaStatus,
		DealBreakers:     p.GetDealBreakers(),
		Preferences:      parseJSONMap(p.Preferences),
	}
}

// modelToDBProfile converts a model.UserProfile to db.UserProfile.
func modelToDBProfile(p *model.UserProfile) *UserProfile {
	dbProf := &UserProfile{
		ID:               p.ID,
		FullName:         p.FullName,
		Email:            p.Email,
		Phone:            p.Phone,
		Location:         p.Location,
		Timezone:         p.Timezone,
		LinkedInURL:      p.LinkedInURL,
		GitHubURL:        p.GitHubURL,
		PortfolioURL:     p.PortfolioURL,
		TwitterURL:       p.TwitterURL,
		Headline:         p.Headline,
		ExitStory:        p.ExitStory,
		CompTarget:       p.CompTarget,
		CompCurrency:     p.CompCurrency,
		CompMinimum:      p.CompMinimum,
		CompLocationFlex: p.CompLocationFlex,
		Country:          p.Country,
		City:             p.City,
		VisaStatus:       p.VisaStatus,
	}
	dbProf.SetSuperpowers(p.Superpowers)
	dbProf.SetProofPoints(modelProofPointsToDB(p.ProofPoints))
	dbProf.SetTargetRoles(p.TargetRoles)
	dbProf.SetArchetypes(modelArchetypesToDB(p.Archetypes))
	dbProf.SetDealBreakers(p.DealBreakers)
	dbProf.Preferences = toJSONMap(p.Preferences)
	return dbProf
}

func dbProofPointsToModel(points []ProofPoint) []model.ProofPoint {
	return lo.Map(points, func(p ProofPoint, _ int) model.ProofPoint {
		return model.ProofPoint{Name: p.Name, URL: p.URL, HeroMetric: p.HeroMetric}
	})
}

func modelProofPointsToDB(points []model.ProofPoint) []ProofPoint {
	return lo.Map(points, func(p model.ProofPoint, _ int) ProofPoint {
		return ProofPoint{Name: p.Name, URL: p.URL, HeroMetric: p.HeroMetric}
	})
}

func dbArchetypesToModel(arcs []Archetype) []model.ArchetypeEntry {
	return lo.Map(arcs, func(a Archetype, _ int) model.ArchetypeEntry {
		return model.ArchetypeEntry{Name: a.Name, Level: a.Level, Fit: a.Fit}
	})
}

func modelArchetypesToDB(arcs []model.ArchetypeEntry) []Archetype {
	return lo.Map(arcs, func(a model.ArchetypeEntry, _ int) Archetype {
		return Archetype{Name: a.Name, Level: a.Level, Fit: a.Fit}
	})
}

func parseJSONMap(s string) map[string]string {
	var result map[string]string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return make(map[string]string)
	}
	return result
}

func toJSONMap(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(data)
}
