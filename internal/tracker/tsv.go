package tracker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/omarluq/career-ops/internal/states"
)

// Addition represents a parsed addition from batch/tracker-additions/*.tsv.
type Addition struct {
	Date    string
	Company string
	Role    string
	Status  string
	Score   string
	PDF     string
	Report  string
	Notes   string
	Num     int
}

var reExtractReportNum = regexp.MustCompile(`\[(\d+)\]`)

// ExtractReportNum extracts the report number from a markdown link like "[123](reports/...)".
func ExtractReportNum(reportStr string) int {
	m := reExtractReportNum.FindStringSubmatch(reportStr)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

// ParseTSVContent parses a TSV file content into a tracker addition.
// Handles 9-col TSV, 8-col TSV, and pipe-delimited markdown.
func ParseTSVContent(content, filename string) *Addition {
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

func parsePipeContent(content, _ string) *Addition {
	parts := strings.Split(content, "|")
	fields := lo.Filter(
		lo.Map(parts, func(p string, _ int) string { return strings.TrimSpace(p) }),
		func(p string, _ int) bool { return p != "" },
	)
	if len(fields) < 8 {
		return nil
	}

	num, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil || num == 0 {
		return nil
	}

	return &Addition{
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

func parseTSVTabContent(content, _ string) *Addition {
	parts := strings.Split(content, "\t")
	if len(parts) < 8 {
		return nil
	}

	num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || num == 0 {
		return nil
	}

	statusCol, scoreCol := detectStatusScoreColumns(
		strings.TrimSpace(parts[4]),
		strings.TrimSpace(parts[5]),
	)

	return &Addition{
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

// detectStatusScoreColumns resolves the ambiguous column order between
// status and score in TSV files. Returns (status, score).
func detectStatusScoreColumns(col4, col5 string) (statusCol, scoreCol string) {
	switch {
	case isStatusFormat(col4) && !isScoreFormat(col4):
		return col4, col5
	case isScoreFormat(col4) && isStatusFormat(col5):
		return col5, col4
	case isScoreFormat(col5) && !isScoreFormat(col4):
		return col4, col5
	default:
		return col4, col5
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
	"repost", "conditional", "hold", "monitor", "evaluated", "applied",
	"responded", "interview", "offer", "rejected", "discarded", "skip",
}

func isStatusFormat(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	return lo.SomeBy(statusPrefixes, func(prefix string) bool {
		return strings.HasPrefix(lower, prefix)
	})
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
func FormatTSVAddition(a *Addition) string {
	return fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
		a.Num, a.Date, a.Company, a.Role, a.Status, a.Score, a.PDF, a.Report, a.Notes)
}
