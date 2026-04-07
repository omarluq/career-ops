package repo

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/tracker"
)

// sqliteRepo implements Repository backed by a SQLite database via the db package.
// It delegates to the internal/db package for raw SQL access and handles domain
// model conversions.
type sqliteRepo struct {
	conn *sql.DB
}

// NewSQLite returns a Repository backed by the given SQLite database connection.
func NewSQLite(conn *sql.DB) Repository {
	return &sqliteRepo{conn: conn}
}

// ListApplications returns all tracked applications ordered by number.
func (r *sqliteRepo) ListApplications(ctx context.Context) ([]model.CareerApplication, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT number, date, company, role, score_raw, score, status,
		       has_pdf, report_number, report_path, notes, job_url,
		       archetype, tldr, remote, comp_estimate
		FROM applications ORDER BY number`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing applications")
	}
	return collectRows(rows, scanApplication)
}

// GetApplication returns a single application by its sequential number.
func (r *sqliteRepo) GetApplication(ctx context.Context, number int) (*model.CareerApplication, error) {
	row := r.conn.QueryRowContext(ctx, `
		SELECT number, date, company, role, score_raw, score, status,
		       has_pdf, report_number, report_path, notes, job_url,
		       archetype, tldr, remote, comp_estimate
		FROM applications WHERE number = ?`, number)

	app, err := scanApplicationFrom(row)
	if err != nil {
		return nil, oops.Wrapf(err, "getting application %d", number)
	}
	return &app, nil
}

// UpsertApplication creates or updates an application keyed by number.
func (r *sqliteRepo) UpsertApplication(ctx context.Context, app *model.CareerApplication) error {
	companyKey := tracker.NormalizeCompanyKey(app.Company)
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO applications (
			number, date, company, company_key, role, score_raw, score,
			status, has_pdf, report_number, report_path, notes, job_url,
			archetype, tldr, remote, comp_estimate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(number) DO UPDATE SET
			date=excluded.date, company=excluded.company,
			company_key=excluded.company_key, role=excluded.role,
			score_raw=excluded.score_raw, score=excluded.score,
			status=excluded.status, has_pdf=excluded.has_pdf,
			report_number=excluded.report_number,
			report_path=excluded.report_path, notes=excluded.notes,
			job_url=excluded.job_url, archetype=excluded.archetype,
			tldr=excluded.tldr, remote=excluded.remote,
			comp_estimate=excluded.comp_estimate`,
		app.Number, app.Date, app.Company, companyKey, app.Role,
		app.ScoreRaw, app.Score, app.Status, app.HasPDF,
		app.ReportNumber, app.ReportPath, app.Notes, app.JobURL,
		app.Archetype, app.TlDr, app.Remote, app.CompEstimate,
	)
	return oops.Wrapf(err, "upserting application %d", app.Number)
}

// DeleteApplication removes the application with the given number.
func (r *sqliteRepo) DeleteApplication(ctx context.Context, number int) error {
	res, err := r.conn.ExecContext(ctx,
		`DELETE FROM applications WHERE number = ?`, number)
	if err != nil {
		return oops.Wrapf(err, "deleting application %d", number)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return oops.Wrapf(err, "checking rows affected")
	}
	if n == 0 {
		return oops.Wrapf(sql.ErrNoRows, "application %d not found", number)
	}
	return nil
}

// FindDuplicates returns clusters of applications that share a company_key and
// have overlapping roles.
func (r *sqliteRepo) FindDuplicates(ctx context.Context) ([]DuplicateCluster, error) {
	apps, err := r.ListApplications(ctx)
	if err != nil {
		return nil, err
	}

	byKey := lo.GroupBy(apps, func(a model.CareerApplication) string {
		return tracker.NormalizeCompanyKey(a.Company)
	})

	return lo.FilterMap(lo.Entries(byKey), func(
		entry lo.Entry[string, []model.CareerApplication], _ int,
	) (DuplicateCluster, bool) {
		if len(entry.Value) < 2 {
			return DuplicateCluster{}, false
		}

		dupes := findRoleMatches(entry.Value)
		if len(dupes) == 0 {
			return DuplicateCluster{}, false
		}

		return DuplicateCluster{
			CompanyKey:   entry.Key,
			Applications: dupes,
		}, true
	}), nil
}

// findRoleMatches identifies applications within a company group that have
// matching roles using pairwise fuzzy comparison.
func findRoleMatches(group []model.CareerApplication) []model.CareerApplication {
	seen := make(map[int]bool)

	lo.ForEach(lo.Range(len(group)), func(i int, _ int) {
		lo.ForEach(lo.RangeFrom(i+1, len(group)-i-1), func(j int, _ int) {
			if tracker.RoleMatch(group[i].Role, group[j].Role) {
				seen[i] = true
				seen[j] = true
			}
		})
	})

	return lo.FilterMap(group, func(app model.CareerApplication, idx int) (model.CareerApplication, bool) {
		return app, seen[idx]
	})
}

// SearchApplications performs a free-text search across key fields.
func (r *sqliteRepo) SearchApplications(
	ctx context.Context, query string,
) ([]model.CareerApplication, error) {
	pattern := "%" + strings.ReplaceAll(query, "%", "%%") + "%"
	rows, err := r.conn.QueryContext(ctx, `
		SELECT number, date, company, role, score_raw, score, status,
		       has_pdf, report_number, report_path, notes, job_url,
		       archetype, tldr, remote, comp_estimate
		FROM applications
		WHERE company LIKE ? OR role LIKE ? OR notes LIKE ? OR archetype LIKE ?
		ORDER BY number`, pattern, pattern, pattern, pattern)
	if err != nil {
		return nil, oops.Wrapf(err, "searching applications")
	}
	return collectRows(rows, scanApplication)
}

