package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"

	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
)

var normalizeCmd = &cobra.Command{
	Use:   "normalize",
	Short: "Normalize status aliases to canonical statuses",
	Long: "Reads applications.md, maps non-canonical statuses " +
		"to their canonical form, creates a .bak backup, " +
		"and writes changes in-place.",
	RunE: runNormalize,
}

var (
	normalizePath   string
	normalizeDryRun bool
)

func init() {
	normalizeCmd.Flags().StringVar(
		&normalizePath, "path", ".",
		"path to career-ops root directory",
	)
	normalizeCmd.Flags().BoolVar(
		&normalizeDryRun, "dry-run", false,
		"show changes without writing",
	)
}

func runNormalize(_ *cobra.Command, _ []string) error {
	careerOpsPath := normalizePath
	dryRun := normalizeDryRun

	// Initialize states from YAML (falls back to defaults).
	states.Init(careerOpsPath)

	// Locate applications.md.
	filePath, err := tracker.FindAppsFile(careerOpsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No applications.md found. Nothing to normalize.")
			return nil
		}
		return oops.Wrapf(err, "finding applications file")
	}

	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return oops.Wrapf(err, "reading %s", filePath)
	}

	lines := strings.Split(string(content), "\n")
	changes, unknowns := processNormalization(lines)

	if len(unknowns) > 0 {
		fmt.Printf("\nWarning: %d unknown statuses:\n", len(unknowns))
		lo.ForEach(unknowns, func(u unknownStatus, _ int) {
			fmt.Printf("  #%d (line %d): %q\n", u.num, u.line, u.raw)
		})
	}

	fmt.Printf("\n%d statuses normalized\n", changes)

	if dryRun {
		fmt.Println("(dry-run -- no changes written)")
		return nil
	}

	if changes == 0 {
		fmt.Println("No changes needed")
		return nil
	}

	output := []byte(strings.Join(lines, "\n"))
	if err := tracker.BackupAndWrite(filePath, output); err != nil {
		return oops.Wrapf(err, "writing %s", filePath)
	}
	fmt.Printf("Written to %s (backup: %s.bak)\n", filePath, filePath)
	return nil
}

// processNormalization iterates over lines, normalizing statuses in-place.
// Returns the count of changes made and any unknown statuses encountered.
func processNormalization(
	lines []string,
) (int, []unknownStatus) {
	type result struct {
		unk     *unknownStatus
		changed bool
	}
	results := lo.Map(lines, func(line string, i int) result {
		changed, unk := normalizeAppLine(lines, i, line)
		return result{changed: changed, unk: unk}
	})
	changes := lo.CountBy(results, func(r result) bool { return r.changed })
	unknowns := lo.FilterMap(results, func(r result, _ int) (unknownStatus, bool) {
		if r.unk != nil {
			return *r.unk, true
		}
		return unknownStatus{}, false
	})
	return changes, unknowns
}

// parseDataRow extracts parts and entry number from a table data row.
// Returns nil parts if the line is not a valid data row.
func parseDataRow(line string) (parts []string, num int) {
	if !strings.HasPrefix(line, "|") {
		return nil, 0
	}

	parts = strings.Split(line, "|")
	if len(parts) < 9 {
		return nil, 0
	}

	// Skip header and separator rows.
	col1 := strings.TrimSpace(parts[1])
	if col1 == "#" || strings.HasPrefix(col1, "---") || col1 == "" {
		return nil, 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(col1))
	if err != nil {
		return nil, 0
	}
	return parts, n
}

// cleanFields handles special status patterns and bold score cleanup.
func cleanFields(parts []string, rawStatus string) {
	// Handle special patterns: move duplicado/repost info to notes.
	lowerRaw := strings.ToLower(strings.TrimSpace(rawStatus))
	if strings.HasPrefix(lowerRaw, "duplicado") ||
		strings.HasPrefix(lowerRaw, "dup") ||
		strings.HasPrefix(lowerRaw, "repost") {
		moveToNotes(parts, rawStatus)
	}

	// Strip markdown bold from score field (parts[5]).
	if strings.Contains(parts[5], "**") {
		score := strings.ReplaceAll(
			strings.TrimSpace(parts[5]), "**", "",
		)
		parts[5] = " " + score + " "
	}
}

// normalizeAppLine processes a single line, returning whether it was changed
// and any unknown status found. Modifies lines[idx] in-place if normalized.
func normalizeAppLine(
	lines []string, idx int, line string,
) (bool, *unknownStatus) {
	parts, num := parseDataRow(line)
	if parts == nil {
		return false, nil
	}

	rawStatus := strings.TrimSpace(parts[6])
	normalizedID := states.Normalize(rawStatus)

	if !states.IsCanonical(normalizedID) {
		return false, &unknownStatus{
			num: num, raw: rawStatus, line: idx + 1,
		}
	}

	label := states.Label(normalizedID)
	if label == rawStatus {
		return false, nil
	}

	oldStatus := rawStatus
	parts[6] = " " + label + " "

	cleanFields(parts, rawStatus)

	lines[idx] = strings.Join(parts, "|")

	fmt.Printf("#%d: %q -> %q\n", num, oldStatus, label)
	return true, nil
}

// moveToNotes appends the raw status text to the notes column if not already present.
func moveToNotes(parts []string, rawStatus string) {
	if len(parts) < 10 {
		return
	}
	raw := strings.TrimSpace(rawStatus)
	existing := strings.TrimSpace(parts[9])
	if existing == "" {
		parts[9] = " " + raw + " "
	} else if !strings.Contains(existing, raw) {
		parts[9] = " " + raw + ". " + existing + " "
	}
}

type unknownStatus struct {
	raw  string
	num  int
	line int
}
