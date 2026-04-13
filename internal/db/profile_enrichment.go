package db

import (
	"context"

	"github.com/samber/oops"
	"github.com/vingarcia/ksql"
)

// ProfileEnrichment maps to the profile_enrichments table.
// Tracks sources used to improve the user profile over time.
type ProfileEnrichment struct {
	SourceType    string `ksql:"source_type"`    // github, linkedin, blog, conversation, article, portfolio
	SourceURL     string `ksql:"source_url"`
	SourceTitle   string `ksql:"source_title"`
	ExtractedData string `ksql:"extracted_data"` // JSON: what was extracted
	AppliedFields string `ksql:"applied_fields"` // JSON array: which profile fields were updated
	Confidence    string `ksql:"confidence"`     // low, medium, high
	AppliedAt     string `ksql:"applied_at"`
	CreatedAt     string `ksql:"created_at"`
	ksql.Table    `ksql:"profile_enrichments"`
	ID            int  `ksql:"id"`
	Applied       bool `ksql:"applied"`
}

// InsertEnrichment records a new enrichment source.
func (d *DB) InsertEnrichment(ctx context.Context, e *ProfileEnrichment) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO profile_enrichments (
			source_type, source_url, source_title,
			extracted_data, applied_fields, confidence, applied
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.SourceType, e.SourceURL, e.SourceTitle,
		e.ExtractedData, e.AppliedFields, e.Confidence, e.Applied,
	)
	return oops.Wrapf(err, "inserting enrichment from %s", e.SourceType)
}

// ListEnrichments returns all enrichment records ordered by creation time.
func (d *DB) ListEnrichments(ctx context.Context) ([]ProfileEnrichment, error) {
	var enrichments []ProfileEnrichment
	err := d.ksql.Query(ctx, &enrichments, `
		SELECT id, source_type, source_url, source_title,
		       extracted_data, applied_fields, confidence,
		       applied, applied_at, created_at
		FROM profile_enrichments
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing enrichments")
	}
	return enrichments, nil
}

// ListPendingEnrichments returns enrichments not yet applied to the profile.
func (d *DB) ListPendingEnrichments(ctx context.Context) ([]ProfileEnrichment, error) {
	var enrichments []ProfileEnrichment
	err := d.ksql.Query(ctx, &enrichments, `
		SELECT id, source_type, source_url, source_title,
		       extracted_data, applied_fields, confidence,
		       applied, applied_at, created_at
		FROM profile_enrichments
		WHERE applied = 0
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing pending enrichments")
	}
	return enrichments, nil
}

// MarkEnrichmentApplied marks an enrichment as applied with the affected fields.
func (d *DB) MarkEnrichmentApplied(ctx context.Context, id int, appliedFields string) error {
	_, err := d.sql.ExecContext(ctx, `
		UPDATE profile_enrichments
		SET applied = 1, applied_at = datetime('now'), applied_fields = ?
		WHERE id = ?`, appliedFields, id)
	return oops.Wrapf(err, "marking enrichment %d as applied", id)
}

// CountEnrichmentsBySource returns the number of enrichments per source type.
func (d *DB) CountEnrichmentsBySource(ctx context.Context) (map[string]int, error) {
	type sourceCount struct {
		SourceType string `ksql:"source_type"`
		Count      int    `ksql:"count"`
	}
	var counts []sourceCount
	err := d.ksql.Query(ctx, &counts, `
		SELECT source_type, COUNT(*) as count
		FROM profile_enrichments
		GROUP BY source_type`)
	if err != nil {
		return nil, oops.Wrapf(err, "counting enrichments by source")
	}

	result := make(map[string]int, len(counts))
	for _, c := range counts {
		result[c.SourceType] = c.Count
	}
	return result, nil
}
