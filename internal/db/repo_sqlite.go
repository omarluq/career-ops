package db

// sqliteRepo implements Repository backed by the DB type (ksql entities).
type sqliteRepo struct {
	d *DB
}

// NewSQLite returns a Repository backed by the given DB instance.
func NewSQLite(d *DB) Repository {
	return &sqliteRepo{d: d}
}

// Close releases the underlying database connection.
func (r *sqliteRepo) Close() error {
	return r.d.Close()
}
