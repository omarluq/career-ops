// Package tracker provides markdown table parsing, writing, and fuzzy matching
// for the career-ops application tracker (applications.md).
package tracker

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
)

var (
	reReportLink     = regexp.MustCompile(`\[(\d+)\]\(([^)]+)\)`)
	reScoreValue     = regexp.MustCompile(`(\d+\.?\d*)/5`)
	reArchetype      = regexp.MustCompile(`(?i)\*\*Arquetipo(?:\s+detectado)?\*\*\s*\|\s*(.+)`)
	reTlDr           = regexp.MustCompile(`(?i)\*\*TL;DR\*\*\s*\|\s*(.+)`)
	reTlDrColon      = regexp.MustCompile(`(?i)\*\*TL;DR:\*\*\s*(.+)`)
	reRemote         = regexp.MustCompile(`(?i)\*\*Remote\*\*\s*\|\s*(.+)`)
	reComp           = regexp.MustCompile(`(?i)\*\*Comp\*\*\s*\|\s*(.+)`)
	reArchetypeColon = regexp.MustCompile(`(?i)\*\*Arquetipo:\*\*\s*(.+)`)
	reReportURL      = regexp.MustCompile(`(?m)^\*\*URL:\*\*\s*(https?://\S+)`)
	reBatchID        = regexp.MustCompile(`(?m)^\*\*Batch ID:\*\*\s*(\d+)`)
)

// FindAppsFile locates applications.md, trying both root and data/ layouts.
func FindAppsFile(careerOpsPath string) (string, error) {
	paths := []string{
		filepath.Join(careerOpsPath, "data", "applications.md"),
		filepath.Join(careerOpsPath, "applications.md"),
	}
	found, ok := lo.Find(paths, func(p string) bool {
		_, err := os.Stat(p)
		return err == nil
	})
	if ok {
		return found, nil
	}
	return "", oops.Wrapf(nil,
		"applications.md not found in %s or %s/data/",
		careerOpsPath, careerOpsPath,
	)
}

// ParseApplications reads applications.md and returns parsed applications
// with URL enrichment.
func ParseApplications(
	careerOpsPath string,
) ([]model.CareerApplication, error) {
	filePath, err := FindAppsFile(careerOpsPath)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, oops.Wrapf(err, "reading %s", filePath)
	}

	lines := strings.Split(string(content), "\n")
	num := 0
	apps := lo.FilterMap(lines, func(line string, _ int) (model.CareerApplication, bool) {
		return parseLine(line, &num)
	})

	// Enrich with job URLs
	enrichURLs(careerOpsPath, apps)

	return apps, nil
}

// isTableDataLine checks whether a line is a data row in the markdown table.
func isTableDataLine(line string) bool {
	if line == "" ||
		strings.HasPrefix(line, "# ") ||
		strings.HasPrefix(line, "|---") ||
		strings.HasPrefix(line, "| #") {
		return false
	}
	return strings.HasPrefix(line, "|")
}

// parseLine parses a single markdown table line into an application.
// Returns false if the line should be skipped.
func parseLine(
	line string,
	num *int,
) (model.CareerApplication, bool) {
	line = strings.TrimSpace(line)
	if !isTableDataLine(line) {
		return model.CareerApplication{}, false
	}

	fields := splitTableLine(line)
	if len(fields) < 8 {
		return model.CareerApplication{}, false
	}

	*num++
	app := model.CareerApplication{
		Number:  *num,
		Date:    fields[1],
		Company: fields[2],
		Role:    fields[3],
		Status:  fields[5],
		HasPDF:  strings.Contains(fields[6], "\u2705"),
	}

	parseScoreField(&app, fields[4])
	parseReportField(&app, fields[7])

	if len(fields) > 8 {
		app.Notes = fields[8]
	}

	return app, true
}

// parseScoreField extracts the numeric score from a table field like "4.2/5".
func parseScoreField(app *model.CareerApplication, field string) {
	app.ScoreRaw = field
	if sm := reScoreValue.FindStringSubmatch(field); sm != nil {
		v, err := strconv.ParseFloat(sm[1], 64)
		if err == nil {
			app.Score = v
		}
	}
}

// parseReportField extracts report number and path from a markdown link field.
func parseReportField(app *model.CareerApplication, field string) {
	if rm := reReportLink.FindStringSubmatch(field); rm != nil {
		app.ReportNumber = rm[1]
		app.ReportPath = rm[2]
	}
}

