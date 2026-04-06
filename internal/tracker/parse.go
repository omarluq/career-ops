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
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
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
	var apps []model.CareerApplication
	num := 0

	for _, line := range lines {
		app, ok := parseLine(line, &num)
		if !ok {
			continue
		}
		apps = append(apps, app)
	}

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
	m := model.PipelineMetrics{
		Total:    len(apps),
		ByStatus: make(map[string]int),
	}

	var totalScore float64
	var scored int

	for i := range apps {
		status := normalizeForMetrics(apps[i].Status)
		m.ByStatus[status]++

		if apps[i].Score > 0 {
			totalScore += apps[i].Score
			scored++
			if apps[i].Score > m.TopScore {
				m.TopScore = apps[i].Score
			}
		}
		if apps[i].HasPDF {
			m.WithPDF++
		}
		if !lo.Contains([]string{states.StatusSkip, states.StatusRejected, states.StatusDiscarded}, status) {
			m.Actionable++
		}
	}

	if scored > 0 {
		m.AvgScore = totalScore / float64(scored)
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
	for _, entry := range metricsNormMap {
		if lo.Contains(entry.exact, s) {
			return entry.id
		}
		if lo.SomeBy(entry.substrs, func(sub string) bool {
			return strings.Contains(s, sub)
		}) {
			return entry.id
		}
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

	for i := range apps {
		enrichAppFromReport(
			careerOpsPath, &apps[i],
			batchURLs, reportNumURLs,
		)
	}

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
	result := make(map[string]string)
	for _, line := range strings.Split(string(inputData), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			continue
		}
		id := fields[0]
		notes := fields[3]
		if idx := strings.LastIndex(notes, "| "); idx >= 0 {
			u := strings.TrimSpace(notes[idx+2:])
			if strings.HasPrefix(u, "http") {
				result[id] = u
				continue
			}
		}
		if strings.HasPrefix(fields[1], "http") {
			result[id] = fields[1]
		}
	}
	return result
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

	entries := make(map[string]batchEntry)
	for _, line := range strings.Split(string(inputData), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			continue
		}
		e := parseBatchInputLine(fields)
		if e.url != "" {
			entries[fields[0]] = e
		}
	}
	return entries
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

	reportToURL := make(map[string]string)
	for _, line := range strings.Split(string(stateData), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 6 || fields[0] == "id" {
			continue
		}
		id := fields[0]
		status := fields[2]
		reportNum := fields[5]
		if status != "completed" || reportNum == "" || reportNum == "-" {
			continue
		}
		if e, ok := entries[id]; ok {
			reportToURL[reportNum] = e.url
			if len(reportNum) < 3 {
				reportToURL[fmt.Sprintf("%03s", reportNum)] = e.url
			}
		}
	}
	return reportToURL
}

type scanEntry struct {
	url, company, title string
}

func enrichFromScanHistory(
	careerOpsPath string,
	apps []model.CareerApplication,
) {
	byCompany := loadScanHistoryByCompany(careerOpsPath)
	if byCompany == nil {
		return
	}

	for i := range apps {
		if apps[i].JobURL != "" {
			continue
		}
		key := NormalizeCompany(apps[i].Company)
		matches := byCompany[key]
		if len(matches) == 1 {
			apps[i].JobURL = matches[0].url
		} else if len(matches) > 1 {
			apps[i].JobURL = bestScanMatch(apps[i].Role, matches)
		}
	}
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
	byCompany := make(map[string][]scanEntry)
	for _, line := range strings.Split(string(scanData), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 5 || fields[0] == "url" {
			continue
		}
		url := fields[0]
		if url == "" || !strings.HasPrefix(url, "http") {
			continue
		}
		key := NormalizeCompany(fields[4])
		byCompany[key] = append(byCompany[key], scanEntry{
			url: url, company: fields[4], title: fields[3],
		})
	}
	return byCompany
}

// companyURLEntry holds a role/URL pair for company-based URL matching.
type companyURLEntry struct {
	role, url string
}

func enrichAppURLsByCompany(
	careerOpsPath string,
	apps []model.CareerApplication,
) {
	byCompany := loadBatchInputByCompany(careerOpsPath)
	if byCompany == nil {
		return
	}

	for i := range apps {
		if apps[i].JobURL != "" {
			continue
		}
		key := NormalizeCompany(apps[i].Company)
		matches := byCompany[key]
		if len(matches) == 1 {
			apps[i].JobURL = matches[0].url
		} else if len(matches) > 1 {
			apps[i].JobURL = bestCompanyMatch(
				apps[i].Role, matches,
			)
		}
	}
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

	byCompany := make(map[string][]companyURLEntry)
	for _, line := range strings.Split(string(inputData), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			continue
		}
		e := parseBatchInputLine(fields)
		if e.url == "" || e.company == "" {
			continue
		}
		key := NormalizeCompany(e.company)
		byCompany[key] = append(byCompany[key], companyURLEntry{
			role: e.role, url: e.url,
		})
	}
	return byCompany
}

// bestCompanyMatch finds the best URL match by role word overlap.
func bestCompanyMatch(
	appRole string,
	matches []companyURLEntry,
) string {
	appRoleLower := strings.ToLower(appRole)
	best := matches[0].url
	bestScore := 0
	for _, m := range matches {
		score := wordOverlapScore(
			appRoleLower, strings.ToLower(m.role),
		)
		if score > bestScore {
			bestScore = score
			best = m.url
		}
	}
	return best
}

func bestScanMatch(appRole string, matches []scanEntry) string {
	role := strings.ToLower(appRole)
	best := matches[0].url
	bestScore := 0
	for _, m := range matches {
		score := wordOverlapScore(role, strings.ToLower(m.title))
		if score > bestScore {
			bestScore = score
			best = m.url
		}
	}
	return best
}

// wordOverlapScore counts how many significant words (len>2) from
// source appear in target.
func wordOverlapScore(source, target string) int {
	return lo.CountBy(strings.Fields(source), func(word string) bool {
		return len(word) > 2 && strings.Contains(target, word)
	})
}
