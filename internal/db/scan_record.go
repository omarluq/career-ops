package db

import (
	"context"

	"github.com/samber/oops"
	"github.com/vingarcia/ksql"
)

// ScanRecord maps to the scan_history table.
type ScanRecord struct {
	URL        string `ksql:"url"`
	Portal     string `ksql:"portal"`
	Title      string `ksql:"title"`
	Company    string `ksql:"company"`
	ScannedAt  string `ksql:"scanned_at"`
	ksql.Table `ksql:"scan_history"`
	ID         int `ksql:"id"`
}

// InsertScanRecord adds a scan history entry, ignoring duplicates.
// Uses raw SQL for ON CONFLICT since ksql does not support it natively.
func (d *DB) InsertScanRecord(ctx context.Context, url, portal, title, company string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO scan_history (url, portal, title, company)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(url, portal) DO NOTHING`, url, portal, title, company)

	return oops.Wrapf(err, "inserting scan record url=%s portal=%s", url, portal)
}

// ListScanRecords returns all scan history records ordered by scanned_at descending.
func (d *DB) ListScanRecords(ctx context.Context) ([]ScanRecord, error) {
	var records []ScanRecord
	err := d.ksql.Query(ctx, &records, `
		SELECT id, url, portal, title, company, scanned_at
		FROM scan_history
		ORDER BY scanned_at DESC`)
	if err != nil {
		return nil, oops.Wrapf(err, "listing scan records")
	}
	return records, nil
}

// HasBeenScanned checks whether a URL+portal combination was already scanned.
func (d *DB) HasBeenScanned(ctx context.Context, url, portal string) (bool, error) {
	var count int
	err := d.sql.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM scan_history WHERE url = ? AND portal = ?`,
		url, portal).Scan(&count)
	if err != nil {
		return false, oops.Wrapf(err, "checking scan history url=%s portal=%s", url, portal)
	}
	return count > 0, nil
}
