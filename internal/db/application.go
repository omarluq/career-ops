package db

import (
	"context"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/vingarcia/ksql"

	"github.com/omarluq/career-ops/internal/model"
)

// Application maps to the applications table.
type Application struct {
	Archetype    string `ksql:"archetype"`
	ReportNumber string `ksql:"report_number"`
	UpdatedAt    string `ksql:"updated_at"`
	Date         string `ksql:"date"`
	Company      string `ksql:"company"`
	CompanyKey   string `ksql:"company_key"`
	Role         string `ksql:"role"`
	CreatedAt    string `ksql:"created_at"`
	ScoreRaw     string `ksql:"score_raw"`
	Status       string `ksql:"status"`
	CompEstimate string `ksql:"comp_estimate"`
	ReportPath   string `ksql:"report_path"`
	Remote       string `ksql:"remote"`
	JobURL       string `ksql:"job_url"`
	Notes        string `ksql:"notes"`
	TlDr         string `ksql:"tldr"`
	ksql.Table   `ksql:"applications"`
	ID           int     `ksql:"id"`
	Score        float64 `ksql:"score"`
	Number       int     `ksql:"number"`
	HasPDF       bool    `ksql:"has_pdf"`
}

// ToModel converts a database Application to the domain model.
func (a *Application) ToModel() model.CareerApplication {
	return model.CareerApplication{
		ReportPath:   a.ReportPath,
		ReportNumber: a.ReportNumber,
		Company:      a.Company,
		Role:         a.Role,
		Status:       a.Status,
		CompEstimate: a.CompEstimate,
		Date:         a.Date,
		ScoreRaw:     a.ScoreRaw,
		Remote:       a.Remote,
		TlDr:         a.TlDr,
		Notes:        a.Notes,
		JobURL:       a.JobURL,
		Archetype:    a.Archetype,
		Number:       a.Number,
		Score:        a.Score,
		HasPDF:       a.HasPDF,
	}
}

// ApplicationFromModel creates a database Application from the domain model.
func ApplicationFromModel(app *model.CareerApplication, companyKey string) Application {
	return Application{
		Number:       app.Number,
		Date:         app.Date,
		Company:      app.Company,
		CompanyKey:   companyKey,
		Role:         app.Role,
		Score:        app.Score,
		ScoreRaw:     app.ScoreRaw,
		Status:       app.Status,
		HasPDF:       app.HasPDF,
		ReportNumber: app.ReportNumber,
		ReportPath:   app.ReportPath,
		JobURL:       app.JobURL,
		Notes:        app.Notes,
		Archetype:    app.Archetype,
		TlDr:         app.TlDr,
		Remote:       app.Remote,
		CompEstimate: app.CompEstimate,
	}
}

// CompanyKeyFromName derives a lowercase, trimmed key from a company name.
func CompanyKeyFromName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ListApplications returns all applications ordered by number descending.
func (d *DB) ListApplications(ctx context.Context) ([]Application, error) {
	var apps []Application
	err := d.ksql.Query(ctx, &apps, `
		SELECT id, number, date, company, company_key, role, score, score_raw,
		       status, has_pdf, report_number, report_path, job_url, notes,
		       archetype, tldr, remote, comp_estimate, created_at, updated_at
		FROM applications
		ORDER BY number DESC`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing applications")
	}
	return apps, nil
}

// GetApplication returns a single application by ID.
func (d *DB) GetApplication(ctx context.Context, id int) (Application, error) {
	var a Application
	err := d.ksql.QueryOne(ctx, &a, `
		SELECT id, number, date, company, company_key, role, score, score_raw,
		       status, has_pdf, report_number, report_path, job_url, notes,
		       archetype, tldr, remote, comp_estimate, created_at, updated_at
		FROM applications
		WHERE id = ?`, id)
	if err != nil {
		return a, oops.Wrapf(err, "getting application id=%d", id)
	}
	return a, nil
}

// GetApplicationByNumber returns a single application by its tracker number.
func (d *DB) GetApplicationByNumber(ctx context.Context, number int) (*Application, error) {
	var a Application
	err := d.ksql.QueryOne(ctx, &a, `
		SELECT id, number, date, company, company_key, role, score, score_raw,
		       status, has_pdf, report_number, report_path, job_url, notes,
		       archetype, tldr, remote, comp_estimate, created_at, updated_at
		FROM applications
		WHERE number = ?`, number)
	if err != nil {
		return nil, oops.Wrapf(err, "getting application number=%d", number)
	}
	return &a, nil
}

