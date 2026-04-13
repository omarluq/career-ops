-- +goose Up
CREATE TABLE IF NOT EXISTS submissions (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    application_number INTEGER NOT NULL,
    portal             TEXT NOT NULL,
    jd_url             TEXT NOT NULL,
    cv_pdf_path        TEXT NOT NULL DEFAULT '',
    status             TEXT NOT NULL DEFAULT 'pending',
    confirm_url        TEXT NOT NULL DEFAULT '',
    error_message      TEXT NOT NULL DEFAULT '',
    attempts           INTEGER NOT NULL DEFAULT 0,
    submitted_at       DATETIME,
    created_at         DATETIME DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_submissions_app ON submissions(application_number);
CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status);

-- +goose Down
DROP INDEX IF EXISTS idx_submissions_status;
DROP INDEX IF EXISTS idx_submissions_app;
DROP TABLE IF EXISTS submissions;
