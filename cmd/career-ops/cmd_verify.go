package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
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

func runVerify(_ *cobra.Command, _ []string) error {
	root := verifyPath

	// Initialize states from YAML
	states.Init(root)

	// Try to find and parse applications.md
	appsFile, err := tracker.FindAppsFile(root)
	if err != nil {
		fmt.Println("\nNo applications.md found. This is normal for a fresh setup.")
		fmt.Println("The file will be created when you evaluate your first offer.")
		return nil
	}

	apps, err := tracker.ParseApplications(root)
	if err != nil {
		return fmt.Errorf("parsing applications: %w", err)
	}

	fmt.Printf("\nChecking %d entries in %s\n\n", len(apps), appsFile)

	result := &verifyResult{}

	// Run all checks
	checkStatuses(result, apps)
	checkDuplicates(result, apps)
	checkReportLinks(result, apps, root)
	checkScores(result, apps)
	checkRowFormat(result, root)
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
		os.Exit(1)
	}
	return nil
}

// checkStatuses validates that all statuses are canonical (or normalizable).
func checkStatuses(r *verifyResult, apps []model.CareerApplication) {
	bad := 0
	for _, app := range apps {
		raw := app.Status
		normalized := states.Normalize(raw)

		// Check if the normalized result is a known canonical ID
		if !states.IsCanonical(normalized) {
			r.errorf("#%d: Non-canonical status %q (normalized to %q)", app.Number, raw, normalized)
			bad++
		}

		// Check for markdown bold in status
		if strings.Contains(raw, "**") {
			r.errorf("#%d: Status contains markdown bold: %q", app.Number, raw)
			bad++
		}

		// Check for dates embedded in status
		if reDateInStr.MatchString(raw) {
			r.errorf("#%d: Status contains date: %q -- dates go in date column", app.Number, raw)
			bad++
		}
	}
	if bad == 0 {
		okMsg("All statuses are canonical")
	}
}

// checkDuplicates detects duplicate company+role entries.
func checkDuplicates(r *verifyResult, apps []model.CareerApplication) {
	type group struct {
		company string
		role    string
		nums    []int
	}

	groups := make(map[string]*group)
	for _, app := range apps {
		key := tracker.NormalizeCompanyKey(app.Company) + "::" + strings.ToLower(app.Role)
		if g, ok := groups[key]; ok {
			g.nums = append(g.nums, app.Number)
		} else {
			groups[key] = &group{
				company: app.Company,
				role:    app.Role,
				nums:    []int{app.Number},
			}
		}
	}

	dupes := 0
	for _, g := range groups {
		if len(g.nums) > 1 {
			numStrs := lo.Map(g.nums, func(n int, _ int) string {
				return fmt.Sprintf("#%d", n)
			})
			r.warnf("Possible duplicates: %s (%s -- %s)",
				strings.Join(numStrs, ", "), g.company, g.role)
			dupes++
		}
	}
	if dupes == 0 {
		okMsg("No exact duplicates found")
	}
}

// checkReportLinks validates that all report markdown links point to existing files.
func checkReportLinks(r *verifyResult, apps []model.CareerApplication, root string) {
	broken := 0
	for _, app := range apps {
		if app.ReportPath == "" {
			continue
		}
		fullPath := filepath.Join(root, app.ReportPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			r.errorf("#%d: Report not found: %s", app.Number, app.ReportPath)
			broken++
		}
	}
	if broken == 0 {
		okMsg("All report links valid")
	}
}

// checkScores validates score format (X.X/5, N/A, or DUP).
func checkScores(r *verifyResult, apps []model.CareerApplication) {
	bad := 0
	for _, app := range apps {
		s := strings.ReplaceAll(app.ScoreRaw, "**", "")
		s = strings.TrimSpace(s)
		if s == "N/A" || s == "DUP" || s == "" {
			continue
		}
		if !reScoreFormat.MatchString(s) {
			r.errorf("#%d: Invalid score format: %q", app.Number, app.ScoreRaw)
			bad++
		}
	}
	if bad == 0 {
		okMsg("All scores valid")
	}
}

// checkRowFormat reads the raw file and checks that every table row has enough columns.
func checkRowFormat(r *verifyResult, root string) {
	appsFile, err := tracker.FindAppsFile(root)
	if err != nil {
		return
	}
	content, err := os.ReadFile(appsFile)
	if err != nil {
		return
	}

	bad := 0
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		// Skip header and separator rows
		if strings.Contains(line, "---") || strings.Contains(line, "Company") || strings.Contains(line, "Empresa") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 9 {
			truncated := line
			if len(truncated) > 80 {
				truncated = truncated[:80] + "..."
			}
			r.errorf("Row with <9 columns: %s", truncated)
			bad++
		}
	}
	if bad == 0 {
		okMsg("All rows properly formatted")
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
	bad := 0
	for _, app := range apps {
		if strings.Contains(app.ScoreRaw, "**") {
			r.warnf("#%d: Score has markdown bold: %q", app.Number, app.ScoreRaw)
			bad++
		}
	}
	if bad == 0 {
		okMsg("No bold in scores")
	}
}

// checkDates validates that all date fields match YYYY-MM-DD format.
func checkDates(r *verifyResult, apps []model.CareerApplication) {
	bad := 0
	for _, app := range apps {
		d := strings.TrimSpace(app.Date)
		if d == "" {
			continue
		}
		if !reDateFormat.MatchString(d) {
			r.errorf("#%d: Invalid date format: %q (expected YYYY-MM-DD)", app.Number, d)
			bad++
		}
	}
	if bad == 0 {
		okMsg("All dates valid")
	}
}
