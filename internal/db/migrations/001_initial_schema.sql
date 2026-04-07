-- +goose Up
CREATE TABLE IF NOT EXISTS applications (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    number        INTEGER UNIQUE NOT NULL,
    date          TEXT NOT NULL DEFAULT '',
    company       TEXT NOT NULL DEFAULT '',
    company_key   TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT '',
    score         REAL NOT NULL DEFAULT 0,
    score_raw     TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'evaluated',
    has_pdf       INTEGER NOT NULL DEFAULT 0,
    report_number TEXT NOT NULL DEFAULT '',
    report_path   TEXT NOT NULL DEFAULT '',
    job_url       TEXT NOT NULL DEFAULT '',
    notes         TEXT NOT NULL DEFAULT '',
    archetype     TEXT NOT NULL DEFAULT '',
    tldr          TEXT NOT NULL DEFAULT '',
    remote        TEXT NOT NULL DEFAULT '',
    comp_estimate TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_apps_company_key ON applications(company_key);
CREATE INDEX IF NOT EXISTS idx_apps_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_apps_score ON applications(score DESC);
CREATE INDEX IF NOT EXISTS idx_apps_company_role ON applications(company_key, role);

CREATE TABLE IF NOT EXISTS pipeline_entries (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    url          TEXT NOT NULL UNIQUE,
    source       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'pending',
    added_at     TEXT NOT NULL DEFAULT (datetime('now')),
    processed_at TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_pipeline_status ON pipeline_entries(status);

CREATE TABLE IF NOT EXISTS scan_history (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    url        TEXT NOT NULL,
    portal     TEXT NOT NULL DEFAULT '',
    title      TEXT NOT NULL DEFAULT '',
    company    TEXT NOT NULL DEFAULT '',
    scanned_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(url, portal)
);

CREATE INDEX IF NOT EXISTS idx_scan_url ON scan_history(url);

CREATE TABLE IF NOT EXISTS evaluations (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id  INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    batch_id        TEXT NOT NULL DEFAULT '',
    evaluation_date TEXT NOT NULL DEFAULT (datetime('now')),
    raw_score       REAL NOT NULL DEFAULT 0,
    archetype       TEXT NOT NULL DEFAULT '',
    tldr            TEXT NOT NULL DEFAULT '',
    remote          TEXT NOT NULL DEFAULT '',
    comp_estimate   TEXT NOT NULL DEFAULT '',
    report_path     TEXT NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE IF EXISTS evaluations;
DROP TABLE IF EXISTS scan_history;
DROP TABLE IF EXISTS pipeline_entries;
DROP TABLE IF EXISTS applications;
