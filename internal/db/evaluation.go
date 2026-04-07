package db

import (
	"context"

	"github.com/samber/oops"
	"github.com/vingarcia/ksql"
)

// evaluationsTable is the ksql table reference for evaluations.
var evaluationsTable = ksql.NewTable("evaluations", "id")

// Evaluation maps to the evaluations table.
type Evaluation struct {
	BatchID        string `ksql:"batch_id"`
	EvaluationDate string `ksql:"evaluation_date"`
	Archetype      string `ksql:"archetype"`
	TlDr           string `ksql:"tldr"`
	Remote         string `ksql:"remote"`
	CompEstimate   string `ksql:"comp_estimate"`
	ReportPath     string `ksql:"report_path"`
	ksql.Table     `ksql:"evaluations"`
	ID             int     `ksql:"id"`
	ApplicationID  int     `ksql:"application_id"`
	RawScore       float64 `ksql:"raw_score"`
}

// InsertEvaluation creates a new evaluation record using ksql.
func (d *DB) InsertEvaluation(ctx context.Context, eval *Evaluation) error {
	err := d.ksql.Insert(ctx, evaluationsTable, eval)
	return oops.Wrapf(err, "inserting evaluation for application_id=%d", eval.ApplicationID)
}

// GetEvaluation returns a single evaluation by ID.
func (d *DB) GetEvaluation(ctx context.Context, id int) (Evaluation, error) {
	var eval Evaluation
	err := d.ksql.QueryOne(ctx, &eval, `
		SELECT id, application_id, batch_id, evaluation_date, raw_score,
		       archetype, tldr, remote, comp_estimate, report_path
		FROM evaluations
		WHERE id = ?`, id)
	if err != nil {
		return eval, oops.Wrapf(err, "getting evaluation id=%d", id)
	}
	return eval, nil
}

// ListEvaluationsByApplication returns all evaluations for a given application.
func (d *DB) ListEvaluationsByApplication(ctx context.Context, appID int) ([]Evaluation, error) {
	var evals []Evaluation
	err := d.ksql.Query(ctx, &evals, `
		SELECT id, application_id, batch_id, evaluation_date, raw_score,
		       archetype, tldr, remote, comp_estimate, report_path
		FROM evaluations
		WHERE application_id = ?
		ORDER BY evaluation_date DESC`, appID)
	if err != nil {
		return nil, oops.Wrapf(err, "listing evaluations for application_id=%d", appID)
	}
	return evals, nil
}

// ListEvaluationsByBatch returns all evaluations for a given batch.
func (d *DB) ListEvaluationsByBatch(ctx context.Context, batchID string) ([]Evaluation, error) {
	var evals []Evaluation
	err := d.ksql.Query(ctx, &evals, `
		SELECT id, application_id, batch_id, evaluation_date, raw_score,
		       archetype, tldr, remote, comp_estimate, report_path
		FROM evaluations
		WHERE batch_id = ?
		ORDER BY evaluation_date DESC`, batchID)
	if err != nil {
		return nil, oops.Wrapf(err, "listing evaluations for batch_id=%s", batchID)
	}
	return evals, nil
}

// UpdateEvaluation patches an existing evaluation using ksql.
func (d *DB) UpdateEvaluation(ctx context.Context, eval *Evaluation) error {
	err := d.ksql.Patch(ctx, evaluationsTable, eval)
	return oops.Wrapf(err, "updating evaluation id=%d", eval.ID)
}

// DeleteEvaluation removes an evaluation by ID.
func (d *DB) DeleteEvaluation(ctx context.Context, id int) error {
	err := d.ksql.Delete(ctx, evaluationsTable, id)
	return oops.Wrapf(err, "deleting evaluation id=%d", id)
}
