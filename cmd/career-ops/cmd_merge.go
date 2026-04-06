package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
	"github.com/spf13/cobra"
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
	Num     int
	Date    string
	Company string
	Role    string
	Score   string
	Status  string
	PDF     string
	Report  string
	Notes   string
	Raw     string
}

// parseAppLineRaw parses a markdown table line into a parsedAppLine.
// Table format: | num | date | company | role | score | status | pdf | report | notes |
func parseAppLineRaw(line string) *parsedAppLine {
	parts := strings.Split(line, "|")
	var fields []string
	for _, p := range parts {
		fields = append(fields, strings.TrimSpace(p))
	}
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
		return fmt.Errorf("resolving path: %w", err)
	}

	// Initialize states from YAML so Normalize/Label work properly.
	states.Init(careerOpsPath)

	// Locate applications.md
	appsFile, err := tracker.FindAppsFile(careerOpsPath)
	if err != nil {
		fmt.Println("No applications.md found. Nothing to merge into.")
		return nil
	}

	content, err := os.ReadFile(appsFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", appsFile, err)
	}
	appLines := strings.Split(string(content), "\n")

	// Parse existing entries.
	var existingApps []*parsedAppLine
	maxNum := 0
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
		if app.Num > maxNum {
			maxNum = app.Num
		}
	}

	fmt.Printf("Existing: %d entries, max #%d\n", len(existingApps), maxNum)

	// Locate tracker additions.
	additionsDir := filepath.Join(careerOpsPath, "batch", "tracker-additions")
	tsvPattern := filepath.Join(additionsDir, "*.tsv")
	tsvFiles, err := filepath.Glob(tsvPattern)
	if err != nil {
		return fmt.Errorf("glob tracker-additions: %w", err)
	}
	if len(tsvFiles) == 0 {
		fmt.Println("No pending additions to merge.")
		return nil
	}

	// Sort numerically for deterministic processing.
	sort.Slice(tsvFiles, func(i, j int) bool {
		return extractLeadingNum(filepath.Base(tsvFiles[i])) < extractLeadingNum(filepath.Base(tsvFiles[j]))
	})

	fmt.Printf("Found %d pending additions\n", len(tsvFiles))

	added := 0
	updated := 0
	skipped := 0
	var newLines []string

	for _, tsvPath := range tsvFiles {
		filename := filepath.Base(tsvPath)
		data, err := os.ReadFile(tsvPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: reading %s: %v\n", filename, err)
			skipped++
			continue
		}

		addition := tracker.ParseTSVContent(string(data), filename)
		if addition == nil {
			skipped++
			continue
		}

		// Check for duplicate:
		// 1. Exact report number match
		// 2. Exact entry number match
		// 3. Company + role fuzzy match
		reportNum := tracker.ExtractReportNum(addition.Report)
		var duplicate *parsedAppLine

		if reportNum != 0 {
			for _, app := range existingApps {
				if tracker.ExtractReportNum(app.Report) == reportNum {
					duplicate = app
					break
				}
			}
		}

		if duplicate == nil {
			for _, app := range existingApps {
				if app.Num == addition.Num {
					duplicate = app
					break
				}
			}
		}

		if duplicate == nil {
			normCompany := tracker.NormalizeCompanyKey(addition.Company)
			for _, app := range existingApps {
				if tracker.NormalizeCompanyKey(app.Company) != normCompany {
					continue
				}
				if tracker.RoleMatch(addition.Role, app.Role) {
					duplicate = app
					break
				}
			}
		}

		if duplicate != nil {
			newScore := tracker.ParseScore(addition.Score)
			oldScore := tracker.ParseScore(duplicate.Score)

			if newScore > oldScore {
				fmt.Printf("Update: #%d %s - %s (%.1f -> %.1f)\n",
					duplicate.Num, addition.Company, addition.Role, oldScore, newScore)

				// Find the raw line in appLines and replace it.
				for i, line := range appLines {
					if line == duplicate.Raw {
						// Output format: score before status (markdown table order).
						notes := fmt.Sprintf("Re-eval %s (%.1f->%.1f). %s",
							addition.Date, oldScore, newScore, addition.Notes)
						appLines[i] = tracker.FormatTableLine(
							duplicate.Num, addition.Date, addition.Company, addition.Role,
							addition.Score, duplicate.Status, duplicate.PDF, addition.Report,
							strings.TrimSpace(notes),
						)
						updated++
						break
					}
				}
			} else {
				fmt.Printf("Skip: %s - %s (existing #%d %.1f >= new %.1f)\n",
					addition.Company, addition.Role, duplicate.Num, oldScore, newScore)
				skipped++
			}
		} else {
			// New entry — use the TSV number if it's beyond maxNum, otherwise assign next.
			entryNum := addition.Num
			if entryNum <= maxNum {
				maxNum++
				entryNum = maxNum
			} else {
				maxNum = entryNum
			}

			// Output format: score before status (markdown table order).
			newLine := tracker.FormatTableLine(
				entryNum, addition.Date, addition.Company, addition.Role,
				addition.Score, addition.Status, addition.PDF, addition.Report,
				addition.Notes,
			)
			newLines = append(newLines, newLine)
			added++
			fmt.Printf("Add #%d: %s - %s (%s)\n", entryNum, addition.Company, addition.Role, addition.Score)
		}
	}

	// Insert new lines after the header separator (|---|...).
	if len(newLines) > 0 {
		insertIdx := -1
		for i, line := range appLines {
			if strings.HasPrefix(line, "|") && strings.Contains(line, "---") {
				insertIdx = i + 1
				break
			}
		}
		if insertIdx >= 0 {
			// Splice new lines in.
			updated := make([]string, 0, len(appLines)+len(newLines))
			updated = append(updated, appLines[:insertIdx]...)
			updated = append(updated, newLines...)
			updated = append(updated, appLines[insertIdx:]...)
			appLines = updated
		}
	}

	// Write back.
	if !mergeDryRun {
		output := strings.Join(appLines, "\n")
		if err := tracker.BackupAndWrite(appsFile, []byte(output)); err != nil {
			return fmt.Errorf("writing %s: %w", appsFile, err)
		}

		// Move processed files to merged/.
		if mergeArchive {
			mergedDir := filepath.Join(additionsDir, "merged")
			if err := os.MkdirAll(mergedDir, 0755); err != nil {
				return fmt.Errorf("creating merged dir: %w", err)
			}
			for _, tsvPath := range tsvFiles {
				dest := filepath.Join(mergedDir, filepath.Base(tsvPath))
				if err := os.Rename(tsvPath, dest); err != nil {
					fmt.Fprintf(os.Stderr, "warning: moving %s: %v\n", filepath.Base(tsvPath), err)
				}
			}
			fmt.Printf("Moved %d TSVs to merged/\n", len(tsvFiles))
		}
	}

	fmt.Printf("\nSummary: +%d added, %d updated, %d skipped\n", added, updated, skipped)
	if mergeDryRun {
		fmt.Println("(dry-run - no changes written)")
	}

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
	n, _ := strconv.Atoi(string(digits))
	return n
}
