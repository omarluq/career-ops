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

	"github.com/omarluq/career-ops/internal/model"
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
	return "", fmt.Errorf("applications.md not found in %s or %s/data/", careerOpsPath, careerOpsPath)
}

// ParseApplications reads applications.md and returns parsed applications with URL enrichment.
func ParseApplications(careerOpsPath string) ([]model.CareerApplication, error) {
	filePath, err := FindAppsFile(careerOpsPath)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	var apps []model.CareerApplication
	num := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "|---") || strings.HasPrefix(line, "| #") {
			continue
		}
		if !strings.HasPrefix(line, "|") {
			continue
		}

		fields := splitTableLine(line)
		if len(fields) < 8 {
			continue
		}

		num++
		app := model.CareerApplication{
			Number:  num,
			Date:    fields[1],
			Company: fields[2],
			Role:    fields[3],
			Status:  fields[5],
			HasPDF:  strings.Contains(fields[6], "\u2705"),
		}

		// Parse score
		app.ScoreRaw = fields[4]
		if sm := reScoreValue.FindStringSubmatch(fields[4]); sm != nil {
			app.Score, _ = strconv.ParseFloat(sm[1], 64)
		}

		// Parse report link
		if rm := reReportLink.FindStringSubmatch(fields[7]); rm != nil {
			app.ReportNumber = rm[1]
			app.ReportPath = rm[2]
		}

		// Notes
		if len(fields) > 8 {
			app.Notes = fields[8]
		}

		apps = append(apps, app)
	}

	// Enrich with job URLs
	enrichURLs(careerOpsPath, apps)

	return apps, nil
}

// ParseAppLine parses a single markdown table line into an application.
func ParseAppLine(line string) (*model.CareerApplication, error) {
	fields := splitTableLine(line)
	if len(fields) < 9 {
		return nil, fmt.Errorf("too few fields: %d", len(fields))
	}
	num, err := strconv.Atoi(fields[1])
	if err != nil || num == 0 {
		return nil, fmt.Errorf("invalid entry number: %s", fields[1])
	}
	app := &model.CareerApplication{
		Number: num,
		Date:   fields[2],
		Company: fields[3],
		Role:    fields[4],
	}
	app.ScoreRaw = fields[5]
	if sm := reScoreValue.FindStringSubmatch(fields[5]); sm != nil {
		app.Score, _ = strconv.ParseFloat(sm[1], 64)
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
		fields := make([]string, len(parts))
		for i, p := range parts {
			fields[i] = strings.TrimSpace(strings.Trim(p, "|"))
		}
		return fields
	}
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	fields := make([]string, len(parts))
	for i, p := range parts {
		fields[i] = strings.TrimSpace(p)
	}
	return fields
}

// LoadReportSummary extracts key fields from a report file header.
func LoadReportSummary(careerOpsPath, reportPath string) (archetype, tldr, remote, comp string) {
	fullPath := filepath.Join(careerOpsPath, reportPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return
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
	return
}

// ComputeMetrics calculates aggregate metrics from applications.
func ComputeMetrics(apps []model.CareerApplication) model.PipelineMetrics {
	// Import states inline to avoid circular dependency.
	// The states package must be initialized before calling this.
	m := model.PipelineMetrics{
		Total:    len(apps),
		ByStatus: make(map[string]int),
	}

	var totalScore float64
	var scored int

	for _, app := range apps {
		status := normalizeForMetrics(app.Status)
		m.ByStatus[status]++

		if app.Score > 0 {
			totalScore += app.Score
			scored++
			if app.Score > m.TopScore {
				m.TopScore = app.Score
			}
		}
		if app.HasPDF {
			m.WithPDF++
		}
		if status != "skip" && status != "rejected" && status != "discarded" {
			m.Actionable++
		}
	}

	if scored > 0 {
		m.AvgScore = totalScore / float64(scored)
	}
	return m
}

// normalizeForMetrics normalizes status for metrics computation.
// Uses inline logic to avoid circular import with states package.
func normalizeForMetrics(raw string) string {
	s := strings.ReplaceAll(raw, "**", "")
	s = strings.TrimSpace(strings.ToLower(s))
	if idx := strings.Index(s, " 202"); idx > 0 {
		s = strings.TrimSpace(s[:idx])
	}
	switch {
	case strings.Contains(s, "no aplicar") || strings.Contains(s, "no_aplicar") || s == "skip" || strings.Contains(s, "geo blocker"):
		return "skip"
	case strings.Contains(s, "interview") || strings.Contains(s, "entrevista"):
		return "interview"
	case s == "offer" || strings.Contains(s, "oferta"):
		return "offer"
	case strings.Contains(s, "responded") || strings.Contains(s, "respondido"):
		return "responded"
	case strings.Contains(s, "applied") || strings.Contains(s, "aplicado") || s == "enviada" || s == "aplicada" || s == "sent":
		return "applied"
	case strings.Contains(s, "rejected") || strings.Contains(s, "rechazado") || s == "rechazada":
		return "rejected"
	case strings.Contains(s, "discarded") || strings.Contains(s, "descartado") || s == "descartada" || s == "cerrada" || s == "cancelada" ||
		strings.HasPrefix(s, "duplicado") || strings.HasPrefix(s, "dup"):
		return "discarded"
	case strings.Contains(s, "evaluated") || strings.Contains(s, "evaluada") || s == "condicional" || s == "hold" || s == "monitor" || s == "evaluar" || s == "verificar":
		return "evaluated"
	default:
		return s
	}
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
		if apps[i].ReportPath == "" {
			continue
		}
		fullReport := filepath.Join(careerOpsPath, apps[i].ReportPath)
		reportContent, err := os.ReadFile(fullReport)
		if err != nil {
			continue
		}
		header := string(reportContent)
		if len(header) > 1000 {
			header = header[:1000]
		}

		// Strategy 1: **URL:** in report
		if m := reReportURL.FindStringSubmatch(header); m != nil {
			apps[i].JobURL = m[1]
			continue
		}
		// Strategy 2: **Batch ID:** -> batch-input.tsv
		if m := reBatchID.FindStringSubmatch(header); m != nil {
			if url, ok := batchURLs[m[1]]; ok {
				apps[i].JobURL = url
				continue
			}
		}
		// Strategy 3: report_num -> batch-state completed mapping
		if reportNumURLs != nil {
			if url, ok := reportNumURLs[apps[i].ReportNumber]; ok {
				apps[i].JobURL = url
				continue
			}
		}
	}

	// Strategy 4: scan-history.tsv
	enrichFromScanHistory(careerOpsPath, apps)
	// Strategy 5: company name fallback
	enrichAppURLsByCompany(careerOpsPath, apps)
}

