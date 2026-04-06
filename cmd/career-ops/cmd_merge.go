package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"

	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge batch TSV additions into the application tracker",
	Long: `Reads TSV files from batch/tracker-additions/, parses each entry,
checks for duplicates by company+role (fuzzy) or report number, and either
updates existing entries (if the new score is higher) or appends new rows.

The TSV format has status before score, but the markdown table has score
before status — the swap is handled automatically.

Creates a .bak backup before writing changes.`,
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

// parsedAppLine holds a parsed application row plus its raw line for in-place updates.
type parsedAppLine struct {
	Date    string
	Company string
	Role    string
	Score   string
	Status  string
	PDF     string
	Report  string
	Notes   string
	Raw     string
	Num     int
}

// parseAppLineRaw parses a markdown table line into a parsedAppLine.
// Table format: | num | date | company | role | score | status | pdf | report | notes |.
func parseAppLineRaw(line string) *parsedAppLine {
	parts := strings.Split(line, "|")
	fields := lo.Map(parts, func(p string, _ int) string { return strings.TrimSpace(p) })
	// After splitting "| a | b | ... |", fields[0] and fields[len-1] are empty.
	if len(fields) < 10 {
		return nil
	}
	num, err := strconv.Atoi(fields[1])
	if err != nil || num == 0 {
		return nil
	}
	return &parsedAppLine{
		Num:     num,
		Date:    fields[2],
		Company: fields[3],
		Role:    fields[4],
		Score:   fields[5],
		Status:  fields[6],
		PDF:     fields[7],
		Report:  fields[8],
		Notes:   safeField(fields, 9),
		Raw:     line,
	}
}

func safeField(fields []string, i int) string {
	if i < len(fields) {
		return fields[i]
	}
	return ""
}

func runMerge(_ *cobra.Command, _ []string) error {
	careerOpsPath, err := filepath.Abs(mergePath)
	if err != nil {
		return oops.Wrapf(err, "resolving path")
	}

	// Initialize states from YAML so Normalize/Label work properly.
	states.Init(careerOpsPath)

	// Locate applications.md
	appsFile, err := tracker.FindAppsFile(careerOpsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No applications.md found. Nothing to merge into.")
			return nil
		}
		return oops.Wrapf(err, "finding applications file")
	}

	appLines, existingApps, maxNum, err := loadExistingApps(appsFile)
	if err != nil {
		return err
	}

	fmt.Printf("Existing: %d entries, max #%d\n", len(existingApps), maxNum)

	tsvFiles, err := loadTSVAdditions(careerOpsPath)
	if err != nil {
		return err
	}
	if tsvFiles == nil {
		return nil
	}

	fmt.Printf("Found %d pending additions\n", len(tsvFiles))

	appLines, _, stats := processTSVAdditions(
		tsvFiles, appLines, existingApps, maxNum,
	)

	// Write back.
	if !mergeDryRun {
		if err := writeAndArchive(
			appsFile, appLines, tsvFiles, careerOpsPath,
		); err != nil {
			return err
		}
	}

	fmt.Printf(
		"\nSummary: +%d added, %d updated, %d skipped\n",
		stats.added, stats.updated, stats.skipped,
	)
	if mergeDryRun {
		fmt.Println("(dry-run - no changes written)")
	}

	return nil
}

// loadExistingApps reads applications.md and parses existing entries.
func loadExistingApps(
	appsFile string,
) (appLines []string, existingApps []*parsedAppLine, maxNum int, err error) {
	content, err := os.ReadFile(filepath.Clean(appsFile))
	if err != nil {
		return nil, nil, 0, oops.Wrapf(err, "reading %s", appsFile)
	}
	appLines = strings.Split(string(content), "\n")

	maxNum = 0
	for _, line := range appLines {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if strings.Contains(line, "---") {
			continue
		}
		app := parseAppLineRaw(line)
		if app == nil {
			continue
		}
		existingApps = append(existingApps, app)
		maxNum = lo.Max([]int{maxNum, app.Num})
	}
	return appLines, existingApps, maxNum, nil
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

// mergeStats tracks addition/update/skip counts.
type mergeStats struct {
	added   int
	updated int
	skipped int
}

// processTSVAdditions processes each TSV file against existing apps.
func processTSVAdditions(
	tsvFiles []string,
	appLines []string,
	existingApps []*parsedAppLine,
	maxNum int,
) (updatedLines []string, newMaxNum int, stats mergeStats) {
	var newLines []string

	for _, tsvPath := range tsvFiles {
		filename := filepath.Base(tsvPath)
		data, err := os.ReadFile(filepath.Clean(tsvPath))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: reading %s: %v\n", filename, err)
			stats.skipped++
			continue
		}

		addition := tracker.ParseTSVContent(string(data), filename)
		if addition == nil {
			stats.skipped++
			continue
		}

		duplicate := findDuplicate(addition, existingApps)

		if duplicate != nil {
			appLines, stats = handleDuplicateAddition(
				addition, duplicate, appLines, stats,
			)
		} else {
			newLine, num := appendNew(addition, maxNum)
			maxNum = num
			newLines = append(newLines, newLine)
			stats.added++
			fmt.Printf(
				"Add #%d: %s - %s (%s)\n",
				num, addition.Company, addition.Role, addition.Score,
			)
		}
	}

	// Insert new lines after the header separator (|---|...).
	if len(newLines) > 0 {
		appLines = insertAfterSeparator(appLines, newLines)
	}

	return appLines, maxNum, stats
}

