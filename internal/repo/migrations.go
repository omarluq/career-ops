package repo

import (
	"context"
	"database/sql"

	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/db"
)

// OpenAndMigrate opens a SQLite database at the given path using the db package,
// enables WAL mode, foreign keys, and runs all pending goose migrations.
// Returns the raw *sql.DB for use with NewSQLite.
func OpenAndMigrate(ctx context.Context, path string) (*sql.DB, error) {
	_ = ctx // reserved for future context-aware migration support

	d, err := db.Open(path)
	if err != nil {
		return nil, oops.Wrapf(err, "opening database at %s", path)
	}

	if err := d.RunMigrations(); err != nil {
		if closeErr := d.Close(); closeErr != nil {
			return nil, oops.Wrapf(err, "running migrations (close error: %v)", closeErr)
		}
		return nil, oops.Wrapf(err, "running migrations")
	}

	// Extract the raw sql.DB so callers can use it directly.
	// The db.DB wrapper is no longer needed after migration.
	raw := d.SQL()

	// Verify migration succeeded by checking table existence.
	requiredTables := []string{"applications", "pipeline_entries", "scan_history"}
	missing := lo.Filter(requiredTables, func(table string, _ int) bool {
		var name string
		qErr := raw.QueryRowContext(ctx,
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		return qErr != nil
	})

	if len(missing) > 0 {
		if closeErr := raw.Close(); closeErr != nil {
			return nil, oops.Wrapf(closeErr, "closing after migration check failure")
		}
		return nil, oops.Errorf("migration incomplete: missing tables %v", missing)
	}

	return raw, nil
}
