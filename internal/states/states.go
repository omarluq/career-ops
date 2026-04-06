// Package states provides canonical status definitions, normalization, and YAML loading
// for the career-ops pipeline. This is the single source of truth for all status handling,
// consolidating logic previously duplicated across 4+ JS scripts and the Go dashboard.
package states

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// State represents a canonical pipeline state from states.yml.
type State struct {
	ID             string   `yaml:"id"`
	Label          string   `yaml:"label"`
	Aliases        []string `yaml:"aliases"`
	Description    string   `yaml:"description"`
	DashboardGroup string   `yaml:"dashboard_group"`
}

type statesFile struct {
	States []State `yaml:"states"`
}

// allStates holds the loaded states (defaults until Init is called).
var allStates = defaultStates()

// extraAliases covers aliases found in JS scripts but not in states.yml.
var extraAliases = map[string]string{
	"condicional": "evaluated",
	"hold":        "evaluated",
	"evaluar":     "evaluated",
	"verificar":   "evaluated",
}

func defaultStates() []State {
	return []State{
		{ID: "evaluated", Label: "Evaluated", Aliases: []string{"evaluada"}, DashboardGroup: "evaluated"},
		{ID: "applied", Label: "Applied", Aliases: []string{"aplicado", "enviada", "aplicada", "sent"}, DashboardGroup: "applied"},
		{ID: "responded", Label: "Responded", Aliases: []string{"respondido"}, DashboardGroup: "responded"},
		{ID: "interview", Label: "Interview", Aliases: []string{"entrevista"}, DashboardGroup: "interview"},
		{ID: "offer", Label: "Offer", Aliases: []string{"oferta"}, DashboardGroup: "offer"},
		{ID: "rejected", Label: "Rejected", Aliases: []string{"rechazado", "rechazada"}, DashboardGroup: "rejected"},
		{ID: "discarded", Label: "Discarded", Aliases: []string{"descartado", "descartada", "cerrada", "cancelada"}, DashboardGroup: "discarded"},
		{ID: "skip", Label: "SKIP", Aliases: []string{"no_aplicar", "no aplicar", "skip", "monitor"}, DashboardGroup: "skip"},
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
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var sf statesFile
		if err := yaml.Unmarshal(data, &sf); err != nil {
			continue
		}
		if len(sf.States) > 0 {
			return sf.States
		}
	}
	return nil
}

// All returns all loaded states.
func All() []State {
	return allStates
}

// Normalize maps a raw status string to its canonical ID.
// Handles markdown bold, trailing dates, Spanish/English aliases,
// and special patterns (duplicado, repost, geo blocker, etc.).
func Normalize(raw string) string {
	// Strip markdown bold and trim
	s := strings.ReplaceAll(raw, "**", "")
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)

	// Strip trailing dates (e.g., "aplicado 2026-03-12")
	if idx := strings.Index(lower, " 202"); idx > 0 {
		lower = strings.TrimSpace(lower[:idx])
	}

	// Empty / dash
	if lower == "" || lower == "—" || lower == "-" {
		return "discarded"
	}

	// Special prefix patterns (from JS normalize-statuses.mjs)
	switch {
	case strings.HasPrefix(lower, "duplicado"),
		strings.HasPrefix(lower, "dup "),
		lower == "dup":
		return "discarded"
	case strings.HasPrefix(lower, "repost"):
		return "discarded"
	case strings.Contains(lower, "geo") && strings.Contains(lower, "blocker"):
		return "skip"
	}

	// Exact match on canonical IDs
	for _, state := range allStates {
		if lower == state.ID {
			return state.ID
		}
	}

	// Match on labels (case-insensitive)
	for _, state := range allStates {
		if lower == strings.ToLower(state.Label) {
			return state.ID
		}
	}

	// Match on YAML aliases
	for _, state := range allStates {
		for _, alias := range state.Aliases {
			if lower == strings.ToLower(alias) {
				return state.ID
			}
		}
	}

	// Extra aliases from JS scripts
	if id, ok := extraAliases[lower]; ok {
		return id
	}

	// Contains-based fallback (from Go dashboard NormalizeStatus)
	switch {
	case strings.Contains(lower, "no aplicar") || strings.Contains(lower, "no_aplicar"):
		return "skip"
	case strings.Contains(lower, "interview") || strings.Contains(lower, "entrevista"):
		return "interview"
	case strings.Contains(lower, "offer") || strings.Contains(lower, "oferta"):
		return "offer"
	case strings.Contains(lower, "responded") || strings.Contains(lower, "respondido"):
		return "responded"
	case strings.Contains(lower, "applied") || strings.Contains(lower, "aplicado"):
		return "applied"
	case strings.Contains(lower, "rejected") || strings.Contains(lower, "rechazado"):
		return "rejected"
	case strings.Contains(lower, "discarded") || strings.Contains(lower, "descartado"):
		return "discarded"
	case strings.Contains(lower, "evaluated") || strings.Contains(lower, "evaluada"):
		return "evaluated"
	}

	return lower
}

// IsCanonical returns true if the given status is a canonical state ID.
func IsCanonical(status string) bool {
	lower := strings.ToLower(strings.TrimSpace(status))
	for _, state := range allStates {
		if lower == state.ID {
			return true
		}
	}
	return false
}

// Label returns the display label for a canonical state ID.
func Label(id string) string {
	for _, state := range allStates {
		if state.ID == id {
			return state.Label
		}
	}
	return id
}

// Priority returns sort priority for dashboard display (lower = higher priority).
func Priority(status string) int {
	switch Normalize(status) {
	case "interview":
		return 0
	case "offer":
		return 1
	case "responded":
		return 2
	case "applied":
		return 3
	case "evaluated":
		return 4
	case "skip":
		return 5
	case "rejected":
		return 6
	case "discarded":
		return 7
	default:
		return 8
	}
}

// StatusRank returns advancement rank for dedup (higher = more advanced in pipeline).
func StatusRank(status string) int {
	switch Normalize(status) {
	case "skip", "discarded":
		return 0
	case "rejected":
		return 1
	case "evaluated":
		return 2
	case "applied":
		return 3
	case "responded":
		return 4
	case "interview":
		return 5
	case "offer":
		return 6
	default:
		return 0
	}
}
