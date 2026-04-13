package db

import (
	"context"

	"github.com/samber/lo"
	"github.com/samber/oops"
)

// OpenAndMigrate opens a SQLite database at the given path using the db package,
// enables WAL mode, foreign keys, and runs all pending goose migrations.
// Returns the DB wrapper for use with NewSQLite.
func OpenAndMigrate(_ context.Context, path string) (*DB, error) {
	d, err := Open(path)
	if err != nil {
		return nil, oops.Wrapf(err, "opening database at %s", path)
	}

	if err := d.RunMigrations(); err != nil {
		if closeErr := d.Close(); closeErr != nil {
			return nil, oops.Wrapf(err, "running migrations (close error: %v)", closeErr)
		}
		return nil, oops.Wrapf(err, "running migrations")
	}

	// Verify migration succeeded by checking table existence.
	raw := d.SQL()
	requiredTables := []string{"applications", "pipeline_entries", "scan_history"}
	missing := lo.Filter(requiredTables, func(table string, _ int) bool {
		var name string
		qErr := raw.QueryRowContext(context.Background(),
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		return qErr != nil
	})

	if len(missing) > 0 {
		if closeErr := d.Close(); closeErr != nil {
			return nil, oops.Wrapf(closeErr, "closing after migration check failure")
		}
		return nil, oops.Errorf("migration incomplete: missing tables %v", missing)
	}

	return d, nil
}