// ParseAppLine parses a single markdown table line into an application.
func ParseAppLine(line string) (*model.CareerApplication, error) {
	fields := splitTableLine(line)
	if len(fields) < 9 {
		return nil, oops.Wrapf(nil, "too few fields: %d", len(fields))
	}
	num, err := strconv.Atoi(fields[1])
	if err != nil || num == 0 {
		return nil, oops.Wrapf(err, "invalid entry number: %s", fields[1])
	}
	app := &model.CareerApplication{
		Number:  num,
		Date:    fields[2],
		Company: fields[3],
		Role:    fields[4],
	}
	app.ScoreRaw = fields[5]
	if sm := reScoreValue.FindStringSubmatch(fields[5]); sm != nil {
		v, err := strconv.ParseFloat(sm[1], 64)
		if err == nil {
			app.Score = v
		}
	}
	app.Status = fields[6]
	app.HasPDF = strings.Contains(fields[7], "\u2705")
	if rm := reReportLink.FindStringSubmatch(fields[8]); rm != nil {
		app.ReportNumber = rm[1]
		app.ReportPath = rm[2]
	}
	if len(fields) > 9 {
		app.Notes = fields[9]
	}
	return app, nil
}

// splitTableLine splits a markdown table line into trimmed fields.
// Handles both pure pipe and mixed pipe/tab formats.
func splitTableLine(line string) []string {
	if strings.Contains(line, "\t") {
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSpace(line)
		parts := strings.Split(line, "\t")
		return lo.Map(parts, func(p string, _ int) string {
			return strings.TrimSpace(strings.Trim(p, "|"))
		})
	}
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	return lo.Map(parts, func(p string, _ int) string {
		return strings.TrimSpace(p)
	})
}

// LoadReportSummary extracts key fields from a report file header.
func LoadReportSummary(
	careerOpsPath, reportPath string,
) (archetype, tldr, remote, comp string) {
	fullPath := filepath.Join(careerOpsPath, reportPath)
	content, err := os.ReadFile(filepath.Clean(fullPath))
	if err != nil {
		return archetype, tldr, remote, comp
	}
	text := string(content)

	if m := reArchetype.FindStringSubmatch(text); m != nil {
		archetype = cleanTableCell(m[1])
	} else if m := reArchetypeColon.FindStringSubmatch(text); m != nil {
		archetype = cleanTableCell(m[1])
	}

	if m := reTlDr.FindStringSubmatch(text); m != nil {
		tldr = cleanTableCell(m[1])
	} else if m := reTlDrColon.FindStringSubmatch(text); m != nil {
		tldr = cleanTableCell(m[1])
	}

	if m := reRemote.FindStringSubmatch(text); m != nil {
		remote = cleanTableCell(m[1])
	}
	if m := reComp.FindStringSubmatch(text); m != nil {
		comp = cleanTableCell(m[1])
	}

	if len(tldr) > 120 {
		tldr = tldr[:117] + "..."
	}
	return archetype, tldr, remote, comp
}

// ComputeMetrics calculates aggregate metrics from applications.
func ComputeMetrics(apps []model.CareerApplication) model.PipelineMetrics {
	nonActionable := []string{states.StatusSkip, states.StatusRejected, states.StatusDiscarded}

	statuses := lo.Map(apps, func(app model.CareerApplication, _ int) string {
		return normalizeForMetrics(app.Status)
	})
	byStatus := lo.CountValues(statuses)

	scoredApps := lo.Filter(apps, func(app model.CareerApplication, _ int) bool {
		return app.Score > 0
	})
	totalScore := lo.SumBy(scoredApps, func(app model.CareerApplication) float64 {
		return app.Score
	})
	topScore := lo.Reduce(scoredApps, func(acc float64, app model.CareerApplication, _ int) float64 {
		return lo.Ternary(app.Score > acc, app.Score, acc)
	}, 0.0)

	m := model.PipelineMetrics{
		Total:      len(apps),
		ByStatus:   byStatus,
		TopScore:   topScore,
		WithPDF:    lo.CountBy(apps, func(app model.CareerApplication) bool { return app.HasPDF }),
		Actionable: lo.CountBy(statuses, func(s string) bool { return !lo.Contains(nonActionable, s) }),
	}

	if len(scoredApps) > 0 {
		m.AvgScore = totalScore / float64(len(scoredApps))
	}
	return m
}

