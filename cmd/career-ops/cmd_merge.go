package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge batch TSV additions into the application database",
	Long: `Reads TSV files from batch/tracker-additions/, parses each entry,
checks for duplicates by company+role (fuzzy) or report number, and either
updates existing entries (if the new score is higher) or inserts new rows
into the SQLite database.`,
	RunE: runMerge,
}

var (
	mergePath    string
	mergeDryRun  bool
	mergeArchive bool
)

func init() {
	mergeCmd.Flags().StringVar(&mergePath, "path", ".", "path to career-ops root directory")
	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "preview changes without writing")
	mergeCmd.Flags().BoolVar(&mergeArchive, "archive", true, "move processed TSVs to merged/ directory")
}

// mergeStats tracks addition/update/skip counts.
type mergeStats struct {
	added   int
	updated int
	skipped int
}

func runMerge(_ *cobra.Command, _ []string) (err error) {
	careerOpsPath, pathErr := filepath.Abs(mergePath)
	if pathErr != nil {
		return oops.Wrapf(pathErr, "resolving path")
	}
	dbPath := viper.GetString("db")
	ctx := context.Background()

	// Initialize states from YAML so Normalize/Label work properly.
	states.Init(careerOpsPath)

	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	// Load existing applications from DB for duplicate detection.
	existingApps, err := r.ListApplications(ctx)
	if err != nil {
		return oops.Wrapf(err, "listing existing applications")
	}

	maxNum := lo.Reduce(existingApps, func(acc int, app model.CareerApplication, _ int) int {
		return lo.Max([]int{acc, app.Number})
	}, 0)

	fmt.Printf("Existing: %d entries, max #%d\n", len(existingApps), maxNum)

	tsvFiles, err := loadTSVAdditions(careerOpsPath)
	if err != nil {
		return err
	}
	if tsvFiles == nil {
		return nil
	}

	fmt.Printf("Found %d pending additions\n", len(tsvFiles))

	stats := mergeStats{}

	lo.ForEach(tsvFiles, func(tsvPath string, _ int) {
		maxNum = processTSVFile(ctx, r, tsvPath, existingApps, maxNum, &stats)
	})

	if !mergeDryRun && mergeArchive {
		if archiveErr := archiveTSVs(tsvFiles, careerOpsPath); archiveErr != nil {
			return archiveErr
		}
	}

	fmt.Printf("\nSummary: +%d added, %d updated, %d skipped\n",
		stats.added, stats.updated, stats.skipped)
	if mergeDryRun {
		fmt.Println("(dry-run - no changes written)")
	}

	return nil
}

// processTSVFile handles a single TSV file: parse, dedup, upsert. Returns updated maxNum.
func processTSVFile(
	ctx context.Context, r db.Repository, tsvPath string,
	existingApps []model.CareerApplication, maxNum int, stats *mergeStats,
) int {
	filename := filepath.Base(tsvPath)
	data, readErr := os.ReadFile(filepath.Clean(tsvPath))
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "warning: reading %s: %v\n", filename, readErr)
		stats.skipped++
		return maxNum
	}

	addition := tracker.ParseTSVContent(string(data), filename)
	if addition == nil {
		stats.skipped++
		return maxNum
	}

	// Check for duplicate by company+role in existing DB entries.
	duplicate, _ := lo.Find(existingApps, func(app model.CareerApplication) bool {
		return tracker.NormalizeCompanyKey(app.Company) == tracker.NormalizeCompanyKey(addition.Company) &&
			tracker.RoleMatch(app.Role, addition.Role)
	})

	if duplicate.Number != 0 {
		return handleMergeDuplicate(ctx, r, addition, &duplicate, stats)
	}
	return handleMergeNew(ctx, r, addition, maxNum, stats)
}

// handleMergeDuplicate updates an existing entry if the new score is higher.
func handleMergeDuplicate(
	ctx context.Context, r db.Repository,
	addition *tracker.Addition, duplicate *model.CareerApplication,
	stats *mergeStats,
) int {
	newScore := tracker.ParseScore(addition.Score)
	oldScore := duplicate.Score

	if newScore <= oldScore {
		fmt.Printf("Skip: %s - %s (existing #%d %.1f >= new %.1f)\n",
			addition.Company, addition.Role, duplicate.Number, oldScore, newScore)
		stats.skipped++
		return 0
	}

	fmt.Printf("Update: #%d %s - %s (%.1f -> %.1f)\n",
		duplicate.Number, addition.Company, addition.Role, oldScore, newScore)

	if !mergeDryRun {
		duplicate.ScoreRaw = addition.Score
		duplicate.Score = newScore
		duplicate.Date = addition.Date
		duplicate.Notes = fmt.Sprintf("Re-eval %s (%.1f->%.1f). %s",
			addition.Date, oldScore, newScore, addition.Notes)
		if upsertErr := r.UpsertApplication(ctx, duplicate); upsertErr != nil {
			fmt.Fprintf(os.Stderr, "warning: upserting #%d: %v\n", duplicate.Number, upsertErr)
		}
	}
	stats.updated++
	return 0
}

