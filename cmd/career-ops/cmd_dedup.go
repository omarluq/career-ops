package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
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

var (
	dedupPath   string
	dedupDryRun bool
)

func init() {
	dedupCmd.Flags().StringVar(&dedupPath, "path", ".", "path to career-ops root directory")
	dedupCmd.Flags().BoolVar(&dedupDryRun, "dry-run", false, "preview changes without writing")
}

// parsedEntry holds a parsed application line together with its original line index.
type parsedEntry struct {
	date    string
	company string
	role    string
	score   string
	status  string
	pdf     string
	report  string
	notes   string
	raw     string
	num     int
	lineIdx int
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

func runDedup(_ *cobra.Command, _ []string) error {
	careerOpsPath := dedupPath

	states.Init(careerOpsPath)

	appsFile, err := tracker.FindAppsFile(careerOpsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No applications.md found. Nothing to dedup.")
			return nil
		}
		return oops.Wrapf(err, "finding applications file")
	}

	content, err := os.ReadFile(filepath.Clean(appsFile))
	if err != nil {
		return oops.Wrapf(err, "reading %s", appsFile)
	}
	lines := strings.Split(string(content), "\n")

	entries := parseEntries(lines)
	fmt.Printf("  %d entries loaded\n", len(entries))

	// Group entries by normalized company key.
	groups := lo.GroupBy(entries, func(e *parsedEntry) string {
		return tracker.NormalizeCompanyKey(e.company)
	})

	linesToRemove, removed := findDuplicateClusters(groups, lines)

	// Remove lines in reverse order to preserve indices.
	lines = removeLines(lines, linesToRemove)

	fmt.Printf("\n  %d duplicates removed\n", removed)

	if dedupDryRun {
		fmt.Println("(dry-run -- no changes written)")
		return nil
	}

	if removed == 0 {
		fmt.Println("  No duplicates found")
		return nil
	}

	if err := tracker.BackupAndWrite(appsFile, []byte(strings.Join(lines, "\n"))); err != nil {
		return oops.Wrapf(err, "writing %s", appsFile)
	}
	fmt.Printf("  Written to %s (backup: %s.bak)\n", appsFile, appsFile)
	return nil
}

// parseEntries scans all lines and returns parsed data rows.
func parseEntries(lines []string) []*parsedEntry {
	return lo.FilterMap(lo.Map(lines, lo.T2[string, int]), func(t lo.Tuple2[string, int], _ int) (*parsedEntry, bool) {
		line, i := t.A, t.B
		if !strings.HasPrefix(line, "|") {
			return nil, false
		}
		if e := parseTableLine(line, i); e != nil {
			return e, true
		}
		return nil, false
	})
}

// findDuplicateClusters identifies duplicate entries within each company group,
// promotes statuses where needed, and returns the set of line indices to remove
// along with the total removal count.
func findDuplicateClusters(
	groups map[string][]*parsedEntry,
	lines []string,
) (linesToRemove map[int]bool, removed int) {
	linesToRemove = make(map[int]bool)
	removed = 0

	lo.ForEach(lo.Values(groups), func(companyEntries []*parsedEntry, _ int) {
		if len(companyEntries) < 2 {
			return
		}

		clusters := buildRoleClusters(companyEntries)
		indices := lo.FlatMap(clusters, func(cluster []*parsedEntry, _ int) []int {
			return mergeCluster(cluster, lines)
		})
		lo.ForEach(indices, func(idx int, _ int) {
			linesToRemove[idx] = true
			removed++
		})
	})
	return linesToRemove, removed
}

// buildRoleClusters groups entries within a single company by role match.
func buildRoleClusters(companyEntries []*parsedEntry) [][]*parsedEntry {
	type clusterState struct {
		processed map[int]bool
		clusters  [][]*parsedEntry
	}
	result := lo.Reduce(lo.Range(len(companyEntries)), func(acc clusterState, i int, _ int) clusterState {
		if acc.processed[i] {
			return acc
		}
		acc.processed[i] = true
		// Find all unprocessed entries with matching roles.
		matches := lo.Filter(lo.Range(len(companyEntries)), func(j int, _ int) bool {
			return j > i && !acc.processed[j] &&
				tracker.RoleMatch(companyEntries[i].role, companyEntries[j].role)
		})
		lo.ForEach(matches, func(j int, _ int) { acc.processed[j] = true })
		cluster := append([]*parsedEntry{companyEntries[i]},
			lo.Map(matches, func(j int, _ int) *parsedEntry { return companyEntries[j] })...,
		)
		if len(cluster) >= 2 {
			acc.clusters = append(acc.clusters, cluster)
		}
		return acc
	}, clusterState{processed: make(map[int]bool)})
	return result.clusters
}

// mergeCluster processes a single duplicate cluster: picks the highest-scored keeper,
// promotes status if needed, and returns the line indices of duplicates to remove.
func mergeCluster(cluster []*parsedEntry, lines []string) []int {
	// Sort by score descending -- keeper is the highest-scored entry.
	sort.Slice(cluster, func(a, b int) bool {
		return tracker.ParseScore(cluster[a].score) > tracker.ParseScore(cluster[b].score)
	})
	keeper := cluster[0]

	// Check if any duplicate has a more advanced pipeline status.
	bestEntry := lo.MaxBy(cluster, func(a, b *parsedEntry) bool {
		return states.StatusRank(a.status) > states.StatusRank(b.status)
	})
	bestStatus := bestEntry.status

	// Promote keeper's status if a removed entry was further along.
	if bestStatus != keeper.status {
		promoter, found := lo.Find(cluster[1:], func(e *parsedEntry) bool {
			return e.status == bestStatus
		})
		promoterNum := 0
		if found {
			promoterNum = promoter.num
		}
		keeper.status = bestStatus
		lines[keeper.lineIdx] = keeper.reformatLine()
		fmt.Printf("  #%d: status promoted to %q (from #%d)\n", keeper.num, bestStatus, promoterNum)
	}

	// Mark duplicates for removal.
	return lo.Map(cluster[1:], func(dup *parsedEntry, _ int) int {
		fmt.Printf("  Remove #%d (%s -- %s, %s) -> kept #%d (%s)\n",
			dup.num, dup.company, dup.role, dup.score, keeper.num, keeper.score)
		return dup.lineIdx
	})
}

// removeLines removes the lines at the given indices, returning the filtered result.
func removeLines(lines []string, toRemove map[int]bool) []string {
	return lo.Filter(lines, func(_ string, i int) bool {
		return !toRemove[i]
	})
}
