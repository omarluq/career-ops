package repo

import (
	"context"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
)

// NormalizeResult holds the outcome of status normalization.
type NormalizeResult struct {
	Unknown []UnknownStatus
	Changed int
}

// UnknownStatus records an application whose status could not be mapped
// to any canonical form.
type UnknownStatus struct {
	Raw    string
	Number int
}

// NormalizeStatuses normalizes all application statuses to canonical form.
// For each application, it runs states.Normalize and updates the status if
// the canonical form differs from the current value. Statuses that cannot
// be mapped to a canonical ID are collected in the Unknown list.
func NormalizeStatuses(ctx context.Context, r Repository) (*NormalizeResult, error) {
	apps, err := r.ListApplications(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing applications")
	}

	result := &NormalizeResult{}

	// Partition into normalizable and unknown.
	result.Unknown = lo.FilterMap(apps, func(app model.CareerApplication, _ int) (UnknownStatus, bool) {
		normalizedID := states.Normalize(app.Status)
		if states.IsCanonical(normalizedID) {
			return UnknownStatus{}, false
		}
		return UnknownStatus{Number: app.Number, Raw: app.Status}, true
	})

	// Filter to apps that need updating (canonical label differs from current).
	toUpdate := lo.Filter(apps, func(app model.CareerApplication, _ int) bool {
		normalizedID := states.Normalize(app.Status)
		return states.IsCanonical(normalizedID) && states.Label(normalizedID) != app.Status
	})

	// Apply updates.
	lo.ForEach(toUpdate, func(app model.CareerApplication, _ int) {
		normalizedID := states.Normalize(app.Status)
		app.Status = states.Label(normalizedID)
		if updateErr := r.UpsertApplication(ctx, &app); updateErr != nil {
			// Capture the first error encountered.
			if err == nil {
				err = oops.Wrapf(updateErr, "updating status for #%d", app.Number)
			}
		} else {
			result.Changed++
		}
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
