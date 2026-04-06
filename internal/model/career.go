// Package model defines data types for the career-ops pipeline.
package model

// CareerApplication represents a single job application from the tracker.
type CareerApplication struct {
	ReportPath   string
	ReportNumber string
	Company      string
	Role         string
	Status       string
	CompEstimate string
	Date         string
	ScoreRaw     string
	Remote       string
	TlDr         string
	Notes        string
	JobURL       string
	Archetype    string
	Number       int
	Score        float64
	HasPDF       bool
}

// PipelineMetrics holds aggregate stats for the pipeline dashboard.
type PipelineMetrics struct {
	ByStatus   map[string]int
	Total      int
	AvgScore   float64
	TopScore   float64
	WithPDF    int
	Actionable int
}
