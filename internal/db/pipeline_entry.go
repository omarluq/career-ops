package db

import (
	"context"

	"github.com/samber/oops"
	"github.com/vingarcia/ksql"
)

// PipelineEntry maps to the pipeline_entries table.
type PipelineEntry struct {
	URL         string `ksql:"url"`
	Source      string `ksql:"source"`
	Status      string `ksql:"status"`
	AddedAt     string `ksql:"added_at"`
	ProcessedAt string `ksql:"processed_at"`
	ksql.Table  `ksql:"pipeline_entries"`
	ID          int `ksql:"id"`
}

// InsertPipelineEntry adds a URL to the pipeline, ignoring duplicates.
// Uses raw SQL for ON CONFLICT since ksql does not support it natively.
func (d *DB) InsertPipelineEntry(ctx context.Context, url, source string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO pipeline_entries (url, source)
		VALUES (?, ?)
		ON CONFLICT(url) DO NOTHING`, url, source)

	return oops.Wrapf(err, "inserting pipeline entry url=%s", url)
}

// ListPipelineEntries returns all pipeline entries ordered by added_at descending.
func (d *DB) ListPipelineEntries(ctx context.Context) ([]PipelineEntry, error) {
	var entries []PipelineEntry
	err := d.ksql.Query(ctx, &entries, `
		SELECT id, url, source, status, added_at, processed_at
		FROM pipeline_entries
		ORDER BY added_at DESC`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing pipeline entries")
	}
	return entries, nil
}

// ListPipelineEntriesByStatus returns pipeline entries filtered by status.
func (d *DB) ListPipelineEntriesByStatus(ctx context.Context, status string) ([]PipelineEntry, error) {
	var entries []PipelineEntry
	err := d.ksql.Query(ctx, &entries, `
		SELECT id, url, source, status, added_at, processed_at
		FROM pipeline_entries
		WHERE status = ?
		ORDER BY added_at DESC`, status)
	if err != nil {
		return nil, oops.Wrapf(err, "listing pipeline entries with status=%s", status)
	}
	return entries, nil
}

// UpdatePipelineEntryStatus transitions a pipeline entry to a new status.
func (d *DB) UpdatePipelineEntryStatus(ctx context.Context, url, status string) error {
	res, err := d.sql.ExecContext(ctx, `
		UPDATE pipeline_entries SET status = ?, processed_at = datetime('now')
		WHERE url = ?`, status, url)
	if err != nil {
		return oops.Wrapf(err, "updating pipeline entry status url=%s", url)
	}

	n, rowErr := res.RowsAffected()
	if rowErr != nil {
		return oops.Wrapf(rowErr, "checking rows affected for pipeline entry url=%s", url)
	}
	if n == 0 {
		return oops.Errorf("pipeline entry url=%s not found", url)
	}

	return nil
}