// findDuplicate checks for an existing entry matching the addition by
// report number, entry number, or company+role fuzzy match.
func findDuplicate(
	addition *tracker.Addition,
	existingApps []*parsedAppLine,
) *parsedAppLine {
	reportNum := tracker.ExtractReportNum(addition.Report)

	if reportNum != 0 {
		if match, found := lo.Find(existingApps, func(app *parsedAppLine) bool {
			return tracker.ExtractReportNum(app.Report) == reportNum
		}); found {
			return match
		}
	}

	if match, found := lo.Find(existingApps, func(app *parsedAppLine) bool {
		return app.Num == addition.Num
	}); found {
		return match
	}

	normCompany := tracker.NormalizeCompanyKey(addition.Company)
	if match, found := lo.Find(existingApps, func(app *parsedAppLine) bool {
		return tracker.NormalizeCompanyKey(app.Company) == normCompany &&
			tracker.RoleMatch(addition.Role, app.Role)
	}); found {
		return match
	}

	return nil
}

// handleDuplicateAddition updates an existing entry if the new score is higher.
func handleDuplicateAddition(
	addition *tracker.Addition,
	duplicate *parsedAppLine,
	appLines []string,
	stats mergeStats,
) ([]string, mergeStats) {
	newScore := tracker.ParseScore(addition.Score)
	oldScore := tracker.ParseScore(duplicate.Score)

	if newScore <= oldScore {
		fmt.Printf(
			"Skip: %s - %s (existing #%d %.1f >= new %.1f)\n",
			addition.Company, addition.Role,
			duplicate.Num, oldScore, newScore,
		)
		stats.skipped++
		return appLines, stats
	}

	fmt.Printf(
		"Update: #%d %s - %s (%.1f -> %.1f)\n",
		duplicate.Num, addition.Company, addition.Role,
		oldScore, newScore,
	)

	// Find the raw line in appLines and replace it.
	for i, line := range appLines {
		if line == duplicate.Raw {
			notes := fmt.Sprintf(
				"Re-eval %s (%.1f->%.1f). %s",
				addition.Date, oldScore, newScore, addition.Notes,
			)
			appLines[i] = tracker.FormatTableLine(
				duplicate.Num, addition.Date,
				addition.Company, addition.Role,
				addition.Score, duplicate.Status,
				duplicate.PDF, addition.Report,
				strings.TrimSpace(notes),
			)
			stats.updated++
			break
		}
	}
	return appLines, stats
}

// appendNew creates a new table line for a non-duplicate addition.
func appendNew(
	addition *tracker.Addition, maxNum int,
) (newLine string, newMaxNum int) {
	entryNum := addition.Num
	if entryNum <= maxNum {
		maxNum++
		entryNum = maxNum
	} else {
		maxNum = entryNum
	}

	newLine = tracker.FormatTableLine(
		entryNum, addition.Date, addition.Company, addition.Role,
		addition.Score, addition.Status, addition.PDF, addition.Report,
		addition.Notes,
	)
	return newLine, maxNum
}

// insertAfterSeparator splices new lines after the markdown table header separator.
func insertAfterSeparator(appLines, newLines []string) []string {
	insertIdx := -1
	for i, line := range appLines {
		if strings.HasPrefix(line, "|") && strings.Contains(line, "---") {
			insertIdx = i + 1
			break
		}
	}
	if insertIdx >= 0 {
		updated := make([]string, 0, len(appLines)+len(newLines))
		updated = append(updated, appLines[:insertIdx]...)
		updated = append(updated, newLines...)
		updated = append(updated, appLines[insertIdx:]...)
		appLines = updated
	}
	return appLines
}

// writeAndArchive writes the merged content and optionally archives TSVs.
func writeAndArchive(
	appsFile string,
	appLines []string,
	tsvFiles []string,
	careerOpsPath string,
) error {
	output := strings.Join(appLines, "\n")
	if err := tracker.BackupAndWrite(appsFile, []byte(output)); err != nil {
		return oops.Wrapf(err, "writing %s", appsFile)
	}

	if mergeArchive {
		return archiveTSVs(tsvFiles, careerOpsPath)
	}
	return nil
}

// archiveTSVs moves processed TSV files to the merged/ subdirectory.
func archiveTSVs(tsvFiles []string, careerOpsPath string) error {
	additionsDir := filepath.Join(careerOpsPath, "batch", "tracker-additions")
	mergedDir := filepath.Join(additionsDir, "merged")
	if err := os.MkdirAll(mergedDir, 0o750); err != nil {
		return oops.Wrapf(err, "creating merged dir")
	}
	for _, tsvPath := range tsvFiles {
		dest := filepath.Join(mergedDir, filepath.Base(tsvPath))
		if err := os.Rename(tsvPath, dest); err != nil {
			fmt.Fprintf(
				os.Stderr, "warning: moving %s: %v\n",
				filepath.Base(tsvPath), err,
			)
		}
	}
	fmt.Printf("Moved %d TSVs to merged/\n", len(tsvFiles))
	return nil
}

// extractLeadingNum extracts the leading number from a filename for sorting.
func extractLeadingNum(name string) int {
	var digits []byte
	for _, b := range []byte(name) {
		if b >= '0' && b <= '9' {
			digits = append(digits, b)
		} else if len(digits) > 0 {
			break
		}
	}
	n, err := strconv.Atoi(string(digits))
	if err != nil {
		return 0
	}
	return n
}
