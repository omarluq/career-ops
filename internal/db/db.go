// Package db provides SQLite database access for the career-ops pipeline.
package db

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/vingarcia/ksql"
	"github.com/vingarcia/ksql/sqldialect"

	_ "modernc.org/sqlite" // register sqlite driver for database/sql
)

//go:embed migrations/*.sql
var migrations embed.FS

// sqlDBAdapter wraps *sql.DB to satisfy ksql.DBAdapter interface,
// bridging the modernc.org/sqlite driver with ksql's type system.
type sqlDBAdapter struct {
	db *sql.DB
}

// ExecContext executes a query without returning rows.
func (a *sqlDBAdapter) ExecContext(ctx context.Context, query string, args ...interface{}) (ksql.Result, error) {
	return a.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows.
func (a *sqlDBAdapter) QueryContext(ctx context.Context, query string, args ...interface{}) (ksql.Rows, error) {
	return a.db.QueryContext(ctx, query, args...)
}

// DB wraps both the underlying sql.DB connection and a ksql.DB for
// type-safe query building.
type DB struct {
	sql  *sql.DB
	ksql ksql.DB
}

// Open connects to a SQLite database at dbPath, enables WAL mode,
// foreign keys, and sets a busy timeout.
func Open(dbPath string) (*DB, error) {
	raw, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, oops.Wrapf(err, "opening database at %s", dbPath)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}

	lo.ForEach(pragmas, func(p string, _ int) {
		if err != nil {
			return
		}
		if _, execErr := raw.ExecContext(context.Background(), p); execErr != nil {
			err = oops.Wrapf(execErr, "executing %s", p)
		}
	})
	if err != nil {
		if closeErr := raw.Close(); closeErr != nil {
			return nil, oops.Wrapf(err, "closing after pragma failure (close error: %v)", closeErr)
		}
		return nil, err
	}

	adapter := &sqlDBAdapter{db: raw}
	k, err := ksql.NewWithAdapter(adapter, sqldialect.Sqlite3Dialect{})
	if err != nil {
		if closeErr := raw.Close(); closeErr != nil {
			return nil, oops.Wrapf(err, "closing after adapter failure (close error: %v)", closeErr)
		}
		return nil, oops.Wrapf(err, "creating ksql adapter")
	}

	return &DB{sql: raw, ksql: k}, nil
}

// Close releases the database connection.
func (d *DB) Close() error {
	errs := lo.Compact([]error{
		d.ksql.Close(),
		d.sql.Close(),
	})
	if len(errs) > 0 {
		return oops.Wrapf(errs[0], "closing database")
	}
	return nil
}

// SQL returns the underlying *sql.DB for direct access.
func (d *DB) SQL() *sql.DB {
	return d.sql
}

// KSQL returns the ksql.DB instance for type-safe queries.
func (d *DB) KSQL() ksql.DB {
	return d.ksql
}

// RunMigrations applies all pending goose migrations embedded in this package.
func (d *DB) RunMigrations() error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return oops.Wrapf(err, "setting goose dialect")
	}

	if err := goose.Up(d.sql, "migrations"); err != nil {
		return oops.Wrapf(err, "running migrations")
	}

	return nil
}
