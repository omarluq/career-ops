-- +goose Up
CREATE TABLE IF NOT EXISTS user_profile (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name     TEXT NOT NULL DEFAULT '',
    email         TEXT NOT NULL DEFAULT '',
    phone         TEXT NOT NULL DEFAULT '',
    location      TEXT NOT NULL DEFAULT '',
    timezone      TEXT NOT NULL DEFAULT '',
    linkedin_url  TEXT NOT NULL DEFAULT '',
    github_url    TEXT NOT NULL DEFAULT '',
    portfolio_url TEXT NOT NULL DEFAULT '',
    twitter_url   TEXT NOT NULL DEFAULT '',
    headline      TEXT NOT NULL DEFAULT '',
    exit_story    TEXT NOT NULL DEFAULT '',
    superpowers   TEXT NOT NULL DEFAULT '[]',    -- JSON array of strings
    proof_points  TEXT NOT NULL DEFAULT '[]',    -- JSON array of {name, url, hero_metric}
    target_roles  TEXT NOT NULL DEFAULT '[]',    -- JSON array of strings (primary roles)
    archetypes    TEXT NOT NULL DEFAULT '[]',    -- JSON array of {name, level, fit}
    comp_target   TEXT NOT NULL DEFAULT '',
    comp_currency TEXT NOT NULL DEFAULT 'USD',
    comp_minimum  TEXT NOT NULL DEFAULT '',
    comp_location_flex TEXT NOT NULL DEFAULT '',
    country       TEXT NOT NULL DEFAULT '',
    city          TEXT NOT NULL DEFAULT '',
    visa_status   TEXT NOT NULL DEFAULT '',
    deal_breakers TEXT NOT NULL DEFAULT '[]',    -- JSON array of strings
    preferences   TEXT NOT NULL DEFAULT '{}',    -- JSON object for misc preferences
    created_at    DATETIME DEFAULT (datetime('now')),
    updated_at    DATETIME DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS user_profile;
