package db

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
)

var (
	reVerifyScore = regexp.MustCompile(`^\d+\.?\d*/5$`)
	reVerifyDate  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// VerifyResult holds all issues found during verification.
type VerifyResult struct {
	InvalidStatuses []StatusIssue
	Duplicates      []DuplicateCluster
	MissingReports  []MissingReport
	InvalidScores   []ScoreIssue
	InvalidDates    []DateIssue
	Total           int
	IssueCount      int
}

// StatusIssue records an application with a non-canonical status.
type StatusIssue struct {
	Raw    string
	Number int
	Line   int
}

// MissingReport records an application whose report file does not exist on disk.
type MissingReport struct {
	ReportPath string
	Number     int
}

// ScoreIssue records an application with a malformed score field.
type ScoreIssue struct {
	ScoreRaw string
	Number   int
}

// DateIssue records an application with a malformed date field.
type DateIssue struct {
	Date   string
	Number int
}

// Verify runs all verification checks against the repository.
func Verify(ctx context.Context, r Repository, careerOpsPath string) (*VerifyResult, error) {
	apps, err := r.ListApplications(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing applications")
	}

	dupes, err := r.FindDuplicates(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "finding duplicates")
	}

	result := &VerifyResult{
		Duplicates: dupes,
		Total:      len(apps),
	}

	checkAppStatuses(result, apps)
	checkAppReports(result, apps, careerOpsPath)
	checkAppScores(result, apps)
	checkAppDates(result, apps)

	result.IssueCount = len(result.InvalidStatuses) +
		len(result.Duplicates) +
		len(result.MissingReports) +
		len(result.InvalidScores) +
		len(result.InvalidDates)

	return result, nil
}

// checkAppStatuses validates that every application has a canonical status.
func checkAppStatuses(result *VerifyResult, apps []model.CareerApplication) {
	result.InvalidStatuses = lo.FilterMap(apps, func(app model.CareerApplication, _ int) (StatusIssue, bool) {
		normalized := states.Normalize(app.Status)
		if states.IsCanonical(normalized) {
			return StatusIssue{}, false
		}
		return StatusIssue{Number: app.Number, Raw: app.Status}, true
	})
}

// checkAppReports validates that every referenced report file exists on disk.
func checkAppReports(result *VerifyResult, apps []model.CareerApplication, root string) {
	result.MissingReports = lo.FilterMap(apps, func(app model.CareerApplication, _ int) (MissingReport, bool) {
		if app.ReportPath == "" {
			return MissingReport{}, false
		}
		fullPath := filepath.Join(root, app.ReportPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return MissingReport{Number: app.Number, ReportPath: app.ReportPath}, true
		}
		return MissingReport{}, false
	})
}

// checkAppScores validates that score fields match the expected X.X/5 format.
func checkAppScores(result *VerifyResult, apps []model.CareerApplication) {
	result.InvalidScores = lo.FilterMap(apps, func(app model.CareerApplication, _ int) (ScoreIssue, bool) {
		s := strings.ReplaceAll(app.ScoreRaw, "**", "")
		s = strings.TrimSpace(s)
		if s == "" || s == "N/A" || s == "DUP" {
			return ScoreIssue{}, false
		}
		if reVerifyScore.MatchString(s) {
			return ScoreIssue{}, false
		}
		return ScoreIssue{Number: app.Number, ScoreRaw: app.ScoreRaw}, true
	})
}

// checkAppDates validates that date fields match YYYY-MM-DD format.
func checkAppDates(result *VerifyResult, apps []model.CareerApplication) {
	result.InvalidDates = lo.FilterMap(apps, func(app model.CareerApplication, _ int) (DateIssue, bool) {
		d := strings.TrimSpace(app.Date)
		if d == "" {
			return DateIssue{}, false
		}
		if reVerifyDate.MatchString(d) {
			return DateIssue{}, false
		}
		return DateIssue{Number: app.Number, Date: d}, true
	})
}
