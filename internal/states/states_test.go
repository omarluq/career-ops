package states_test

import (
	"testing"

	"github.com/omarluq/career-ops/internal/states"
	"github.com/samber/lo"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

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
		{"conditional", "evaluated"},
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
		{"\u2014", "discarded"},
		{"-", "discarded"},
		{"", "discarded"},

		// Contains-based fallback
		{"Interview - Round 2", "interview"},
		{"RE: Applied", "applied"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := states.Normalize(tt.input)
			assert.Equal(t, tt.want, got,
				"Normalize(%q)", tt.input)
		})
	}
}

func TestIsCanonical(t *testing.T) {
	t.Parallel()

	canonicals := []string{
		"evaluated", "applied", "responded",
		"interview", "offer", "rejected",
		"discarded", "skip",
	}
	lo.ForEach(canonicals, func(c string, _ int) {
		t.Run("canonical/"+c, func(t *testing.T) {
			t.Parallel()
			assert.True(t,
				states.IsCanonical(c),
				"IsCanonical(%q) should be true", c,
			)
		})
	})

	nonCanonicals := []string{
		"evaluada", "aplicado", "unknown", "foo",
	}
	lo.ForEach(nonCanonicals, func(nc string, _ int) {
		t.Run("non-canonical/"+nc, func(t *testing.T) {
			t.Parallel()
			assert.False(t,
				states.IsCanonical(nc),
				"IsCanonical(%q) should be false", nc,
			)
		})
	})
}

func TestLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want string
	}{
		{"evaluated", "Evaluated"},
		{"applied", "Applied"},
		{"skip", "SKIP"},
		{"interview", "Interview"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want,
				states.Label(tt.id),
				"Label(%q)", tt.id)
		})
	}
}

func TestPriority(t *testing.T) {
	t.Parallel()

	// Interview should be highest priority (lowest number)
	assert.Less(t,
		states.Priority("interview"),
		states.Priority("evaluated"),
		"interview should have higher priority than evaluated",
	)
	assert.Less(t,
		states.Priority("offer"),
		states.Priority("applied"),
		"offer should have higher priority than applied",
	)
	assert.Greater(t,
		states.Priority("discarded"),
		states.Priority("evaluated"),
		"discarded should have lower priority than evaluated",
	)
}

func TestStatusRank(t *testing.T) {
	t.Parallel()

	// Offer should be highest rank
	assert.Greater(t,
		states.StatusRank("offer"),
		states.StatusRank("applied"),
		"offer should rank higher than applied",
	)
	assert.Greater(t,
		states.StatusRank("applied"),
		states.StatusRank("evaluated"),
		"applied should rank higher than evaluated",
	)
	assert.Equal(t,
		states.StatusRank("skip"),
		states.StatusRank("discarded"),
		"skip and discarded should have same rank",
	)
}