// handleMergeNew inserts a new application entry. Returns updated maxNum.
func handleMergeNew(
	ctx context.Context, r db.Repository,
	addition *tracker.Addition, maxNum int,
	stats *mergeStats,
) int {
	entryNum := addition.Num
	if entryNum <= maxNum {
		maxNum++
		entryNum = maxNum
	} else {
		maxNum = entryNum
	}

	fmt.Printf("Add #%d: %s - %s (%s)\n",
		entryNum, addition.Company, addition.Role, addition.Score)

	if !mergeDryRun {
		app := model.CareerApplication{
			Number:       entryNum,
			Date:         addition.Date,
			Company:      addition.Company,
			Role:         addition.Role,
			ScoreRaw:     addition.Score,
			Score:        tracker.ParseScore(addition.Score),
			Status:       addition.Status,
			HasPDF:       addition.PDF == "✅",
			ReportNumber: fmt.Sprintf("%d", tracker.ExtractReportNum(addition.Report)),
			ReportPath:   extractReportPath(addition.Report),
			Notes:        addition.Notes,
		}
		if upsertErr := r.UpsertApplication(ctx, &app); upsertErr != nil {
			fmt.Fprintf(os.Stderr, "warning: inserting #%d: %v\n", entryNum, upsertErr)
		}
	}
	stats.added++
	return maxNum
}

// loadTSVAdditions locates and returns sorted TSV files, or nil if none found.
func loadTSVAdditions(careerOpsPath string) ([]string, error) {
	additionsDir := filepath.Join(careerOpsPath, "batch", "tracker-additions")
	tsvPattern := filepath.Join(additionsDir, "*.tsv")
	tsvFiles, err := filepath.Glob(tsvPattern)
	if err != nil {
		return nil, oops.Wrapf(err, "glob tracker-additions")
	}
	if len(tsvFiles) == 0 {
		fmt.Println("No pending additions to merge.")
		return nil, nil
	}

	// Sort numerically for deterministic processing.
	sort.Slice(tsvFiles, func(i, j int) bool {
		return extractLeadingNum(filepath.Base(tsvFiles[i])) <
			extractLeadingNum(filepath.Base(tsvFiles[j]))
	})
	return tsvFiles, nil
}

// archiveTSVs moves processed TSV files to the merged/ subdirectory.
func archiveTSVs(tsvFiles []string, careerOpsPath string) error {
	additionsDir := filepath.Join(careerOpsPath, "batch", "tracker-additions")
	mergedDir := filepath.Join(additionsDir, "merged")
	if err := os.MkdirAll(mergedDir, 0o750); err != nil {
		return oops.Wrapf(err, "creating merged dir")
	}
	lo.ForEach(tsvFiles, func(tsvPath string, _ int) {
		dest := filepath.Join(mergedDir, filepath.Base(tsvPath))
		if err := os.Rename(tsvPath, dest); err != nil {
			fmt.Fprintf(
				os.Stderr, "warning: moving %s: %v\n",
				filepath.Base(tsvPath), err,
			)
		}
	})
	fmt.Printf("Moved %d TSVs to merged/\n", len(tsvFiles))
	return nil
}

// extractReportPath pulls the file path from a markdown link like [123](reports/123-foo.md).
func extractReportPath(reportField string) string {
	start := strings.Index(reportField, "(")
	end := strings.Index(reportField, ")")
	if start >= 0 && end > start {
		return reportField[start+1 : end]
	}
	return ""
}

// extractLeadingNum extracts the first consecutive digit run from a filename for sorting.
func extractLeadingNum(name string) int {
	isDigit := func(b byte) bool { return b >= '0' && b <= '9' }
	// Skip leading non-digits, then take consecutive digits.
	afterSkip := lo.DropWhile([]byte(name), func(b byte) bool { return !isDigit(b) })
	if len(afterSkip) == 0 {
		return 0
	}
	// Split into [digits, rest] at the first non-digit boundary.
	_, firstNonDigit, hasNonDigit := lo.FindIndexOf(afterSkip, func(b byte) bool { return !isDigit(b) })
	digits := lo.Ternary(hasNonDigit, afterSkip[:firstNonDigit], afterSkip)
	n, err := strconv.Atoi(string(digits))
	if err != nil {
		return 0
	}
	return n
}
