package db

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/model"
)

// AddToPipeline inserts a new URL into the processing pipeline.
func (r *sqliteRepo) AddToPipeline(ctx context.Context, url, source string) error {
	return r.d.InsertPipelineEntry(ctx, url, source)
}

// ListPipeline returns pipeline entries, optionally filtered by status.
func (r *sqliteRepo) ListPipeline(ctx context.Context, status string) ([]RepoPipelineEntry, error) {
	var entries []PipelineEntry
	var err error

	if status == "" {
		entries, err = r.d.ListPipelineEntries(ctx)
	} else {
		entries, err = r.d.ListPipelineEntriesByStatus(ctx, status)
	}
	if err != nil {
		return nil, err
	}

	return lo.Map(entries, func(e PipelineEntry, _ int) RepoPipelineEntry {
		return RepoPipelineEntry{
			URL:       e.URL,
			Source:    e.Source,
			Status:    e.Status,
			CreatedAt: parseTimeOrZero(e.AddedAt),
			UpdatedAt: parseTimeOrZero(e.ProcessedAt),
		}
	}), nil
}

// UpdatePipelineStatus transitions a pipeline entry to a new status.
func (r *sqliteRepo) UpdatePipelineStatus(ctx context.Context, url, status string) error {
	return r.d.UpdatePipelineEntryStatus(ctx, url, status)
}

// RecordScan saves a scan-history entry for deduplication.
func (r *sqliteRepo) RecordScan(ctx context.Context, url, portal, title, company string) error {
	return r.d.InsertScanRecord(ctx, url, portal, title, company)
}

// HasBeenScanned checks whether a URL+portal combination was already scanned.
func (r *sqliteRepo) HasBeenScanned(ctx context.Context, url, portal string) (bool, error) {
	return r.d.HasBeenScanned(ctx, url, portal)
}

// ListScanHistory returns all scan-history records ordered by scan time.
func (r *sqliteRepo) ListScanHistory(ctx context.Context) ([]RepoScanRecord, error) {
	dbRecs, err := r.d.ListScanRecords(ctx)
	if err != nil {
		return nil, err
	}
	return lo.Map(dbRecs, func(rec ScanRecord, _ int) RepoScanRecord {
		return RepoScanRecord{
			URL:       rec.URL,
			Portal:    rec.Portal,
			Title:     rec.Title,
			Company:   rec.Company,
			ScannedAt: parseTimeOrZero(rec.ScannedAt),
		}
	}), nil
}

// ComputeMetrics calculates aggregate pipeline statistics.
func (r *sqliteRepo) ComputeMetrics(ctx context.Context) (model.PipelineMetrics, error) {
	return r.d.ComputeMetrics(ctx)
}

// parseTimeOrZero parses a datetime string, returning zero time on failure.
func parseTimeOrZero(s string) time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