// UpsertApplication inserts or updates an application keyed by number.
// Uses raw SQL for ON CONFLICT upsert since ksql does not support it natively.
func (d *DB) UpsertApplication(ctx context.Context, app *Application) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO applications (
			number, date, company, company_key, role, score, score_raw,
			status, has_pdf, report_number, report_path, job_url, notes,
			archetype, tldr, remote, comp_estimate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(number) DO UPDATE SET
			date          = excluded.date,
			company       = excluded.company,
			company_key   = excluded.company_key,
			role          = excluded.role,
			score         = excluded.score,
			score_raw     = excluded.score_raw,
			status        = excluded.status,
			has_pdf       = excluded.has_pdf,
			report_number = excluded.report_number,
			report_path   = excluded.report_path,
			job_url       = excluded.job_url,
			notes         = excluded.notes,
			archetype     = excluded.archetype,
			tldr          = excluded.tldr,
			remote        = excluded.remote,
			comp_estimate = excluded.comp_estimate,
			updated_at    = datetime('now')`,
		app.Number, app.Date, app.Company, app.CompanyKey, app.Role,
		app.Score, app.ScoreRaw, app.Status, app.HasPDF,
		app.ReportNumber, app.ReportPath, app.JobURL, app.Notes,
		app.Archetype, app.TlDr, app.Remote, app.CompEstimate,
	)

	return oops.Wrapf(err, "upserting application number=%d", app.Number)
}

// DeleteApplication removes an application by ID.
func (d *DB) DeleteApplication(ctx context.Context, id int) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM applications WHERE id = ?`, id)
	if err != nil {
		return oops.Wrapf(err, "deleting application id=%d", id)
	}

	n, rowErr := res.RowsAffected()
	if rowErr != nil {
		return oops.Wrapf(rowErr, "checking rows affected for application id=%d", id)
	}
	if n == 0 {
		return oops.Errorf("application id=%d not found", id)
	}

	return nil
}

// FindDuplicates returns groups of applications sharing the same company_key and role.
func (d *DB) FindDuplicates(ctx context.Context) ([][]Application, error) {
	var all []Application
	err := d.ksql.Query(ctx, &all, `
		SELECT id, number, date, company, company_key, role, score, score_raw,
		       status, has_pdf, report_number, report_path, job_url, notes,
		       archetype, tldr, remote, comp_estimate, created_at, updated_at
		FROM applications
		WHERE (company_key, role) IN (
			SELECT company_key, role
			FROM applications
			GROUP BY company_key, role
			HAVING COUNT(*) > 1
		)
		ORDER BY company_key, role, number`)
	if err != nil {
		return nil, oops.Wrapf(err, "finding duplicates")
	}

	groups := lo.GroupBy(all, func(a Application) string {
		return a.CompanyKey + "|" + a.Role
	})

	return lo.Values(groups), nil
}

// SearchApplications performs full-text search on applications via FTS5.
// Uses raw SQL since ksql does not support virtual tables.
func (d *DB) SearchApplications(ctx context.Context, query string) ([]Application, error) {
	var apps []Application
	err := d.ksql.Query(ctx, &apps, `
		SELECT a.id, a.number, a.date, a.company, a.company_key, a.role,
		       a.score, a.score_raw, a.status, a.has_pdf, a.report_number,
		       a.report_path, a.job_url, a.notes, a.archetype, a.tldr,
		       a.remote, a.comp_estimate, a.created_at, a.updated_at
		FROM applications a
		JOIN applications_fts f ON a.id = f.rowid
		WHERE applications_fts MATCH ?
		ORDER BY rank`, query)
	if err != nil {
		return nil, oops.Wrapf(err, "searching applications for %q", query)
	}
	return apps, nil
}

// statusCount is a helper for scanning status breakdown rows.
type statusCount struct {
	Status string `ksql:"status"`
	Count  int    `ksql:"count"`
}

// ComputeMetrics calculates aggregate pipeline statistics.
// Uses raw SQL for complex aggregation since ksql is not suited for it.
func (d *DB) ComputeMetrics(ctx context.Context) (model.PipelineMetrics, error) {
	var m model.PipelineMetrics

	row := d.sql.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(AVG(score), 0), COALESCE(MAX(score), 0),
		       COALESCE(SUM(CASE WHEN has_pdf THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN status IN ('evaluated','applied','interview','responded') THEN 1 ELSE 0 END), 0)
		FROM applications`)

	if err := row.Scan(&m.Total, &m.AvgScore, &m.TopScore, &m.WithPDF, &m.Actionable); err != nil {
		return m, oops.Wrapf(err, "computing aggregate metrics")
	}

	var counts []statusCount
	err := d.ksql.Query(ctx, &counts, `
		SELECT status, COUNT(*) as count FROM applications GROUP BY status`)
	if err != nil {
		return m, oops.Wrapf(err, "computing status breakdown")
	}

	m.ByStatus = lo.Associate(counts, func(sc statusCount) (string, int) {
		return sc.Status, sc.Count
	})

	return m, nil
}
