package main

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
)

var normalizeCmd = &cobra.Command{
	Use:   "normalize",
	Short: "Normalize status aliases to canonical statuses",
	Long: "Reads applications from the database, maps non-canonical statuses " +
		"to their canonical form, and updates them in-place.",
	RunE: runNormalize,
}

var (
	normalizePath   string
	normalizeDryRun bool
)

func init() {
	normalizeCmd.Flags().StringVar(
		&normalizePath, "path", ".",
		"path to career-ops root directory",
	)
	normalizeCmd.Flags().BoolVar(
		&normalizeDryRun, "dry-run", false,
		"show changes without writing",
	)
}

func runNormalize(_ *cobra.Command, _ []string) (err error) {
	careerOpsPath := normalizePath
	dbPath := viper.GetString("db")
	ctx := context.Background()

	// Initialize states from YAML (falls back to defaults).
	states.Init(careerOpsPath)

	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	apps, err := r.ListApplications(ctx)
	if err != nil {
		return oops.Wrapf(err, "listing applications")
	}

	if len(apps) == 0 {
		fmt.Println("No applications found in database. Nothing to normalize.")
		return nil
	}

	changes := 0
	var unknowns []unknownStatus

	lo.ForEach(apps, func(app model.CareerApplication, _ int) {
		rawStatus := app.Status
		normalizedID := states.Normalize(rawStatus)

		if !states.IsCanonical(normalizedID) {
			unknowns = append(unknowns, unknownStatus{
				num: app.Number, raw: rawStatus, line: 0,
			})
			return
		}

		label := states.Label(normalizedID)
		if label == rawStatus {
			return
		}

		fmt.Printf("#%d: %q -> %q\n", app.Number, rawStatus, label)

		if !normalizeDryRun {
			app.Status = label
			if upsertErr := r.UpsertApplication(ctx, &app); upsertErr != nil {
				fmt.Printf("  warning: updating #%d: %v\n", app.Number, upsertErr)
			}
		}
		changes++
	})

	if len(unknowns) > 0 {
		fmt.Printf("\nWarning: %d unknown statuses:\n", len(unknowns))
		lo.ForEach(unknowns, func(u unknownStatus, _ int) {
			fmt.Printf("  #%d: %q\n", u.num, u.raw)
		})
	}

	fmt.Printf("\n%d statuses normalized\n", changes)

	if normalizeDryRun {
		fmt.Println("(dry-run -- no changes written)")
	}

	return nil
}

type unknownStatus struct {
	raw  string
	num  int
	line int
}