// metricsNormMap maps raw status substrings to canonical status IDs
// for the normalizeForMetrics function.
var metricsNormMap = []struct {
	id      string
	substrs []string
	exact   []string
}{
	{
		substrs: []string{"no aplicar", "no_aplicar"},
		exact:   []string{states.StatusSkip},
		id:      states.StatusSkip,
	},
	{
		substrs: []string{"geo blocker"},
		id:      states.StatusSkip,
	},
	{
		substrs: []string{"interview", "entrevista"},
		id:      states.StatusInterview,
	},
	{
		substrs: nil,
		exact:   []string{states.StatusOffer},
		id:      states.StatusOffer,
	},
	{
		substrs: []string{"oferta"},
		id:      states.StatusOffer,
	},
	{
		substrs: []string{"responded", "respondido"},
		id:      states.StatusResponded,
	},
	{
		substrs: []string{"applied", "aplicado"},
		exact:   []string{"enviada", "aplicada", "sent"},
		id:      states.StatusApplied,
	},
	{
		substrs: []string{"rejected", "rechazado"},
		exact:   []string{"rechazada"},
		id:      states.StatusRejected,
	},
	{
		substrs: []string{"discarded", "descartado"},
		exact: []string{
			"descartada", "cerrada", "cancelada",
		},
		id: states.StatusDiscarded,
	},
	{
		substrs: []string{"evaluated", "evaluada"},
		exact: []string{
			"conditional", "hold", "monitor",
			"evaluar", "verificar",
		},
		id: states.StatusEvaluated,
	},
}

// dupPrefixes are prefixes that map to "discarded" in metrics normalization.
var dupPrefixes = []string{"duplicado", "dup"}

// normalizeForMetrics normalizes status for metrics computation.
// Uses inline logic to avoid circular import with states package.
func normalizeForMetrics(raw string) string {
	s := strings.ReplaceAll(raw, "**", "")
	s = strings.TrimSpace(strings.ToLower(s))
	if idx := strings.Index(s, " 202"); idx > 0 {
		s = strings.TrimSpace(s[:idx])
	}

	// Check dup prefixes
	if lo.SomeBy(dupPrefixes, func(prefix string) bool {
		return strings.HasPrefix(s, prefix)
	}) {
		return states.StatusDiscarded
	}

	// Check map entries
	match, found := lo.Find(metricsNormMap, func(entry struct {
		id      string
		substrs []string
		exact   []string
	}) bool {
		return lo.Contains(entry.exact, s) ||
			lo.SomeBy(entry.substrs, func(sub string) bool {
				return strings.Contains(s, sub)
			})
	})
	if found {
		return match.id
	}

	return s
}

func cleanTableCell(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "|")
	return strings.TrimSpace(s)
}

// --- URL Enrichment (5-tier strategy) ---

func enrichURLs(careerOpsPath string, apps []model.CareerApplication) {
	batchURLs := loadBatchInputURLs(careerOpsPath)
	reportNumURLs := loadJobURLs(careerOpsPath)

	lo.ForEach(lo.Range(len(apps)), func(i int, _ int) {
		enrichAppFromReport(
			careerOpsPath, &apps[i],
			batchURLs, reportNumURLs,
		)
	})

	// Strategy 4: scan-history.tsv
	enrichFromScanHistory(careerOpsPath, apps)
	// Strategy 5: company name fallback
	enrichAppURLsByCompany(careerOpsPath, apps)
}

// enrichAppFromReport tries strategies 1-3 for a single app.
func enrichAppFromReport(
	careerOpsPath string,
	app *model.CareerApplication,
	batchURLs, reportNumURLs map[string]string,
) {
	if app.ReportPath == "" {
		return
	}
	fullReport := filepath.Join(careerOpsPath, app.ReportPath)
	reportContent, err := os.ReadFile(filepath.Clean(fullReport))
	if err != nil {
		return
	}
	header := string(reportContent)
	if len(header) > 1000 {
		header = header[:1000]
	}

	// Strategy 1: **URL:** in report
	if m := reReportURL.FindStringSubmatch(header); m != nil {
		app.JobURL = m[1]
		return
	}
	// Strategy 2: **Batch ID:** -> batch-input.tsv
	if m := reBatchID.FindStringSubmatch(header); m != nil {
		if url, ok := batchURLs[m[1]]; ok {
			app.JobURL = url
			return
		}
	}
	// Strategy 3: report_num -> batch-state completed mapping
	if reportNumURLs != nil {
		if url, ok := reportNumURLs[app.ReportNumber]; ok {
			app.JobURL = url
		}
	}
}

