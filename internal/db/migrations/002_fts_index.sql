-- +goose Up
CREATE VIRTUAL TABLE IF NOT EXISTS applications_fts USING fts5(
    company, role, notes, archetype, tldr,
    content=applications, content_rowid=id
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS applications_ai AFTER INSERT ON applications BEGIN
    INSERT INTO applications_fts(rowid, company, role, notes, archetype, tldr)
    VALUES (new.id, new.company, new.role, new.notes, new.archetype, new.tldr);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS applications_ad AFTER DELETE ON applications BEGIN
    INSERT INTO applications_fts(applications_fts, rowid, company, role, notes, archetype, tldr)
    VALUES ('delete', old.id, old.company, old.role, old.notes, old.archetype, old.tldr);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS applications_au AFTER UPDATE ON applications BEGIN
    INSERT INTO applications_fts(applications_fts, rowid, company, role, notes, archetype, tldr)
    VALUES ('delete', old.id, old.company, old.role, old.notes, old.archetype, old.tldr);
    INSERT INTO applications_fts(rowid, company, role, notes, archetype, tldr)
    VALUES (new.id, new.company, new.role, new.notes, new.archetype, new.tldr);
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS applications_au;
DROP TRIGGER IF EXISTS applications_ad;
DROP TRIGGER IF EXISTS applications_ai;
DROP TABLE IF EXISTS applications_fts;