// AddToPipeline inserts a new URL into the processing pipeline.
func (r *sqliteRepo) AddToPipeline(ctx context.Context, url, source string) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO pipeline (url, source, status, created_at, updated_at)
		VALUES (?, ?, 'pending', ?, ?)
		ON CONFLICT(url) DO NOTHING`,
		url, source, time.Now(), time.Now())
	return oops.Wrapf(err, "adding %s to pipeline", url)
}

// ListPipeline returns pipeline entries, optionally filtered by status.
func (r *sqliteRepo) ListPipeline(
	ctx context.Context, status string,
) ([]PipelineEntry, error) {
	query := lo.Ternary(status == "",
		`SELECT url, source, status, created_at, updated_at FROM pipeline ORDER BY created_at DESC`,
		`SELECT url, source, status, created_at, updated_at FROM pipeline WHERE status = ? ORDER BY created_at DESC`,
	)
	args := lo.Ternary[[]any](status == "", nil, []any{status})

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, oops.Wrapf(err, "listing pipeline")
	}
	return collectRows(rows, scanPipelineEntry)
}

// UpdatePipelineStatus transitions a pipeline entry to a new status.
func (r *sqliteRepo) UpdatePipelineStatus(ctx context.Context, url, status string) error {
	res, err := r.conn.ExecContext(ctx, `
		UPDATE pipeline SET status = ?, updated_at = ? WHERE url = ?`,
		status, time.Now(), url)
	if err != nil {
		return oops.Wrapf(err, "updating pipeline status for %s", url)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return oops.Wrapf(err, "checking rows affected")
	}
	if n == 0 {
		return oops.Wrapf(sql.ErrNoRows, "pipeline entry %s not found", url)
	}
	return nil
}

// RecordScan saves a scan-history entry for deduplication.
func (r *sqliteRepo) RecordScan(ctx context.Context, url, portal, title, company string) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO scan_history (url, portal, title, company, scanned_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(url, portal) DO NOTHING`,
		url, portal, title, company, time.Now())
	return oops.Wrapf(err, "recording scan for %s", url)
}

// ListScanHistory returns all scan-history records ordered by scan time.
func (r *sqliteRepo) ListScanHistory(ctx context.Context) ([]ScanRecord, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT url, portal, title, company, scanned_at
		FROM scan_history ORDER BY scanned_at DESC`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing scan history")
	}
	return collectRows(rows, scanScanRecord)
}

// HasBeenScanned checks whether a URL+portal combination was already scanned.
func (r *sqliteRepo) HasBeenScanned(ctx context.Context, url, portal string) (bool, error) {
	var count int
	err := r.conn.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM scan_history WHERE url = ? AND portal = ?`,
		url, portal).Scan(&count)
	if err != nil {
		return false, oops.Wrapf(err, "checking scan history for %s", url)
	}
	return count > 0, nil
}

// ComputeMetrics calculates aggregate pipeline statistics.
func (r *sqliteRepo) ComputeMetrics(ctx context.Context) (model.PipelineMetrics, error) {
	apps, err := r.ListApplications(ctx)
	if err != nil {
		return model.PipelineMetrics{}, err
	}
	return tracker.ComputeMetrics(apps), nil
}

// Close releases the underlying database connection.
func (r *sqliteRepo) Close() error {
	return r.conn.Close()
}

// scanner is satisfied by both *sql.Rows and *sql.Row, allowing shared scan logic.
type scanner interface {
	Scan(dest ...any) error
}

// collectRows is a generic row scanner that collects results using a scan function.
// Replaces all manual for rows.Next() { ... } loops.
func collectRows[T any](rows *sql.Rows, scan func(*sql.Rows) (T, error)) ([]T, error) {
	var result []T
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Close(); err != nil {
		return result, oops.Wrapf(err, "closing rows")
	}
	return result, oops.Wrapf(rows.Err(), "iterating rows")
}

// scanApplicationFrom scans a full application row from any scanner (Rows or Row).
func scanApplicationFrom(s scanner) (model.CareerApplication, error) {
	var app model.CareerApplication
	err := s.Scan(
		&app.Number, &app.Date, &app.Company, &app.Role,
		&app.ScoreRaw, &app.Score, &app.Status, &app.HasPDF,
		&app.ReportNumber, &app.ReportPath, &app.Notes, &app.JobURL,
		&app.Archetype, &app.TlDr, &app.Remote, &app.CompEstimate,
	)
	return app, oops.Wrapf(err, "scanning application row")
}

// scanApplication reads a full application row from a sql.Rows cursor.
func scanApplication(rows *sql.Rows) (model.CareerApplication, error) {
	return scanApplicationFrom(rows)
}

// scanPipelineEntry reads a pipeline entry from a sql.Rows cursor.
func scanPipelineEntry(rows *sql.Rows) (PipelineEntry, error) {
	var e PipelineEntry
	err := rows.Scan(&e.URL, &e.Source, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, oops.Wrapf(err, "scanning pipeline entry")
}

// scanScanRecord reads a scan record from a sql.Rows cursor.
func scanScanRecord(rows *sql.Rows) (ScanRecord, error) {
	var rec ScanRecord
	err := rows.Scan(&rec.URL, &rec.Portal, &rec.Title, &rec.Company, &rec.ScannedAt)
	return rec, oops.Wrapf(err, "scanning scan-history row")
}
