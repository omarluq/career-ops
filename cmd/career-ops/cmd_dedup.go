package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
)

var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Remove duplicate tracker entries",
	Long: `Groups entries by normalized company key, checks role matches within groups,
keeps the entry with the highest score, promotes status if a discarded duplicate
had a more advanced pipeline status, and merges notes. Creates a .bak backup before writing.`,
	RunE: runDedup,
}

func init() {
	dedupCmd.Flags().String("path", ".", "path to career-ops root directory")
	dedupCmd.Flags().Bool("dry-run", false, "preview changes without writing")
}

// parsedEntry holds a parsed application line together with its original line index.
type parsedEntry struct {
	num     int
	date    string
	company string
	role    string
	score   string
	status  string
	pdf     string
	report  string
	notes   string
	lineIdx int
	raw     string
}

// parseTableLine splits a markdown table row into a parsedEntry.
// Returns nil if the line is not a valid data row.
func parseTableLine(line string, lineIdx int) *parsedEntry {
	parts := strings.Split(line, "|")
	trimmed := lo.Map(parts, func(s string, _ int) string { return strings.TrimSpace(s) })
	// Expect at least: empty | num | date | company | role | score | status | pdf | report | notes | empty
	if len(trimmed) < 10 {
		return nil
	}
	// trimmed[0] is empty (before first |), data starts at index 1
	numStr := trimmed[1]
	var num int
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil || num == 0 {
		return nil
	}
	return &parsedEntry{
		num:     num,
		date:    trimmed[2],
		company: trimmed[3],
		role:    trimmed[4],
		score:   trimmed[5],
		status:  trimmed[6],
		pdf:     trimmed[7],
		report:  trimmed[8],
		notes:   trimmed[9],
		lineIdx: lineIdx,
		raw:     line,
	}
}

// reformatLine rebuilds a markdown table row from the entry's fields.
func (e *parsedEntry) reformatLine() string {
	return fmt.Sprintf("| %d | %s | %s | %s | %s | %s | %s | %s | %s |",
		e.num, e.date, e.company, e.role, e.score, e.status, e.pdf, e.report, e.notes)
}

func runDedup(cmd *cobra.Command, _ []string) error {
	careerOpsPath, _ := cmd.Flags().GetString("path")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	states.Init(careerOpsPath)

	appsFile, err := tracker.FindAppsFile(careerOpsPath)
	if err != nil {
		fmt.Println("No applications.md found. Nothing to dedup.")
		return nil
	}

	content, err := os.ReadFile(appsFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", appsFile, err)
	}
	lines := strings.Split(string(content), "\n")

	// Parse all data rows, recording their line indices.
	var entries []*parsedEntry
	entryByNum := make(map[int]*parsedEntry)

	for i, line := range lines {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		e := parseTableLine(line, i)
		if e != nil {
			entries = append(entries, e)
			entryByNum[e.num] = e
		}
	}

	fmt.Printf("  %d entries loaded\n", len(entries))

	// Group entries by normalized company key.
	groups := lo.GroupBy(entries, func(e *parsedEntry) string {
		return tracker.NormalizeCompanyKey(e.company)
	})

	// Track which line indices to remove.
	linesToRemove := make(map[int]bool)
	removed := 0

	for _, companyEntries := range groups {
		if len(companyEntries) < 2 {
			continue
		}

		// Within the same company, cluster by role match.
		processed := make(map[int]bool) // index into companyEntries
		for i := 0; i < len(companyEntries); i++ {
			if processed[i] {
				continue
			}
			cluster := []*parsedEntry{companyEntries[i]}
			processed[i] = true

			for j := i + 1; j < len(companyEntries); j++ {
				if processed[j] {
					continue
				}
				if tracker.RoleMatch(companyEntries[i].role, companyEntries[j].role) {
					cluster = append(cluster, companyEntries[j])
					processed[j] = true
				}
			}

			if len(cluster) < 2 {
				continue
			}

			// Sort by score descending -- keeper is the highest-scored entry.
			sort.Slice(cluster, func(a, b int) bool {
				return tracker.ParseScore(cluster[a].score) > tracker.ParseScore(cluster[b].score)
			})
			keeper := cluster[0]

			// Check if any duplicate has a more advanced pipeline status.
			bestRank := states.StatusRank(keeper.status)
			bestStatus := keeper.status
			for _, dup := range cluster[1:] {
				rank := states.StatusRank(dup.status)
				if rank > bestRank {
					bestRank = rank
					bestStatus = dup.status
				}
			}

			// Promote keeper's status if a removed entry was further along.
			if bestStatus != keeper.status {
				promoter := lo.Filter(cluster[1:], func(e *parsedEntry, _ int) bool {
					return e.status == bestStatus
				})
				promoterNum := 0
				if len(promoter) > 0 {
					promoterNum = promoter[0].num
				}
				keeper.status = bestStatus
				lines[keeper.lineIdx] = keeper.reformatLine()
				fmt.Printf("  #%d: status promoted to %q (from #%d)\n", keeper.num, bestStatus, promoterNum)
			}

			// Mark duplicates for removal.
			for _, dup := range cluster[1:] {
				linesToRemove[dup.lineIdx] = true
				removed++
				fmt.Printf("  Remove #%d (%s -- %s, %s) -> kept #%d (%s)\n",
					dup.num, dup.company, dup.role, dup.score, keeper.num, keeper.score)
			}
		}
	}

	// Remove lines in reverse order to preserve indices.
	indices := lo.Keys(linesToRemove)
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	for _, idx := range indices {
		lines = append(lines[:idx], lines[idx+1:]...)
	}

	fmt.Printf("\n  %d duplicates removed\n", removed)

	if dryRun {
		fmt.Println("(dry-run -- no changes written)")
		return nil
	}

	if removed == 0 {
		fmt.Println("  No duplicates found")
		return nil
	}

	if err := tracker.BackupAndWrite(appsFile, []byte(strings.Join(lines, "\n"))); err != nil {
		return fmt.Errorf("writing %s: %w", appsFile, err)
	}
	fmt.Printf("  Written to %s (backup: %s.bak)\n", appsFile, appsFile)
	return nil
}
