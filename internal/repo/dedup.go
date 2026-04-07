package repo

import (
	"context"
	"sort"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
)

// DedupResult holds the outcome of deduplication.
type DedupResult struct {
	Promoted      []PromotedStatus
	ClustersFound int
	Removed       int
}

// PromotedStatus records an application whose status was promoted during dedup.
type PromotedStatus struct {
	Company   string
	OldStatus string
	NewStatus string
	Number    int
}

// Dedup finds and resolves duplicate applications.
// Within each duplicate cluster, it keeps the highest-scored entry, promotes
// status if a discarded duplicate was further in the pipeline, and deletes
// the rest.
func Dedup(ctx context.Context, r Repository) (*DedupResult, error) {
	clusters, err := r.FindDuplicates(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "finding duplicates")
	}

	result := &DedupResult{}

	roleClusters := lo.FlatMap(clusters, func(c DuplicateCluster, _ int) [][]model.CareerApplication {
		return buildAppRoleClusters(c.Applications)
	})

	lo.ForEach(roleClusters, func(rc []model.CareerApplication, _ int) {
		result.ClustersFound++
		if resolveErr := resolveCluster(ctx, r, rc, result); resolveErr != nil {
			if err == nil {
				err = oops.Wrapf(resolveErr, "resolving cluster")
			}
		}
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// buildAppRoleClusters groups applications within a company cluster by fuzzy
// role match. Only groups with 2+ entries are returned.
func buildAppRoleClusters(apps []model.CareerApplication) [][]model.CareerApplication {
	processed := make(map[int]bool)
	var clusters [][]model.CareerApplication

	lo.ForEach(apps, func(_ model.CareerApplication, i int) {
		if processed[i] {
			return
		}

		matches := lo.FilterMap(apps[i+1:], func(other model.CareerApplication, offset int) (int, bool) {
			j := i + 1 + offset
			if processed[j] {
				return 0, false
			}
			return j, tracker.RoleMatch(apps[i].Role, other.Role)
		})

		if len(matches) == 0 {
			processed[i] = true
			return
		}

		// Build the cluster: seed + all matches.
		cluster := []model.CareerApplication{apps[i]}
		processed[i] = true
		lo.ForEach(matches, func(j int, _ int) {
			cluster = append(cluster, apps[j])
			processed[j] = true
		})

		clusters = append(clusters, cluster)
	})

	return clusters
}

// resolveCluster picks the highest-scored entry as keeper, promotes its status
// if any duplicate was further along, and deletes the rest.
func resolveCluster(
	ctx context.Context,
	r Repository,
	cluster []model.CareerApplication,
	result *DedupResult,
) error {
	sort.Slice(cluster, func(a, b int) bool {
		return tracker.ParseScore(cluster[a].ScoreRaw) > tracker.ParseScore(cluster[b].ScoreRaw)
	})

	keeper := cluster[0]

	if promoted := promoteStatus(&keeper, cluster[1:]); promoted != nil {
		result.Promoted = append(result.Promoted, *promoted)
		if err := r.UpsertApplication(ctx, &keeper); err != nil {
			return oops.Wrapf(err, "promoting status for #%d", keeper.Number)
		}
	}

	// Delete all duplicates (everything except the keeper).
	errs := lo.FilterMap(cluster[1:], func(dup model.CareerApplication, _ int) (error, bool) {
		if err := r.DeleteApplication(ctx, dup.Number); err != nil {
			return oops.Wrapf(err, "deleting duplicate #%d", dup.Number), true
		}
		result.Removed++
		return nil, false
	})

	return lo.Ternary(len(errs) > 0, errs[0], nil)
}

// promoteStatus checks whether any duplicate has a more advanced pipeline status
// than the keeper. If so, it updates the keeper's status in place and returns the
// promotion record. Returns nil if no promotion is needed.
func promoteStatus(keeper *model.CareerApplication, dupes []model.CareerApplication) *PromotedStatus {
	best := lo.Max(lo.Map(dupes, func(dup model.CareerApplication, _ int) int {
		return states.StatusRank(dup.Status)
	}))

	keeperRank := states.StatusRank(keeper.Status)
	if best <= keeperRank {
		return nil
	}

	// Find the status string with the best rank.
	bestDup, _ := lo.Find(dupes, func(dup model.CareerApplication) bool {
		return states.StatusRank(dup.Status) == best
	})

	promoted := &PromotedStatus{
		Number:    keeper.Number,
		Company:   keeper.Company,
		OldStatus: keeper.Status,
		NewStatus: bestDup.Status,
	}

	keeper.Status = bestDup.Status

	return promoted
}

// DedupDryRun returns duplicate clusters and what would happen without modifying data.
func DedupDryRun(ctx context.Context, r Repository) (*DedupResult, error) {
	clusters, err := r.FindDuplicates(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "finding duplicates")
	}

	result := &DedupResult{}

	roleClusters := lo.FlatMap(clusters, func(c DuplicateCluster, _ int) [][]model.CareerApplication {
		return buildAppRoleClusters(c.Applications)
	})

	lo.ForEach(roleClusters, func(rc []model.CareerApplication, _ int) {
		result.ClustersFound++
		dryResolveCluster(rc, result)
	})

	return result, nil
}

// dryResolveCluster simulates cluster resolution without modifying the repository.
func dryResolveCluster(cluster []model.CareerApplication, result *DedupResult) {
	sort.Slice(cluster, func(a, b int) bool {
		return tracker.ParseScore(cluster[a].ScoreRaw) > tracker.ParseScore(cluster[b].ScoreRaw)
	})

	keeper := cluster[0]

	if promoted := promoteStatus(&keeper, cluster[1:]); promoted != nil {
		result.Promoted = append(result.Promoted, *promoted)
	}

	result.Removed += lo.Max([]int{len(cluster) - 1, 0})
}
