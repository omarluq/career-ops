// Package states provides canonical status definitions, normalization, and YAML loading
// for the career-ops pipeline. This is the single source of truth for all status handling,
// consolidating logic previously duplicated across 4+ JS scripts and the Go dashboard.
package states

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// Canonical status ID constants.
const (
	StatusEvaluated = "evaluated"
	StatusApplied   = "applied"
	StatusInterview = "interview"
	StatusOffer     = "offer"
	StatusResponded = "responded"
	StatusRejected  = "rejected"
	StatusDiscarded = "discarded"
	StatusSkip      = "skip"
)

// State represents a canonical pipeline state from states.yml.
type State struct {
	ID             string   `yaml:"id"`
	Label          string   `yaml:"label"`
	Description    string   `yaml:"description"`
	DashboardGroup string   `yaml:"dashboard_group"`
	Aliases        []string `yaml:"aliases"`
}

type statesFile struct {
	States []State `yaml:"states"`
}

// allStates holds the loaded states (defaults until Init is called).
var allStates = defaultStates()

// extraAliases covers aliases found in JS scripts but not in states.yml.
var extraAliases = map[string]string{
	"conditional": StatusEvaluated,
	"hold":        StatusEvaluated,
	"evaluar":     StatusEvaluated,
	"verificar":   StatusEvaluated,
}

func defaultStates() []State {
	return []State{
		{
			ID: StatusEvaluated, Label: "Evaluated",
			Aliases: []string{"evaluada"}, DashboardGroup: StatusEvaluated,
		},
		{
			ID: StatusApplied, Label: "Applied",
			Aliases:        []string{"aplicado", "enviada", "aplicada", "sent"},
			DashboardGroup: StatusApplied,
		},
		{
			ID: StatusResponded, Label: "Responded",
			Aliases: []string{"respondido"}, DashboardGroup: StatusResponded,
		},
		{
			ID: StatusInterview, Label: "Interview",
			Aliases: []string{"entrevista"}, DashboardGroup: StatusInterview,
		},
		{
			ID: StatusOffer, Label: "Offer",
			Aliases: []string{"oferta"}, DashboardGroup: StatusOffer,
		},
		{
			ID: StatusRejected, Label: "Rejected",
			Aliases:        []string{"rechazado", "rechazada"},
			DashboardGroup: StatusRejected,
		},
		{
			ID: StatusDiscarded, Label: "Discarded",
			Aliases:        []string{"descartado", "descartada", "cerrada", "cancelada"},
			DashboardGroup: StatusDiscarded,
		},
		{
			ID: StatusSkip, Label: "SKIP",
			Aliases:        []string{"no_aplicar", "no aplicar", "skip", "monitor"},
			DashboardGroup: StatusSkip,
		},
	}
}

// Init loads states from templates/states.yml in the given careerOpsPath.
// Falls back to defaults if the file is missing or malformed.
func Init(careerOpsPath string) {
	loaded := LoadStates(careerOpsPath)
	if len(loaded) > 0 {
		allStates = loaded
	}
}

// LoadStates reads states.yml from the given path.
// Tries templates/states.yml then states.yml. Returns nil if not found.
func LoadStates(careerOpsPath string) []State {
	paths := []string{
		filepath.Join(careerOpsPath, "templates", "states.yml"),
		filepath.Join(careerOpsPath, "states.yml"),
	}
	states, found := lo.Find(
		lo.FilterMap(paths, func(p string, _ int) ([]State, bool) {
			data, err := os.ReadFile(filepath.Clean(p))
			if err != nil {
				return nil, false
			}
			var sf statesFile
			if err := yaml.Unmarshal(data, &sf); err != nil {
				return nil, false
			}
			return sf.States, len(sf.States) > 0
		}),
		func(s []State) bool { return len(s) > 0 },
	)
	if found {
		return states
	}
	return nil
}

// All returns all loaded states.
func All() []State {
	return allStates
}

// containsFallback maps substring patterns to canonical status IDs
// for the contains-based fallback in Normalize.
var containsFallback = []struct {
	id      string
	substrs []string
}{
	{id: StatusSkip, substrs: []string{"no aplicar", "no_aplicar"}},
	{id: StatusInterview, substrs: []string{"interview", "entrevista"}},
	{id: StatusOffer, substrs: []string{"offer", "oferta"}},
	{id: StatusResponded, substrs: []string{"responded", "respondido"}},
	{id: StatusApplied, substrs: []string{"applied", "aplicado"}},
	{id: StatusRejected, substrs: []string{"rejected", "rechazado"}},
	{id: StatusDiscarded, substrs: []string{"discarded", "descartado"}},
	{id: StatusEvaluated, substrs: []string{"evaluated", "evaluada"}},
}

