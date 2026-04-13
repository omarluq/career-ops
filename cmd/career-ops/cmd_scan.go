package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
	"github.com/omarluq/career-ops/internal/scanner"
)

var scanConcurrency int

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan job portals for new listings",
	Long: "Scans configured portals concurrently using headless Chrome.\n" +
		"Reads portals.yml for portal configuration and reports new findings.",
	RunE: runScan,
}

func init() {
	scanCmd.Flags().IntVar(&scanConcurrency, "concurrency", 3, "number of concurrent browser contexts")
}

func runScan(cmd *cobra.Command, _ []string) (err error) {
	ctx := cmd.Context()
	dbPath := viper.GetString("db")

	portalsPath := "portals.yml"
	if _, statErr := os.Stat(portalsPath); statErr != nil {
		return oops.Errorf("portals.yml not found -- run onboarding first or copy templates/portals.example.yml")
	}

	portals, titleFilter, err := scanner.LoadPortals(portalsPath)
	if err != nil {
		return oops.Wrapf(err, "loading portals config")
	}

	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	stderrf("Scanning %d portal(s) with concurrency=%d...\n", len(portals), scanConcurrency)

	sc := scanner.NewScanner(scanConcurrency)

	results, scanErr := sc.ScanPortals(ctx, portals, func(portal string, found int) {
		stderrf("  [%s] found %d listing(s)\n", portal, found)
	})

	if scanErr != nil {
		stderrf("warning: %v\n", scanErr)
	}

	// Apply title filter.
	if titleFilter != nil {
		results = lo.Filter(results, func(r scanner.ScanResult, _ int) bool {
			return matchesTitleFilter(r.Title, titleFilter)
		})
	}

	// Dedup against scan history in SQLite.
	newResults := filterNewResults(ctx, r, results)

	stderrf("\nTotal found: %d | After filter: %d | New: %d\n",
		len(results), len(results), len(newResults))

	if len(newResults) == 0 {
		if _, printErr := fmt.Println("No new listings found."); printErr != nil {
			return oops.Wrapf(printErr, "writing output")
		}
		return nil
	}

	mw := &mdWriter{w: os.Stdout}
	mw.writef("\n--- New Listings (%d) ---\n\n", len(newResults))

	lo.ForEach(newResults, func(r scanner.ScanResult, i int) {
		company := lo.Ternary(r.Company != "", r.Company, r.Portal)
		mw.writef("%d. [%s] %s\n   %s\n\n", i+1, company, r.Title, r.URL)
	})

	return mw.err
}

// filterNewResults returns only results not already in scan history.
func filterNewResults(
	ctx context.Context, r db.Repository, results []scanner.ScanResult,
) []scanner.ScanResult {
	return lo.Filter(results, func(sr scanner.ScanResult, _ int) bool {
		scanned, err := r.HasBeenScanned(ctx, strings.ToLower(sr.URL), sr.Portal)
		if err != nil {
			return true // on error, include the result
		}
		return !scanned
	})
}

// matchesTitleFilter returns true if the title matches the positive/negative keyword rules.
func matchesTitleFilter(title string, tf *scanner.TitleFilter) bool {
	lower := strings.ToLower(title)

	// At least one positive keyword must match.
	hasPositive := lo.SomeBy(tf.Positive, func(kw string) bool {
		return strings.Contains(lower, strings.ToLower(kw))
	})
	if !hasPositive {
		return false
	}

	// No negative keyword may match.
	hasNegative := lo.SomeBy(tf.Negative, func(kw string) bool {
		return strings.Contains(lower, strings.ToLower(kw))
	})
	return !hasNegative
}

// stderrf writes a formatted message to stderr (best-effort, errors are not actionable).
func stderrf(format string, args ...any) {
	if _, err := fmt.Fprintf(os.Stderr, format, args...); err != nil {
		// Stderr write failure is non-recoverable; nothing to do.
		return
	}
}