func loadBatchInputURLs(careerOpsPath string) map[string]string {
	inputPath := filepath.Join(careerOpsPath, "batch", "batch-input.tsv")
	inputData, err := os.ReadFile(inputPath)
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

func loadJobURLs(careerOpsPath string) map[string]string {
	inputPath := filepath.Join(careerOpsPath, "batch", "batch-input.tsv")
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return nil
	}

	type batchEntry struct {
		url     string
		company string
		role    string
	}
	entries := make(map[string]batchEntry)
	for _, line := range strings.Split(string(inputData), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			continue
		}
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
		if e.url != "" {
			entries[fields[0]] = e
		}
	}

	statePath := filepath.Join(careerOpsPath, "batch", "batch-state.tsv")
	stateData, err := os.ReadFile(statePath)
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

func enrichFromScanHistory(careerOpsPath string, apps []model.CareerApplication) {
	scanPath := filepath.Join(careerOpsPath, "scan-history.tsv")
	scanData, err := os.ReadFile(scanPath)
	if err != nil {
		return
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
		byCompany[key] = append(byCompany[key], scanEntry{url: url, company: fields[4], title: fields[3]})
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

func enrichAppURLsByCompany(careerOpsPath string, apps []model.CareerApplication) {
	inputPath := filepath.Join(careerOpsPath, "batch", "batch-input.tsv")
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return
	}

	type entry struct {
		role, url string
	}
	byCompany := make(map[string][]entry)
	for _, line := range strings.Split(string(inputData), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 || fields[0] == "id" {
			continue
		}
		notes := fields[3]
		var url string
		if idx := strings.LastIndex(notes, "| "); idx >= 0 {
			u := strings.TrimSpace(notes[idx+2:])
			if strings.HasPrefix(u, "http") {
				url = u
			}
		}
		if url == "" && strings.HasPrefix(fields[1], "http") {
			url = fields[1]
		}
		if url == "" {
			continue
		}
		notesPart := notes
		if pipeIdx := strings.Index(notesPart, " | "); pipeIdx >= 0 {
			notesPart = notesPart[:pipeIdx]
		}
		if atIdx := strings.LastIndex(notesPart, " @ "); atIdx >= 0 {
			role := strings.TrimSpace(notesPart[:atIdx])
			company := strings.TrimSpace(notesPart[atIdx+3:])
			key := NormalizeCompany(company)
			byCompany[key] = append(byCompany[key], entry{role: role, url: url})
		}
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
			appRole := strings.ToLower(apps[i].Role)
			best := matches[0].url
			bestScore := 0
			for _, m := range matches {
				score := 0
				mRole := strings.ToLower(m.role)
				for _, word := range strings.Fields(appRole) {
					if len(word) > 2 && strings.Contains(mRole, word) {
						score++
					}
				}
				if score > bestScore {
					bestScore = score
					best = m.url
				}
			}
			apps[i].JobURL = best
		}
	}
}

func bestScanMatch(appRole string, matches []scanEntry) string {
	role := strings.ToLower(appRole)
	best := matches[0].url
	bestScore := 0
	for _, m := range matches {
		score := 0
		mTitle := strings.ToLower(m.title)
		for _, word := range strings.Fields(role) {
			if len(word) > 2 && strings.Contains(mTitle, word) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			best = m.url
		}
	}
	return best
}
