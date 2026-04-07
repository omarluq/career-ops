package db_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omarluq/career-ops/internal/db"
)

func validApp() *db.Application {
	return &db.Application{
		Number: 1, Date: "2026-04-06",
		Company: "Acme Corp", Role: "Senior Engineer",
		Score: 4.2, Status: "Evaluated",
	}
}

func TestValidateApplication_Valid(t *testing.T) {
	t.Parallel()

	t.Run("full valid app", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, db.ValidateApplication(validApp()))
	})

	t.Run("lowercase status", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Status = "evaluated"
		a.Number = 2
		require.NoError(t, db.ValidateApplication(a))
	})

	t.Run("zero score", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Score = 0
		a.Status = "skip"
		require.NoError(t, db.ValidateApplication(a))
	})
}

func TestValidateApplication_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		err := db.ValidateApplication(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be nil")
	})

	t.Run("missing company", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Company = ""
		err := db.ValidateApplication(a)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "company")
	})

	t.Run("missing role", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Role = ""
		err := db.ValidateApplication(a)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "role")
	})

	t.Run("missing date", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Date = ""
		err := db.ValidateApplication(a)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "date")
	})

	t.Run("score too high", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Score = 6.0
		err := db.ValidateApplication(a)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "score")
	})

	t.Run("negative score", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Score = -1.0
		err := db.ValidateApplication(a)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "score")
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()
		a := validApp()
		a.Status = "banana"
		err := db.ValidateApplication(a)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestValidatePipelineEntry(t *testing.T) {
	t.Parallel()

	valid := &db.PipelineEntry{
		URL: "https://example.com/j", Source: "li", Status: "pending",
	}

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		require.Error(t, db.ValidatePipelineEntry(nil))
	})
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, db.ValidatePipelineEntry(valid))
	})
	t.Run("no url", func(t *testing.T) {
		t.Parallel()
		e := *valid
		e.URL = ""
		require.Error(t, db.ValidatePipelineEntry(&e))
	})
	t.Run("no source", func(t *testing.T) {
		t.Parallel()
		e := *valid
		e.Source = ""
		require.Error(t, db.ValidatePipelineEntry(&e))
	})
	t.Run("no status", func(t *testing.T) {
		t.Parallel()
		e := *valid
		e.Status = ""
		require.Error(t, db.ValidatePipelineEntry(&e))
	})
}

func TestValidateScanRecord(t *testing.T) {
	t.Parallel()

	valid := &db.ScanRecord{
		URL: "https://example.com/j", Portal: "gh", Title: "SE",
	}

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		require.Error(t, db.ValidateScanRecord(nil))
	})
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, db.ValidateScanRecord(valid))
	})
	t.Run("no url", func(t *testing.T) {
		t.Parallel()
		r := *valid
		r.URL = ""
		require.Error(t, db.ValidateScanRecord(&r))
	})
	t.Run("no portal", func(t *testing.T) {
		t.Parallel()
		r := *valid
		r.Portal = ""
		require.Error(t, db.ValidateScanRecord(&r))
	})
	t.Run("no title", func(t *testing.T) {
		t.Parallel()
		r := *valid
		r.Title = ""
		require.Error(t, db.ValidateScanRecord(&r))
	})
	t.Run("company optional", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, db.ValidateScanRecord(valid))
	})
}

func TestValidateEvaluation(t *testing.T) {
	t.Parallel()

	valid := &db.Evaluation{
		ApplicationID: 1, EvaluationDate: "2026-04-06",
		RawScore: 4.5, Archetype: "backend",
	}

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		require.Error(t, db.ValidateEvaluation(nil))
	})
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, db.ValidateEvaluation(valid))
	})
	t.Run("zero app id", func(t *testing.T) {
		t.Parallel()
		e := *valid
		e.ApplicationID = 0
		require.Error(t, db.ValidateEvaluation(&e))
	})
	t.Run("no date", func(t *testing.T) {
		t.Parallel()
		e := *valid
		e.EvaluationDate = ""
		require.Error(t, db.ValidateEvaluation(&e))
	})
	t.Run("score too high", func(t *testing.T) {
		t.Parallel()
		e := *valid
		e.RawScore = 5.5
		require.Error(t, db.ValidateEvaluation(&e))
	})
	t.Run("negative score", func(t *testing.T) {
		t.Parallel()
		e := *valid
		e.RawScore = -0.5
		require.Error(t, db.ValidateEvaluation(&e))
	})
}

func TestIsValidStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"lowercase id", "evaluated", true},
		{"label", "Evaluated", true},
		{"applied", "applied", true},
		{"SKIP", "SKIP", true},
		{"case insensitive", "EVALUATED", true},
		{"invalid", "banana", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, db.IsValidStatus(tt.status))
		})
	}
}
