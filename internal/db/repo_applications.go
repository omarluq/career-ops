package db

import (
	"context"

	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/tracker"
)

// ListApplications returns all tracked applications ordered by number.
func (r *sqliteRepo) ListApplications(ctx context.Context) ([]model.CareerApplication, error) {
	dbApps, err := r.d.ListApplications(ctx)
	if err != nil {
		return nil, err
	}
	return lo.Map(dbApps, func(a Application, _ int) model.CareerApplication {
		return a.ToModel()
	}), nil
}

// GetApplication returns a single application by its sequential number.
func (r *sqliteRepo) GetApplication(ctx context.Context, number int) (*model.CareerApplication, error) {
	dbApp, err := r.d.GetApplicationByNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	m := dbApp.ToModel()
	return &m, nil
}

// UpsertApplication creates or updates an application keyed by number.
func (r *sqliteRepo) UpsertApplication(ctx context.Context, app *model.CareerApplication) error {
	dbApp := ApplicationFromModel(app, tracker.NormalizeCompanyKey(app.Company))
	return r.d.UpsertApplication(ctx, &dbApp)
}

// DeleteApplication removes the application with the given number.
func (r *sqliteRepo) DeleteApplication(ctx context.Context, number int) error {
	dbApp, err := r.d.GetApplicationByNumber(ctx, number)
	if err != nil {
		return err
	}
	return r.d.DeleteApplication(ctx, dbApp.ID)
}

// FindDuplicates returns clusters of applications that appear to be duplicates.
func (r *sqliteRepo) FindDuplicates(ctx context.Context) ([]DuplicateCluster, error) {
	groups, err := r.d.FindDuplicates(ctx)
	if err != nil {
		return nil, err
	}

	return lo.Map(groups, func(group []Application, _ int) DuplicateCluster {
		apps := lo.Map(group, func(a Application, _ int) model.CareerApplication {
			return a.ToModel()
		})
		key := lo.Ternary(len(group) > 0, group[0].CompanyKey, "")
		return DuplicateCluster{CompanyKey: key, Applications: apps}
	}), nil
}

// SearchApplications performs a free-text search across key fields.
func (r *sqliteRepo) SearchApplications(ctx context.Context, query string) ([]model.CareerApplication, error) {
	dbApps, err := r.d.SearchApplications(ctx, query)
	if err != nil {
		return nil, err
	}
	return lo.Map(dbApps, func(a Application, _ int) model.CareerApplication {
		return a.ToModel()
	}), nil
}
