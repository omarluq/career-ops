package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
	"github.com/omarluq/career-ops/internal/model"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export SQLite data to markdown",
	Long: "Reads the SQLite database and writes applications as a markdown table " +
		"to stdout or an output file. Use --format to choose between applications, " +
		"pipeline, or scan-history.",
	RunE: runExport,
}

func init() {
	exportCmd.Flags().StringP("output", "o", "", "write output to file instead of stdout")
	exportCmd.Flags().String("format", "applications",
		"what to export: applications, pipeline, scan-history")
}

// mdWriter wraps an io.Writer and accumulates the first write error.
type mdWriter struct {
	w   io.Writer
	err error
}

// writeln writes a line followed by a newline to the writer.
func (m *mdWriter) writeln(s string) {
	if m.err != nil {
		return
	}
	_, m.err = fmt.Fprintln(m.w, s)
}

// writef writes a formatted string to the writer.
func (m *mdWriter) writef(format string, args ...any) {
	if m.err != nil {
		return
	}
	_, m.err = fmt.Fprintf(m.w, format, args...)
}

func runExport(cmd *cobra.Command, _ []string) (err error) {
	ctx := cmd.Context()
	dbPath := viper.GetString("db")
	outputPath, err := cmd.Flags().GetString("output")
	if err != nil {
		return oops.Wrapf(err, "reading --output flag")
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return oops.Wrapf(err, "reading --format flag")
	}

	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	w := cmd.OutOrStdout()
	if outputPath != "" {
		f, createErr := os.Create(filepath.Clean(outputPath))
		if createErr != nil {
			return oops.Wrapf(createErr, "creating output file %s", outputPath)
		}
		defer g.Close(f)
		w = f
	}

	switch format {
	case "applications":
		return exportApplications(ctx, r, w)
	case "pipeline":
		return exportPipeline(ctx, r, w)
	case "scan-history":
		return exportScanHistory(ctx, r, w)
	default:
		return oops.Errorf("unknown export format: %s", format)
	}
}

// exportApplications writes all applications as a markdown table matching
// the applications.md format.
func exportApplications(ctx context.Context, r db.Repository, w io.Writer) error {
	apps, err := r.ListApplications(ctx)
	if err != nil {
		return err
	}

	mw := &mdWriter{w: w}

	mw.writeln("# Applications Tracker")
	mw.writeln("")
	mw.writeln("| # | Date | Company | Role | Score | Status | PDF | Report | Notes |")
	mw.writeln("|---|------|---------|------|-------|--------|-----|--------|-------|")

	lo.ForEach(apps, func(app model.CareerApplication, _ int) {
		pdf := "❌"
		if app.HasPDF {
			pdf = "✅"
		}

		report := ""
		if app.ReportNumber != "" && app.ReportPath != "" {
			report = fmt.Sprintf("[%s](%s)", app.ReportNumber, app.ReportPath)
		}

		score := app.ScoreRaw
		if score == "" && app.Score > 0 {
			score = fmt.Sprintf("%.1f/5", app.Score)
		}

		mw.writef("| %d | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			app.Number,
			app.Date,
			app.Company,
			app.Role,
			score,
			app.Status,
			pdf,
			report,
			strings.TrimSpace(app.Notes),
		)
	})

	return mw.err
}

// exportPipeline writes all pipeline entries as a markdown list.
func exportPipeline(ctx context.Context, r db.Repository, w io.Writer) error {
	entries, err := r.ListPipeline(ctx, "")
	if err != nil {
		return err
	}

	mw := &mdWriter{w: w}
	mw.writeln("# Pipeline")
	mw.writeln("")

	lo.ForEach(entries, func(e db.RepoPipelineEntry, _ int) {
		line := fmt.Sprintf("- %s", e.URL)
		if e.Source != "" {
			line += fmt.Sprintf(" (source: %s)", e.Source)
		}
		if e.Status != "pending" {
			line += fmt.Sprintf(" [%s]", e.Status)
		}
		mw.writeln(line)
	})

	return mw.err
}

// exportScanHistory writes all scan records as a TSV matching scan-history.tsv format.
func exportScanHistory(ctx context.Context, r db.Repository, w io.Writer) error {
	records, err := r.ListScanHistory(ctx)
	if err != nil {
		return err
	}

	mw := &mdWriter{w: w}
	mw.writeln("url\tportal\tdate\ttitle\tcompany")

	lo.ForEach(records, func(rec db.RepoScanRecord, _ int) {
		mw.writef("%s\t%s\t%s\t%s\t%s\n",
			rec.URL,
			rec.Portal,
			rec.ScannedAt.Format("2006-01-02"),
			rec.Title,
			rec.Company,
		)
	})

	return mw.err
}
