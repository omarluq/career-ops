package tracker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/omarluq/career-ops/internal/states"
)

// TrackerAddition represents a parsed addition from batch/tracker-additions/*.tsv.
type TrackerAddition struct {
	Num     int
	Date    string
	Company string
	Role    string
	Status  string
	Score   string
	PDF     string
	Report  string
	Notes   string
}

var reExtractReportNum = regexp.MustCompile(`\[(\d+)\]`)

// ExtractReportNum extracts the report number from a markdown link like "[123](reports/...)".
func ExtractReportNum(reportStr string) int {
	m := reExtractReportNum.FindStringSubmatch(reportStr)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// ParseTSVContent parses a TSV file content into a tracker addition.
// Handles 9-col TSV, 8-col TSV, and pipe-delimited markdown.
func ParseTSVContent(content, filename string) *TrackerAddition {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	// Detect pipe-delimited (markdown table row)
	if strings.HasPrefix(content, "|") {
		return parsePipeContent(content, filename)
	}
	return parseTSVTabContent(content, filename)
}

func parsePipeContent(content, filename string) *TrackerAddition {
	parts := strings.Split(content, "|")
	var fields []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			fields = append(fields, p)
		}
	}
	if len(fields) < 8 {
		return nil
	}

	num, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil || num == 0 {
		return nil
	}

	return &TrackerAddition{
		Num:     num,
		Date:    fields[1],
		Company: fields[2],
		Role:    fields[3],
		Score:   fields[4],
		Status:  states.Label(states.Normalize(fields[5])),
		PDF:     fields[6],
		Report:  fields[7],
		Notes:   safeIndex(fields, 8),
	}
}

func parseTSVTabContent(content, filename string) *TrackerAddition {
	parts := strings.Split(content, "\t")
	if len(parts) < 8 {
		return nil
	}

	// Detect column order: col4 and col5 could be (status, score) or (score, status)
	col4 := strings.TrimSpace(parts[4])
	col5 := strings.TrimSpace(parts[5])

	col4IsScore := isScoreFormat(col4)
	col5IsScore := isScoreFormat(col5)
	col4IsStatus := isStatusFormat(col4)
	col5IsStatus := isStatusFormat(col5)

	var statusCol, scoreCol string
	switch {
	case col4IsStatus && !col4IsScore:
		statusCol, scoreCol = col4, col5
	case col4IsScore && col5IsStatus:
		statusCol, scoreCol = col5, col4
	case col5IsScore && !col4IsScore:
		statusCol, scoreCol = col4, col5
	default:
		statusCol, scoreCol = col4, col5
	}

	num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || num == 0 {
		return nil
	}

	return &TrackerAddition{
		Num:     num,
		Date:    strings.TrimSpace(parts[1]),
		Company: strings.TrimSpace(parts[2]),
		Role:    strings.TrimSpace(parts[3]),
		Status:  states.Label(states.Normalize(statusCol)),
		Score:   strings.TrimSpace(scoreCol),
		PDF:     strings.TrimSpace(parts[6]),
		Report:  strings.TrimSpace(parts[7]),
		Notes:   safeIndexTab(parts, 8),
	}
}

var reScoreFormat = regexp.MustCompile(`^\d+\.?\d*/5$`)

func isScoreFormat(s string) bool {
	s = strings.TrimSpace(s)
	return reScoreFormat.MatchString(s) || s == "N/A" || s == "DUP"
}

var statusPrefixes = []string{
	"evaluada", "aplicado", "respondido", "entrevista", "oferta",
	"rechazado", "descartado", "no aplicar", "cerrada", "duplicado",
	"repost", "condicional", "hold", "monitor", "evaluated", "applied",
	"responded", "interview", "offer", "rejected", "discarded", "skip",
}

func isStatusFormat(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, prefix := range statusPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

func safeIndex(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return ""
}

func safeIndexTab(s []string, i int) string {
	if i < len(s) {
		return strings.TrimSpace(s[i])
	}
	return ""
}

// FormatTSVAddition formats a tracker addition as a TSV line (for batch output).
func FormatTSVAddition(a *TrackerAddition) string {
	return fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
		a.Num, a.Date, a.Company, a.Role, a.Status, a.Score, a.PDF, a.Report, a.Notes)
}
