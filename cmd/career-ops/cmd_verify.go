package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
)

var (
	reScoreFormat = regexp.MustCompile(`^\d+\.?\d*/5$`)
	reDateFormat  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	reDateInStr   = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
)

var verifyPath string

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run pipeline health checks",
	Long:  "Validates statuses, detects duplicates, and checks report links in the application tracker.",
	RunE:  runVerify,
}

func init() {
	verifyCmd.Flags().StringVar(&verifyPath, "path", ".", "path to career-ops root directory")
}

type verifyResult struct {
	errors   int
	warnings int
}

func (v *verifyResult) errorf(format string, args ...any) {
	fmt.Printf("  ERR  %s\n", fmt.Sprintf(format, args...))
	v.errors++
}

func (v *verifyResult) warnf(format string, args ...any) {
	fmt.Printf("  WARN %s\n", fmt.Sprintf(format, args...))
	v.warnings++
}

func okMsg(msg string) {
	fmt.Printf("  OK   %s\n", msg)
}

func runVerify(_ *cobra.Command, _ []string) (err error) {
	root := verifyPath
	dbPath := viper.GetString("db")
	ctx := context.Background()

	// Initialize states from YAML
	states.Init(root)

	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	apps, err := r.ListApplications(ctx)
	if err != nil {
		return oops.Wrapf(err, "listing applications")
	}

	if len(apps) == 0 {
		fmt.Println("\nNo applications found in database. This is normal for a fresh setup.")
		fmt.Println("Run 'career-ops import' to import existing data, or evaluate your first offer.")
		return nil
	}

	fmt.Printf("\nChecking %d entries from database\n\n", len(apps))

	result := &verifyResult{}

	// Run all checks
	checkStatuses(result, apps)
	checkDuplicates(result, apps)
	checkReportLinks(result, apps, root)
	checkScores(result, apps)
	checkPendingTSVs(result, root)
	checkBoldScores(result, apps)
	checkDates(result, apps)

	// Summary
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Pipeline Health: %d errors, %d warnings\n", result.errors, result.warnings)

	switch {
	case result.errors == 0 && result.warnings == 0:
		fmt.Println("Pipeline is clean!")
	case result.errors == 0:
		fmt.Println("Pipeline OK with warnings")
	default:
		fmt.Println("Pipeline has errors -- fix before proceeding")
	}

	if result.errors > 0 {
		return oops.Errorf("pipeline has %d error(s)", result.errors)
	}
	return nil
}

// checkStatuses validates that all statuses are canonical (or normalizable).
func checkStatuses(r *verifyResult, apps []model.CareerApplication) {
	bad := lo.CountBy(apps, func(app model.CareerApplication) bool {
		raw := app.Status
		normalized := states.Normalize(raw)

		hasErr := false
		// Check if the normalized result is a known canonical ID
		if !states.IsCanonical(normalized) {
			r.errorf("#%d: Non-canonical status %q (normalized to %q)", app.Number, raw, normalized)
			hasErr = true
		}

		// Check for markdown bold in status
		if strings.Contains(raw, "**") {
			r.errorf("#%d: Status contains markdown bold: %q", app.Number, raw)
			hasErr = true
		}

		// Check for dates embedded in status
		if reDateInStr.MatchString(raw) {
			r.errorf("#%d: Status contains date: %q -- dates go in date column", app.Number, raw)
			hasErr = true
		}

		return hasErr
	})
	if bad == 0 {
		okMsg("All statuses are canonical")
	}
}

// checkDuplicates detects duplicate company+role entries.
func checkDuplicates(r *verifyResult, apps []model.CareerApplication) {
	type appKey struct {
		key     string
		company string
		role    string
		num     int
	}

	keyed := lo.Map(apps, func(app model.CareerApplication, _ int) appKey {
		return appKey{
			key:     strings.ToLower(app.Company) + "::" + strings.ToLower(app.Role),
			company: app.Company,
			role:    app.Role,
			num:     app.Number,
		}
	})
	groups := lo.GroupBy(keyed, func(k appKey) string { return k.key })

	dupes := lo.CountBy(lo.Values(groups), func(entries []appKey) bool {
		if len(entries) <= 1 {
			return false
		}
		numStrs := lo.Map(entries, func(e appKey, _ int) string {
			return fmt.Sprintf("#%d", e.num)
		})
		r.warnf("Possible duplicates: %s (%s -- %s)",
			strings.Join(numStrs, ", "), entries[0].company, entries[0].role)
		return true
	})
	if dupes == 0 {
		okMsg("No exact duplicates found")
	}
}

// checkReportLinks validates that all report markdown links point to existing files.
func checkReportLinks(r *verifyResult, apps []model.CareerApplication, root string) {
	broken := lo.CountBy(apps, func(app model.CareerApplication) bool {
		if app.ReportPath == "" {
			return false
		}
		fullPath := filepath.Join(root, app.ReportPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			r.errorf("#%d: Report not found: %s", app.Number, app.ReportPath)
			return true
		}
		return false
	})
	if broken == 0 {
		okMsg("All report links valid")
	}
}

// checkScores validates score format (X.X/5, N/A, or DUP).
func checkScores(r *verifyResult, apps []model.CareerApplication) {
	bad := lo.CountBy(apps, func(app model.CareerApplication) bool {
		s := strings.ReplaceAll(app.ScoreRaw, "**", "")
		s = strings.TrimSpace(s)
		if s == "N/A" || s == "DUP" || s == "" {
			return false
		}
		if !reScoreFormat.MatchString(s) {
			r.errorf("#%d: Invalid score format: %q", app.Number, app.ScoreRaw)
			return true
		}
		return false
	})
	if bad == 0 {
		okMsg("All scores valid")
	}
}

// checkPendingTSVs warns about unmerged TSV files in tracker-additions/.
func checkPendingTSVs(r *verifyResult, root string) {
	additionsDir := filepath.Join(root, "batch", "tracker-additions")
	entries, err := os.ReadDir(additionsDir)
	if err != nil {
		// Directory doesn't exist -- that's fine
		okMsg("No pending TSVs")
		return
	}

	tsvFiles := lo.Filter(entries, func(e os.DirEntry, _ int) bool {
		return !e.IsDir() && strings.HasSuffix(e.Name(), ".tsv")
	})

	if len(tsvFiles) > 0 {
		r.warnf("%d pending TSVs in tracker-additions/ (not merged)", len(tsvFiles))
	} else {
		okMsg("No pending TSVs")
	}
}

// checkBoldScores warns about markdown bold in score fields.
func checkBoldScores(r *verifyResult, apps []model.CareerApplication) {
	bad := lo.CountBy(apps, func(app model.CareerApplication) bool {
		if strings.Contains(app.ScoreRaw, "**") {
			r.warnf("#%d: Score has markdown bold: %q", app.Number, app.ScoreRaw)
			return true
		}
		return false
	})
	if bad == 0 {
		okMsg("No bold in scores")
	}
}

// checkDates validates that all date fields match YYYY-MM-DD format.
func checkDates(r *verifyResult, apps []model.CareerApplication) {
	bad := lo.CountBy(apps, func(app model.CareerApplication) bool {
		d := strings.TrimSpace(app.Date)
		if d == "" {
			return false
		}
		if !reDateFormat.MatchString(d) {
			r.errorf("#%d: Invalid date format: %q (expected YYYY-MM-DD)", app.Number, d)
			return true
		}
		return false
	})
	if bad == 0 {
		okMsg("All dates valid")
	}
}
