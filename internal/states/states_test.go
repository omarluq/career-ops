package states

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Canonical IDs (passthrough)
		{"evaluated", "evaluated"},
		{"applied", "applied"},
		{"interview", "interview"},
		{"skip", "skip"},

		// Labels (case-insensitive)
		{"Evaluated", "evaluated"},
		{"Applied", "applied"},
		{"SKIP", "skip"},

		// Spanish aliases from states.yml
		{"evaluada", "evaluated"},
		{"aplicado", "applied"},
		{"respondido", "responded"},
		{"entrevista", "interview"},
		{"oferta", "offer"},
		{"rechazado", "rejected"},
		{"rechazada", "rejected"},
		{"descartado", "discarded"},
		{"descartada", "discarded"},
		{"cerrada", "discarded"},
		{"cancelada", "discarded"},

		// English aliases
		{"enviada", "applied"},
		{"aplicada", "applied"},
		{"sent", "applied"},
		{"no_aplicar", "skip"},
		{"no aplicar", "skip"},
		{"monitor", "skip"},

		// Extra aliases from JS scripts
		{"condicional", "evaluated"},
		{"hold", "evaluated"},
		{"evaluar", "evaluated"},
		{"verificar", "evaluated"},

		// Markdown bold stripping
		{"**Evaluada**", "evaluated"},
		{"**Aplicado**", "applied"},
		{"**NO APLICAR**", "skip"},

		// Trailing dates
		{"aplicado 2026-03-12", "applied"},
		{"rechazado 2026-01-15", "rejected"},

		// Special patterns
		{"DUPLICADO #42", "discarded"},
		{"dup", "discarded"},
		{"Repost #123", "discarded"},
		{"GEO BLOCKER", "skip"},
		{"geo blocker", "skip"},
		{"—", "discarded"},
		{"-", "discarded"},
		{"", "discarded"},

		// Contains-based fallback
		{"Interview - Round 2", "interview"},
		{"RE: Applied", "applied"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Normalize(tt.input)
			assert.Equal(t, tt.want, got, "Normalize(%q)", tt.input)
		})
	}
}

func TestIsCanonical(t *testing.T) {
	canonicals := []string{"evaluated", "applied", "responded", "interview", "offer", "rejected", "discarded", "skip"}
	for _, c := range canonicals {
		assert.True(t, IsCanonical(c), "IsCanonical(%q) should be true", c)
	}
	nonCanonicals := []string{"evaluada", "aplicado", "unknown", "foo"}
	for _, nc := range nonCanonicals {
		assert.False(t, IsCanonical(nc), "IsCanonical(%q) should be false", nc)
	}
}

func TestLabel(t *testing.T) {
	tests := map[string]string{
		"evaluated": "Evaluated",
		"applied":   "Applied",
		"skip":      "SKIP",
		"interview": "Interview",
		"unknown":   "unknown",
	}
	for id, want := range tests {
		assert.Equal(t, want, Label(id), "Label(%q)", id)
	}
}

func TestPriority(t *testing.T) {
	// Interview should be highest priority (lowest number)
	assert.Less(t, Priority("interview"), Priority("evaluated"), "interview should have higher priority than evaluated")
	assert.Less(t, Priority("offer"), Priority("applied"), "offer should have higher priority than applied")
	assert.Greater(t, Priority("discarded"), Priority("evaluated"), "discarded should have lower priority than evaluated")
}

func TestStatusRank(t *testing.T) {
	// Offer should be highest rank
	assert.Greater(t, StatusRank("offer"), StatusRank("applied"), "offer should rank higher than applied")
	assert.Greater(t, StatusRank("applied"), StatusRank("evaluated"), "applied should rank higher than evaluated")
	assert.Equal(t, StatusRank("skip"), StatusRank("discarded"), "skip and discarded should have same rank")
}