func loadBatchInputURLs(careerOpsPath string) map[string]string {
	inputPath := filepath.Join(
		careerOpsPath, "batch", "batch-input.tsv",
	)
	inputData, err := os.ReadFile(filepath.Clean(inputPath))
	if err != nil {
		return nil
	}
	lines := strings.Split(string(inputData), "\n")
	pairs := lo.FilterMap(lines, func(line string, _ int) (lo.Tuple2[string, string], bool) {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			return lo.Tuple2[string, string]{}, false
		}
		id := fields[0]
		notes := fields[3]
		if idx := strings.LastIndex(notes, "| "); idx >= 0 {
			u := strings.TrimSpace(notes[idx+2:])
			if strings.HasPrefix(u, "http") {
				return lo.T2(id, u), true
			}
		}
		if strings.HasPrefix(fields[1], "http") {
			return lo.T2(id, fields[1]), true
		}
		return lo.Tuple2[string, string]{}, false
	})
	return lo.Associate(pairs, func(t lo.Tuple2[string, string]) (string, string) {
		return t.A, t.B
	})
}

// batchEntry holds parsed batch input data for URL resolution.
type batchEntry struct {
	url     string
	company string
	role    string
}

func loadJobURLs(careerOpsPath string) map[string]string {
	entries := parseBatchInputEntries(careerOpsPath)
	if entries == nil {
		return nil
	}
	return matchBatchStateToReports(careerOpsPath, entries)
}

// parseBatchInputEntries reads batch-input.tsv and extracts URL/company/role.
func parseBatchInputEntries(
	careerOpsPath string,
) map[string]batchEntry {
	inputPath := filepath.Join(
		careerOpsPath, "batch", "batch-input.tsv",
	)
	inputData, err := os.ReadFile(filepath.Clean(inputPath))
	if err != nil {
		return nil
	}

	lines := strings.Split(string(inputData), "\n")
	pairs := lo.FilterMap(lines, func(line string, _ int) (lo.Tuple2[string, batchEntry], bool) {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			return lo.Tuple2[string, batchEntry]{}, false
		}
		e := parseBatchInputLine(fields)
		if e.url == "" {
			return lo.Tuple2[string, batchEntry]{}, false
		}
		return lo.T2(fields[0], e), true
	})
	return lo.Associate(pairs, func(t lo.Tuple2[string, batchEntry]) (string, batchEntry) {
		return t.A, t.B
	})
}

// parseBatchInputLine extracts a batchEntry from TSV fields.
func parseBatchInputLine(fields []string) batchEntry {
	e := batchEntry{}
	notes := fields[3]
	if idx := strings.LastIndex(notes, "| "); idx >= 0 {
		u := strings.TrimSpace(notes[idx+2:])
		if strings.HasPrefix(u, "http") {
			e.url = u
		}
	}
	if e.url == "" && strings.HasPrefix(fields[1], "http") {
		e.url = fields[1]
	}
	notesPart := notes
	if pipeIdx := strings.Index(notesPart, " | "); pipeIdx >= 0 {
		notesPart = notesPart[:pipeIdx]
	}
	if atIdx := strings.LastIndex(notesPart, " @ "); atIdx >= 0 {
		e.role = strings.TrimSpace(notesPart[:atIdx])
		e.company = strings.TrimSpace(notesPart[atIdx+3:])
	}
	return e
}

// matchBatchStateToReports reads batch-state.tsv and maps report numbers
// to URLs using the provided batch entries.
func matchBatchStateToReports(
	careerOpsPath string,
	entries map[string]batchEntry,
) map[string]string {
	statePath := filepath.Join(
		careerOpsPath, "batch", "batch-state.tsv",
	)
	stateData, err := os.ReadFile(filepath.Clean(statePath))
	if err != nil {
		return nil
	}

	lines := strings.Split(string(stateData), "\n")
	pairs := lo.FlatMap(lines, func(line string, _ int) []lo.Tuple2[string, string] {
		fields := strings.Split(line, "\t")
		if len(fields) < 6 || fields[0] == "id" {
			return nil
		}
		id, status, reportNum := fields[0], fields[2], fields[5]
		if status != "completed" || reportNum == "" || reportNum == "-" {
			return nil
		}
		e, ok := entries[id]
		if !ok {
			return nil
		}
		result := []lo.Tuple2[string, string]{lo.T2(reportNum, e.url)}
		if len(reportNum) < 3 {
			result = append(result, lo.T2(fmt.Sprintf("%03s", reportNum), e.url))
		}
		return result
	})
	return lo.Associate(pairs, func(t lo.Tuple2[string, string]) (string, string) {
		return t.A, t.B
	})
}

type scanEntry struct {
	url, company, title string
}

// companyURLEntry holds a role/URL pair for company-based URL matching.
type companyURLEntry struct {
	role, url string
}

