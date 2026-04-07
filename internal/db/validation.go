package db

import (
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/states"
)

// ---- helpers ----------------------------------------------------------------

// canonicalStatusIDs returns all canonical status IDs from the states package.
func canonicalStatusIDs() []string {
	return lo.Map(states.All(), func(s states.State, _ int) string {
		return s.ID
	})
}

// canonicalStatusLabels returns all canonical status labels from the states package.
func canonicalStatusLabels() []string {
	return lo.Map(states.All(), func(s states.State, _ int) string {
		return s.Label
	})
}

// allValidStatuses returns both IDs and labels as valid status values.
func allValidStatuses() []string {
	ids := canonicalStatusIDs()
	labels := canonicalStatusLabels()
	return lo.Uniq(append(ids, labels...))
}

// IsValidStatus checks whether a status string matches a canonical ID or label (case-insensitive).
func IsValidStatus(status string) bool {
	lower := strings.ToLower(strings.TrimSpace(status))
	return lo.ContainsBy(allValidStatuses(), func(s string) bool {
		return strings.EqualFold(s, lower)
	})
}

// ---- validation helpers ----------------------------------------------------

// validationError collects multiple field-level errors into one.
type validationError struct {
	issues []string
}

// addf records a validation issue for the given field.
func (v *validationError) addf(field, msg string) {
	v.issues = append(v.issues, field+": "+msg)
}

// err returns a combined error or nil when no issues were recorded.
func (v *validationError) err() error {
	if len(v.issues) == 0 {
		return nil
	}
	return oops.Errorf("%s", strings.Join(v.issues, "; "))
}

// requireString checks that a string field is non-empty.
func requireString(ve *validationError, field, value, msg string) {
	if strings.TrimSpace(value) == "" {
		ve.addf(field, msg)
	}
}

// ---- public validators ------------------------------------------------------

// ValidateApplication validates an Application struct and checks that the
// status is a canonical value.
func ValidateApplication(app *Application) error {
	if app == nil {
		return oops.Errorf("application must not be nil")
	}

	ve := &validationError{}

	requireString(ve, "company", app.Company, "company is required")
	requireString(ve, "role", app.Role, "role is required")
	requireString(ve, "status", app.Status, "status is required")
	requireString(ve, "date", app.Date, "date is required")

	if app.Score < 0 {
		ve.addf("score", "score must be >= 0")
	}
	if app.Score > 5 {
		ve.addf("score", "score must be <= 5")
	}
	if app.Number < 0 {
		ve.addf("number", "number must be non-negative")
	}

	if err := ve.err(); err != nil {
		return oops.Wrapf(err, "validating application %q (%s)", app.Company, app.Role)
	}

	if app.Status != "" && !IsValidStatus(app.Status) {
		return oops.Errorf(
			"invalid status %q for application %q (%s): must be one of %v",
			app.Status, app.Company, app.Role, allValidStatuses(),
		)
	}

	return nil
}

// ValidatePipelineEntry validates a PipelineEntry struct against required fields.
func ValidatePipelineEntry(entry *PipelineEntry) error {
	if entry == nil {
		return oops.Errorf("pipeline entry must not be nil")
	}

	ve := &validationError{}

	requireString(ve, "url", entry.URL, "url is required")
	requireString(ve, "source", entry.Source, "source is required")
	requireString(ve, "status", entry.Status, "status is required")

	if err := ve.err(); err != nil {
		return oops.Wrapf(err, "validating pipeline entry (url=%s)", entry.URL)
	}

	return nil
}

// ValidateScanRecord validates a ScanRecord struct against required fields.
func ValidateScanRecord(record *ScanRecord) error {
	if record == nil {
		return oops.Errorf("scan record must not be nil")
	}

	ve := &validationError{}

	requireString(ve, "url", record.URL, "url is required")
	requireString(ve, "portal", record.Portal, "portal is required")
	requireString(ve, "title", record.Title, "title is required")

	if err := ve.err(); err != nil {
		return oops.Wrapf(err, "validating scan record (url=%s)", record.URL)
	}

	return nil
}

// ValidateEvaluation validates an Evaluation struct against required fields and ranges.
func ValidateEvaluation(eval *Evaluation) error {
	if eval == nil {
		return oops.Errorf("evaluation must not be nil")
	}

	ve := &validationError{}

	if eval.ApplicationID < 1 {
		ve.addf("application_id", "application_id must be positive")
	}
	requireString(ve, "evaluation_date", eval.EvaluationDate, "evaluation_date is required")

	if eval.RawScore < 0 {
		ve.addf("raw_score", "raw_score must be >= 0")
	}
	if eval.RawScore > 5 {
		ve.addf("raw_score", "raw_score must be <= 5")
	}

	if err := ve.err(); err != nil {
		return oops.Wrapf(err, "validating evaluation (app_id=%d)", eval.ApplicationID)
	}

	return nil
}