// Normalize maps a raw status string to its canonical ID.
// Handles markdown bold, trailing dates, Spanish/English aliases,
// and special patterns (duplicado, repost, geo blocker, etc.).
func Normalize(raw string) string {
	lower := prepareForNormalize(raw)

	// Empty / dash
	if lower == "" || lower == "\u2014" || lower == "-" {
		return StatusDiscarded
	}

	// Special prefix patterns (from JS normalize-statuses.mjs)
	if id, ok := matchSpecialPrefix(lower); ok {
		return id
	}

	// Exact match on canonical IDs, labels, and aliases
	if id, ok := matchExact(lower); ok {
		return id
	}

	// Extra aliases from JS scripts
	if id, ok := extraAliases[lower]; ok {
		return id
	}

	// Contains-based fallback (from Go dashboard NormalizeStatus)
	if id, ok := matchContains(lower); ok {
		return id
	}

	return lower
}

// prepareForNormalize strips markdown bold, trims, lowercases,
// and removes trailing dates.
func prepareForNormalize(raw string) string {
	s := strings.ReplaceAll(raw, "**", "")
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	if idx := strings.Index(lower, " 202"); idx > 0 {
		lower = strings.TrimSpace(lower[:idx])
	}
	return lower
}

// matchSpecialPrefix handles duplicado/dup/repost/geo blocker patterns.
func matchSpecialPrefix(lower string) (string, bool) {
	switch {
	case strings.HasPrefix(lower, "duplicado"),
		strings.HasPrefix(lower, "dup "),
		lower == "dup":
		return StatusDiscarded, true
	case strings.HasPrefix(lower, "repost"):
		return StatusDiscarded, true
	case strings.Contains(lower, "geo") && strings.Contains(lower, "blocker"):
		return StatusSkip, true
	}
	return "", false
}

// matchExact checks canonical IDs, labels, and YAML aliases.
func matchExact(lower string) (string, bool) {
	// Check canonical IDs
	if state, found := lo.Find(allStates, func(s State) bool {
		return lower == s.ID
	}); found {
		return state.ID, true
	}

	// Check labels (case-insensitive)
	if state, found := lo.Find(allStates, func(s State) bool {
		return strings.EqualFold(lower, s.Label)
	}); found {
		return state.ID, true
	}

	// Check aliases
	if state, found := lo.Find(allStates, func(s State) bool {
		return lo.Contains(
			lo.Map(s.Aliases, func(a string, _ int) string { return strings.ToLower(a) }),
			lower,
		)
	}); found {
		return state.ID, true
	}

	return "", false
}

// matchContains uses substring matching as a last resort.
func matchContains(lower string) (string, bool) {
	match, found := lo.Find(containsFallback, func(entry struct {
		id      string
		substrs []string
	}) bool {
		return lo.SomeBy(entry.substrs, func(sub string) bool {
			return strings.Contains(lower, sub)
		})
	})
	if found {
		return match.id, true
	}
	return "", false
}

// IsCanonical returns true if the given status is a canonical state ID.
func IsCanonical(status string) bool {
	lower := strings.ToLower(strings.TrimSpace(status))
	return lo.ContainsBy(allStates, func(s State) bool {
		return lower == s.ID
	})
}

// Label returns the display label for a canonical state ID.
func Label(id string) string {
	state, found := lo.Find(allStates, func(s State) bool {
		return s.ID == id
	})
	if found {
		return state.Label
	}
	return id
}

// priorityMap maps canonical status IDs to sort priority
// (lower = higher priority).
var priorityMap = map[string]int{
	StatusInterview: 0,
	StatusOffer:     1,
	StatusResponded: 2,
	StatusApplied:   3,
	StatusEvaluated: 4,
	StatusSkip:      5,
	StatusRejected:  6,
	StatusDiscarded: 7,
}

// Priority returns sort priority for dashboard display (lower = higher priority).
func Priority(status string) int {
	if p, ok := priorityMap[Normalize(status)]; ok {
		return p
	}
	return 8
}

// statusRankMap maps canonical status IDs to advancement rank
// (higher = more advanced in pipeline).
var statusRankMap = map[string]int{
	StatusSkip:      0,
	StatusDiscarded: 0,
	StatusRejected:  1,
	StatusEvaluated: 2,
	StatusApplied:   3,
	StatusResponded: 4,
	StatusInterview: 5,
	StatusOffer:     6,
}

// StatusRank returns advancement rank for dedup (higher = more advanced in pipeline).
func StatusRank(status string) int {
	if r, ok := statusRankMap[Normalize(status)]; ok {
		return r
	}
	return 0
}