// enrichByCompanyLookup is a generic enrichment strategy that loads entries grouped
// by company, then matches each app's company to find a URL. The getURL function
// extracts a URL from a single match; bestMatch selects the best URL from multiple matches.
func enrichByCompanyLookup[T any](
	apps []model.CareerApplication,
	lookup map[string][]T,
	getURL func(T) string,
	bestMatch func(appRole string, matches []T) string,
) {
	if lookup == nil {
		return
	}
	lo.ForEach(lo.Range(len(apps)), func(i int, _ int) {
		if apps[i].JobURL != "" {
			return
		}
		key := NormalizeCompany(apps[i].Company)
		matches := lookup[key]
		switch {
		case len(matches) == 1:
			apps[i].JobURL = getURL(matches[0])
		case len(matches) > 1:
			apps[i].JobURL = bestMatch(apps[i].Role, matches)
		}
	})
}

func enrichFromScanHistory(
	careerOpsPath string,
	apps []model.CareerApplication,
) {
	byCompany := loadScanHistoryByCompany(careerOpsPath)
	enrichByCompanyLookup(apps, byCompany,
		func(e scanEntry) string { return e.url },
		bestScanMatch,
	)
}

// loadScanHistoryByCompany reads scan-history.tsv and groups entries
// by normalized company name.
func loadScanHistoryByCompany(
	careerOpsPath string,
) map[string][]scanEntry {
	scanPath := filepath.Join(careerOpsPath, "scan-history.tsv")
	scanData, err := os.ReadFile(filepath.Clean(scanPath))
	if err != nil {
		return nil
	}
	lines := strings.Split(string(scanData), "\n")
	entries := lo.FilterMap(lines, func(line string, _ int) (scanEntry, bool) {
		fields := strings.Split(line, "\t")
		if len(fields) < 5 || fields[0] == "url" {
			return scanEntry{}, false
		}
		u := fields[0]
		if u == "" || !strings.HasPrefix(u, "http") {
			return scanEntry{}, false
		}
		return scanEntry{url: u, company: fields[4], title: fields[3]}, true
	})
	return lo.GroupBy(entries, func(e scanEntry) string {
		return NormalizeCompany(e.company)
	})
}

func enrichAppURLsByCompany(
	careerOpsPath string,
	apps []model.CareerApplication,
) {
	byCompany := loadBatchInputByCompany(careerOpsPath)
	enrichByCompanyLookup(apps, byCompany,
		func(e companyURLEntry) string { return e.url },
		bestCompanyMatch,
	)
}

// loadBatchInputByCompany reads batch-input.tsv and groups entries
// by normalized company name.
func loadBatchInputByCompany(
	careerOpsPath string,
) map[string][]companyURLEntry {
	inputPath := filepath.Join(
		careerOpsPath, "batch", "batch-input.tsv",
	)
	inputData, err := os.ReadFile(filepath.Clean(inputPath))
	if err != nil {
		return nil
	}

	lines := strings.Split(string(inputData), "\n")
	type keyedEntry struct {
		key   string
		entry companyURLEntry
	}
	keyed := lo.FilterMap(lines, func(line string, _ int) (keyedEntry, bool) {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			return keyedEntry{}, false
		}
		e := parseBatchInputLine(fields)
		if e.url == "" || e.company == "" {
			return keyedEntry{}, false
		}
		return keyedEntry{
			key:   NormalizeCompany(e.company),
			entry: companyURLEntry{role: e.role, url: e.url},
		}, true
	})
	grouped := lo.GroupBy(keyed, func(k keyedEntry) string { return k.key })
	return lo.MapValues(grouped, func(entries []keyedEntry, _ string) []companyURLEntry {
		return lo.Map(entries, func(k keyedEntry, _ int) companyURLEntry { return k.entry })
	})
}

// bestCompanyMatch finds the best URL match by role word overlap.
func bestCompanyMatch(
	appRole string,
	matches []companyURLEntry,
) string {
	appRoleLower := strings.ToLower(appRole)
	best := lo.MaxBy(matches, func(a, b companyURLEntry) bool {
		return wordOverlapScore(appRoleLower, strings.ToLower(a.role)) >
			wordOverlapScore(appRoleLower, strings.ToLower(b.role))
	})
	return best.url
}

func bestScanMatch(appRole string, matches []scanEntry) string {
	role := strings.ToLower(appRole)
	best := lo.MaxBy(matches, func(a, b scanEntry) bool {
		return wordOverlapScore(role, strings.ToLower(a.title)) >
			wordOverlapScore(role, strings.ToLower(b.title))
	})
	return best.url
}

// wordOverlapScore counts how many significant words (len>2) from
// source appear in target.
func wordOverlapScore(source, target string) int {
	return lo.CountBy(strings.Fields(source), func(word string) bool {
		return len(word) > 2 && strings.Contains(target, word)
	})
}
