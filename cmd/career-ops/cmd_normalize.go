package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
	"github.com/spf13/cobra"
)

var normalizeCmd = &cobra.Command{
	Use:   "normalize",
	Short: "Normalize status aliases to canonical statuses",
	Long:  "Reads applications.md, maps non-canonical statuses to their canonical form, creates a .bak backup, and writes changes in-place.",
	RunE:  runNormalize,
}

func init() {
	normalizeCmd.Flags().String("path", ".", "path to career-ops root directory")
	normalizeCmd.Flags().Bool("dry-run", false, "show changes without writing")
}

func runNormalize(cmd *cobra.Command, args []string) error {
	careerOpsPath, _ := cmd.Flags().GetString("path")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Initialize states from YAML (falls back to defaults).
	states.Init(careerOpsPath)

	// Locate applications.md.
	filePath, err := tracker.FindAppsFile(careerOpsPath)
	if err != nil {
		fmt.Println("No applications.md found. Nothing to normalize.")
		return nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	changes := 0
	var unknowns []unknownStatus

	for i, line := range lines {
		if !strings.HasPrefix(line, "|") {
			continue
		}

		parts := strings.Split(line, "|")
		// Expect: ['', '#', 'date', 'company', 'role', 'score', 'STATUS', 'pdf', 'report', 'notes', '']
		if len(parts) < 9 {
			continue
		}

		// Skip header and separator rows.
		col1 := strings.TrimSpace(parts[1])
		if col1 == "#" || strings.HasPrefix(col1, "---") || col1 == "" {
			continue
		}
		num, err := strconv.Atoi(strings.TrimSpace(col1))
		if err != nil {
			continue
		}

		rawStatus := strings.TrimSpace(parts[6])

		// Normalize via states package.
		normalizedID := states.Normalize(rawStatus)

		// Check if the result maps to a known canonical ID.
		if !states.IsCanonical(normalizedID) {
			unknowns = append(unknowns, unknownStatus{num: num, raw: rawStatus, line: i + 1})
			continue
		}

		label := states.Label(normalizedID)
		if label == rawStatus {
			continue // Already canonical, no change needed.
		}

		oldStatus := rawStatus
		parts[6] = " " + label + " "

		// Handle special patterns: move duplicado/repost info to notes.
		lowerRaw := strings.ToLower(strings.TrimSpace(rawStatus))
		if strings.HasPrefix(lowerRaw, "duplicado") || strings.HasPrefix(lowerRaw, "dup") || strings.HasPrefix(lowerRaw, "repost") {
			moveToNotes(parts, rawStatus)
		}

		// Strip markdown bold from score field (parts[5]).
		if strings.Contains(parts[5], "**") {
			parts[5] = " " + strings.ReplaceAll(strings.TrimSpace(parts[5]), "**", "") + " "
		}

		lines[i] = strings.Join(parts, "|")
		changes++

		fmt.Printf("#%d: %q -> %q\n", num, oldStatus, label)
	}

	if len(unknowns) > 0 {
		fmt.Printf("\nWarning: %d unknown statuses:\n", len(unknowns))
		for _, u := range unknowns {
			fmt.Printf("  #%d (line %d): %q\n", u.num, u.line, u.raw)
		}
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
		return fmt.Errorf("writing %s: %w", filePath, err)
	}
	fmt.Printf("Written to %s (backup: %s.bak)\n", filePath, filePath)
	return nil
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
	num  int
	raw  string
	line int
}
