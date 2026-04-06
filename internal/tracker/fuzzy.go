package tracker

import (
	"strconv"
	"strings"
)

// NormalizeCompany strips common suffixes and normalizes a company name for matching.
func NormalizeCompany(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	for _, suffix := range []string{
		" inc.", " inc", " llc", " ltd", " corp", " corporation",
		" technologies", " technology", " group", " co.",
	} {
		s = strings.TrimSuffix(s, suffix)
	}
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// NormalizeCompanyKey returns a key suitable for map lookups (no spaces, no special chars).
func NormalizeCompanyKey(name string) string {
	s := NormalizeCompany(name)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// NormalizeRole normalizes a role title for comparison.
func NormalizeRole(role string) string {
	s := strings.ToLower(strings.TrimSpace(role))
	s = strings.ReplaceAll(s, "(", " ")
	s = strings.ReplaceAll(s, ")", " ")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '/' {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(strings.Join(strings.Fields(b.String()), " "))
}

// RoleMatch returns true if two roles are similar enough to be considered duplicates.
// Uses word overlap: at least 2 significant words (>3 chars) must match.
func RoleMatch(a, b string) bool {
	wordsA := significantWords(NormalizeRole(a))
	wordsB := significantWords(NormalizeRole(b))
	overlap := 0
	for _, wa := range wordsA {
		for _, wb := range wordsB {
			if strings.Contains(wa, wb) || strings.Contains(wb, wa) {
				overlap++
				break
			}
		}
	}
	return overlap >= 2
}

// ParseScore extracts a numeric score from a score string like "4.2/5" or "**3.8/5**".
func ParseScore(s string) float64 {
	s = strings.ReplaceAll(s, "**", "")
	m := reScoreValue.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	v, _ := strconv.ParseFloat(m[1], 64)
	return v
}

func significantWords(s string) []string {
	words := strings.Fields(s)
	var result []string
	for _, w := range words {
		if len(w) > 3 {
			result = append(result, w)
		}
	}
	return result
}
