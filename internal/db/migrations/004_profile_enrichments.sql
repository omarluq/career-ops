-- +goose Up
CREATE TABLE IF NOT EXISTS profile_enrichments (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type    TEXT NOT NULL,     -- github, linkedin, blog, conversation, article, portfolio
    source_url     TEXT NOT NULL DEFAULT '',
    source_title   TEXT NOT NULL DEFAULT '',
    extracted_data TEXT NOT NULL DEFAULT '{}',  -- JSON: what was extracted
    applied_fields TEXT NOT NULL DEFAULT '[]',  -- JSON array: which profile fields were updated
    confidence     TEXT NOT NULL DEFAULT 'medium', -- low, medium, high
    applied        BOOLEAN NOT NULL DEFAULT 0,
    applied_at     DATETIME,
    created_at     DATETIME DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_enrichments_source_type ON profile_enrichments(source_type);
CREATE INDEX IF NOT EXISTS idx_enrichments_applied ON profile_enrichments(applied);

-- +goose Down
DROP INDEX IF EXISTS idx_enrichments_applied;
DROP INDEX IF EXISTS idx_enrichments_source_type;
DROP TABLE IF EXISTS profile_enrichments;
