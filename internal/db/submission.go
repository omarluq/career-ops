package db

import (
	"context"

	"github.com/samber/oops"
	"github.com/vingarcia/ksql"
)

// Submission maps to the submissions table for tracking application
// submission attempts and their outcomes.
type Submission struct {
	Portal            string `ksql:"portal"`
	JDURL             string `ksql:"jd_url"`
	CVPDFPath         string `ksql:"cv_pdf_path"`
	Status            string `ksql:"status"`
	ConfirmURL        string `ksql:"confirm_url"`
	ErrorMessage      string `ksql:"error_message"`
	SubmittedAt       string `ksql:"submitted_at"`
	CreatedAt         string `ksql:"created_at"`
	ksql.Table        `ksql:"submissions"`
	ID                int `ksql:"id"`
	ApplicationNumber int `ksql:"application_number"`
	Attempts          int `ksql:"attempts"`
}

// InsertSubmission creates a new submission record.
func (d *DB) InsertSubmission(ctx context.Context, s *Submission) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO submissions (
			application_number, portal, jd_url, cv_pdf_path,
			status, confirm_url, error_message, attempts, submitted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ApplicationNumber, s.Portal, s.JDURL, s.CVPDFPath,
		s.Status, s.ConfirmURL, s.ErrorMessage, s.Attempts, s.SubmittedAt,
	)

	return oops.Wrapf(err, "inserting submission for app=%d", s.ApplicationNumber)
}

// UpdateSubmissionStatus transitions a submission to a new status, optionally
// setting the confirmation URL and error message.
func (d *DB) UpdateSubmissionStatus(
	ctx context.Context,
	id int,
	status, confirmURL, errorMsg string,
) error {
	res, err := d.sql.ExecContext(ctx, `
		UPDATE submissions
		SET status = ?, confirm_url = ?, error_message = ?, attempts = attempts + 1
		WHERE id = ?`,
		status, confirmURL, errorMsg, id,
	)
	if err != nil {
		return oops.Wrapf(err, "updating submission id=%d status=%s", id, status)
	}

	n, rowErr := res.RowsAffected()
	if rowErr != nil {
		return oops.Wrapf(rowErr, "checking rows affected for submission id=%d", id)
	}
	if n == 0 {
		return oops.Errorf("submission id=%d not found", id)
	}

	return nil
}

// ListSubmissionsByApp returns all submissions for a given application number,
// ordered by creation time descending (most recent first).
func (d *DB) ListSubmissionsByApp(ctx context.Context, appNumber int) ([]Submission, error) {
	var subs []Submission
	err := d.ksql.Query(ctx, &subs, `
		SELECT id, application_number, portal, jd_url, cv_pdf_path,
		       status, confirm_url, error_message, attempts,
		       submitted_at, created_at
		FROM submissions
		WHERE application_number = ?
		ORDER BY created_at DESC`, appNumber)
	if err != nil {
		return nil, oops.Wrapf(err, "listing submissions for app=%d", appNumber)
	}

	return subs, nil
}

// ListPendingSubmissions returns all submissions with status 'pending',
// ordered by creation time ascending (oldest first for fairness).
func (d *DB) ListPendingSubmissions(ctx context.Context) ([]Submission, error) {
	var subs []Submission
	err := d.ksql.Query(ctx, &subs, `
		SELECT id, application_number, portal, jd_url, cv_pdf_path,
		       status, confirm_url, error_message, attempts,
		       submitted_at, created_at
		FROM submissions
		WHERE status = 'pending'
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing pending submissions")
	}

	return subs, nil
}
