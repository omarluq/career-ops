package tracker_test

import (
	"fmt"
	"testing"

	"github.com/omarluq/career-ops/internal/tracker"
	"github.com/samber/lo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeCompany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input, want string
	}{
		{"Google LLC", "google"},
		{"Google Inc.", "google"},
		{"  Meta  ", "meta"},
		{"Stripe Technologies", "stripe"},
		{"OpenAI (SF)", "openai sf"},
		{"Amazon Corp", "amazon"},
		{"plain name", "plain name"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want,
				tracker.NormalizeCompany(tt.input),
				"NormalizeCompany(%q)", tt.input)
		})
	}
}

func TestNormalizeCompanyKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input, want string
	}{
		{"Google LLC", "google"},
		{"Open AI (SF)", "openaisf"},
		{"Stripe Technologies", "stripe"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want,
				tracker.NormalizeCompanyKey(tt.input),
				"NormalizeCompanyKey(%q)", tt.input)
		})
	}
}

func TestRoleMatch(t *testing.T) {
	t.Parallel()

	positives := [][2]string{
		{"Senior Backend Engineer", "Backend Engineer"},
		{"Staff Software Engineer", "Software Engineer - Platform"},
		{"Senior Platform Engineer", "Platform Engineer - Infrastructure"},
	}
	lo.ForEach(positives, func(pair [2]string, _ int) {
		name := fmt.Sprintf("match/%s_vs_%s", pair[0], pair[1])
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.True(t,
				tracker.RoleMatch(pair[0], pair[1]),
				"RoleMatch(%q, %q) should be true",
				pair[0], pair[1])
		})
	})

	negatives := [][2]string{
		{"Frontend Developer", "Backend Engineer"},
		{"Product Manager", "Software Engineer"},
		{"Data Scientist", "DevOps Engineer"},
	}
	lo.ForEach(negatives, func(pair [2]string, _ int) {
		name := fmt.Sprintf("no-match/%s_vs_%s", pair[0], pair[1])
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.False(t,
				tracker.RoleMatch(pair[0], pair[1]),
				"RoleMatch(%q, %q) should be false",
				pair[0], pair[1])
		})
	})
}

func TestParseScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  float64
	}{
		{"4.2/5", 4.2},
		{"**3.8/5**", 3.8},
		{"N/A", 0},
		{"DUP", 0},
		{"", 0},
		{"5/5", 5.0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want,
				tracker.ParseScore(tt.input),
				"ParseScore(%q)", tt.input)
		})
	}
}

func TestExtractReportNum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  int
	}{
		{"[123](reports/123-company-2026-01-01.md)", 123},
		{"[001](reports/001-test.md)", 1},
		{"no report", 0},
		{"", 0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want,
				tracker.ExtractReportNum(tt.input),
				"ExtractReportNum(%q)", tt.input)
		})
	}
}

func TestParseTSVContent(t *testing.T) {
	t.Parallel()

	// 9-col TSV
	tsv9 := "42\t2026-04-01\tGoogle\tSenior SWE\t" +
		"Evaluated\t4.2/5\t\u2705\t" +
		"[42](reports/042-google-2026-04-01.md)\tGreat match"
	result := tracker.ParseTSVContent(tsv9, "042-google.tsv")
	require.NotNil(t, result,
		"ParseTSVContent returned nil for valid 9-col TSV")
	assert.Equal(t, 42, result.Num)
	assert.Equal(t, "Google", result.Company)

	// Pipe-delimited
	pipe := "| 7 | 2026-04-01 | Stripe | Backend Engineer " +
		"| 4.5/5 | Evaluated | \u2705 " +
		"| [7](reports/007-stripe.md) | Nice |"
	result = tracker.ParseTSVContent(pipe, "007-stripe.tsv")
	require.NotNil(t, result,
		"ParseTSVContent returned nil for valid pipe content")
	assert.Equal(t, 7, result.Num)

	// Empty content
	assert.Nil(t,
		tracker.ParseTSVContent("", "empty.tsv"),
		"ParseTSVContent should return nil for empty content",
	)

	// Malformed
	assert.Nil(t,
		tracker.ParseTSVContent("not\tenough\tcols", "bad.tsv"),
		"ParseTSVContent should return nil for too few columns",
	)
}
