// Package applicator provides high-throughput application submission automation
// with per-portal rate limiting and ATS-specific form handling.
package applicator

import "time"

// SubmissionJob represents a single application to submit.
type SubmissionJob struct {
	// FormData holds pre-filled form field values keyed by field name.
	FormData          map[string]string
	Company           string
	Role              string
	JDURL             string     // original job listing URL
	CVPDF             string     // path to generated CV PDF
	CoverLetter       string     // generated cover letter text
	Portal            PortalType // detected or configured ATS portal
	ApplicationNumber int
}

// SubmissionResult captures the outcome of a single submission attempt.
type SubmissionResult struct {
	SubmittedAt  time.Time
	Status       SubmissionStatus
	ConfirmURL   string
	ErrorMessage string
	Job          SubmissionJob
	Duration     time.Duration
}

// SubmissionStatus enumerates possible submission outcomes.
type SubmissionStatus string

const (
	// StatusSubmitted indicates the application was successfully submitted.
	StatusSubmitted SubmissionStatus = "submitted"
	// StatusFailed indicates the submission encountered an error.
	StatusFailed SubmissionStatus = "failed"
	// StatusSkipped indicates the form was too complex and needs human review.
	StatusSkipped SubmissionStatus = "skipped"
	// StatusPending indicates the job is queued for retry.
	StatusPending SubmissionStatus = "pending"
)

// PortalType identifies an applicant tracking system.
type PortalType string

// Portal type constants for supported ATS platforms.
const (
	PortalGreenhouse PortalType = "greenhouse"
	PortalLever      PortalType = "lever"
	PortalWorkable   PortalType = "workable"
	PortalLinkedIn   PortalType = "linkedin"
	PortalAshby      PortalType = "ashby"
	PortalGeneric    PortalType = "generic"
)
