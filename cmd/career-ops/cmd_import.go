package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // register sqlite driver for database/sql

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/repo"
	"github.com/omarluq/career-ops/internal/tracker"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import markdown data into SQLite",
	Long: "Reads existing applications.md, pipeline.md, and scan-history.tsv " +
		"and inserts all records into the SQLite database.",
	RunE: runImport,
}

func runImport(cmd *cobra.Command, _ []string) (err error) {
	ctx := cmd.Context()
	dbPath := viper.GetString("db")
	basePath, err := os.Getwd()
	if err != nil {
		return oops.Wrapf(err, "getting working directory")
	}

	g := closer.Guard{Err: &err}

	db, err := repo.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(db)

	r := repo.NewSQLite(db)
	defer g.Close(r)

	appCount, err := importApplications(ctx, r, basePath)
	if err != nil {
		return err
	}

	pipeCount, err := importPipeline(ctx, r, basePath)
	if err != nil {
		return err
	}

	scanCount, err := importScanHistory(ctx, r, basePath)
	if err != nil {
		return err
	}

	if _, printErr := fmt.Fprintf(cmd.OutOrStdout(),
		"Imported %d applications, %d pipeline entries, %d scan records\n",
		appCount, pipeCount, scanCount); printErr != nil {
		return oops.Wrapf(printErr, "writing import summary")
	}
	return nil
}

// importApplications parses applications.md and upserts every entry.
func importApplications(
	ctx context.Context, r repo.Repository, basePath string,
) (int, error) {
	apps, err := tracker.ParseApplications(basePath)
	if err != nil {
		return 0, oops.Wrapf(err, "parsing applications")
	}

	for i := range apps {
		if err := r.UpsertApplication(ctx, &apps[i]); err != nil {
			return 0, oops.Wrapf(err, "upserting application %d", apps[i].Number)
		}
	}
	return len(apps), nil
}

// importPipeline reads pipeline.md and adds each URL to the pipeline table.
func importPipeline(
	ctx context.Context, r repo.Repository, basePath string,
) (count int, err error) {
	g := closer.Guard{Err: &err}
	pipePath := filepath.Join(basePath, "data", "pipeline.md")
	f, err := os.Open(filepath.Clean(pipePath))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, oops.Wrapf(err, "opening pipeline.md")
	}
	defer g.Close(f)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		url, source := parsePipelineLine(line)
		if url == "" {
			continue
		}
		if err := r.AddToPipeline(ctx, url, source); err != nil {
			return count, oops.Wrapf(err, "adding pipeline entry %s", url)
		}
		count++
	}
	return count, oops.Wrapf(sc.Err(), "scanning pipeline.md")
}

// parsePipelineLine extracts a URL and optional source from a pipeline.md line.
// Supported formats:
//
//	https://example.com/job
//	- https://example.com/job
//	- https://example.com/job (source: linkedin)
//	local:jds/filename.md
func parsePipelineLine(line string) (url, source string) {
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimPrefix(line, "* ")

	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "<!--") {
		return "", ""
	}

	// Extract source annotation if present.
	if idx := strings.Index(line, "(source:"); idx > 0 {
		source = strings.TrimSpace(strings.TrimSuffix(
			line[idx+len("(source:"):], ")"))
		line = strings.TrimSpace(line[:idx])
	}

	if strings.HasPrefix(line, "http") || strings.HasPrefix(line, "local:") {
		return line, source
	}
	return "", ""
}

// readScanHistoryTSV tries to read scan-history.tsv from data/ or root.
// Returns nil data (not an error) if the file does not exist.
func readScanHistoryTSV(basePath string) ([]byte, error) {
	candidates := []string{
		filepath.Join(basePath, "data", "scan-history.tsv"),
		filepath.Join(basePath, "scan-history.tsv"),
	}
	for _, p := range candidates {
		data, err := os.ReadFile(filepath.Clean(p))
		if err == nil {
			return data, nil
		}
		if !os.IsNotExist(err) {
			return nil, oops.Wrapf(err, "reading %s", p)
		}
	}
	return nil, nil
}

// importScanEntry holds a parsed scan-history.tsv row for import.
type importScanEntry struct {
	URL, Portal, Title, Company string
}

// importScanHistory reads scan-history.tsv and records each entry.
func importScanHistory(
	ctx context.Context, r repo.Repository, basePath string,
) (int, error) {
	data, err := readScanHistoryTSV(basePath)
	if err != nil {
		return 0, err
	}
	if data == nil {
		return 0, nil
	}

	lines := strings.Split(string(data), "\n")
	entries := lo.FilterMap(lines, func(line string, _ int) (importScanEntry, bool) {
		fields := strings.Split(line, "\t")
		if len(fields) < 5 || fields[0] == "url" || fields[0] == "" {
			return importScanEntry{}, false
		}
		if !strings.HasPrefix(fields[0], "http") {
			return importScanEntry{}, false
		}
		return importScanEntry{
			URL:     fields[0],
			Portal:  fields[1],
			Title:   fields[3],
			Company: fields[4],
		}, true
	})

	count := 0
	for _, e := range entries {
		if err := r.RecordScan(ctx, e.URL, e.Portal, e.Title, e.Company); err != nil {
			return count, oops.Wrapf(err, "recording scan %s", e.URL)
		}
		count++
	}
	return count, nil
}
