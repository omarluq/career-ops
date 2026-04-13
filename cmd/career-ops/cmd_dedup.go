package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
)

var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Remove duplicate tracker entries",
	Long: `Groups entries by normalized company key, checks role matches within groups,
keeps the entry with the highest score, promotes status if a discarded duplicate
had a more advanced pipeline status, and deletes duplicates from the database.`,
	RunE: runDedup,
}

var (
	dedupPath   string
	dedupDryRun bool
)

func init() {
	dedupCmd.Flags().StringVar(&dedupPath, "path", ".", "path to career-ops root directory")
	dedupCmd.Flags().BoolVar(&dedupDryRun, "dry-run", false, "preview changes without writing")
}

func runDedup(_ *cobra.Command, _ []string) (err error) {
	careerOpsPath := dedupPath
	dbPath := viper.GetString("db")
	ctx := context.Background()

	states.Init(careerOpsPath)

	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	clusters, err := r.FindDuplicates(ctx)
	if err != nil {
		return oops.Wrapf(err, "finding duplicates")
	}

	if len(clusters) == 0 {
		fmt.Println("  No duplicates found")
		return nil
	}

	removed := lo.SumBy(clusters, func(cluster db.DuplicateCluster) int {
		return processDedupCluster(ctx, r, cluster)
	})

	fmt.Printf("\n  %d duplicates removed\n", removed)

	if dedupDryRun {
		fmt.Println("(dry-run -- no changes written)")
	}

	return nil
}

// processDedupCluster handles a single duplicate cluster: picks keeper, promotes
// status, and deletes duplicates. Returns the number removed.
func processDedupCluster(
	ctx context.Context, r db.Repository, cluster db.DuplicateCluster,
) int {
	apps := cluster.Applications
	if len(apps) < 2 {
		return 0
	}

	// Sort by score descending -- keeper is the highest-scored entry.
	sort.Slice(apps, func(a, b int) bool {
		return apps[a].Score > apps[b].Score
	})
	keeper := apps[0]

	// Check if any duplicate has a more advanced pipeline status.
	bestEntry := lo.MaxBy(apps, func(a, b model.CareerApplication) bool {
		return states.StatusRank(a.Status) > states.StatusRank(b.Status)
	})

	if bestEntry.Status != keeper.Status {
		keeper.Status = bestEntry.Status
		fmt.Printf("  #%d: status promoted to %q (from #%d)\n",
			keeper.Number, bestEntry.Status, bestEntry.Number)

		if !dedupDryRun {
			if upsertErr := r.UpsertApplication(ctx, &keeper); upsertErr != nil {
				fmt.Printf("  warning: promoting #%d: %v\n", keeper.Number, upsertErr)
			}
		}
	}

	// Remove duplicates (all except keeper).
	lo.ForEach(apps[1:], func(dup model.CareerApplication, _ int) {
		fmt.Printf("  Remove #%d (%s -- %s, %.1f/5) -> kept #%d (%.1f/5)\n",
			dup.Number, dup.Company, dup.Role, dup.Score,
			keeper.Number, keeper.Score)

		if !dedupDryRun {
			if delErr := r.DeleteApplication(ctx, dup.Number); delErr != nil {
				fmt.Printf("  warning: deleting #%d: %v\n", dup.Number, delErr)
			}
		}
	})

	return len(apps) - 1
}
