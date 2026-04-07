# Career-Ops Evolution Plan

## Three Pillars

1. **Upstream Sync** -- fold santifer/career-ops improvements
2. **YAML -> SQLite** -- modern data layer with KSQL
3. **Concurrency & Agent Teams** -- framework on steroids

---

## Phase 0: Upstream Sync (Quick Wins)

No architecture changes. Documentation, config, and mode files only.

| # | Task | Priority | Files | Depends |
|---|------|----------|-------|---------|
| 0.1 | System/user layer split | HIGH | `modes/_shared.md`, create `modes/_profile.template.md`, create `DATA_CONTRACT.md` | -- |
| 0.2 | Low-fit threshold 3.0 -> 4.0 | HIGH | `CLAUDE.md` | -- |
| 0.3 | Onboarding Step 5: "Get to know the user" | HIGH | `CLAUDE.md` | 0.1 |
| 0.4 | German language modes | MEDIUM | Create `modes/de/` | -- |
| 0.5 | Dashboard crash fix on empty apps | MEDIUM | `internal/ui/screens/pipeline.go` | -- |
| 0.6 | Batch Playwright fallback rule | MEDIUM | `CLAUDE.md` | -- |
| 0.7 | DACH/European companies in portals | MEDIUM | `templates/portals.example.yml` | -- |
| 0.8 | Dual-track example | LOW | `config/profile.example.yml` | -- |
| 0.9 | CONTRIBUTING.md issue-first rule | LOW | `CONTRIBUTING.md` | -- |

---

## Phase 1: SQLite Foundation

Introduce SQLite as data backend with schema, migrations, import/export.

| # | Task | Files | Depends |
|---|------|-------|---------|
| 1.1 | Add SQLite + KSQL + goose deps | `go.mod` | -- |
| 1.2 | Design schema + migrations | `internal/db/schema.go`, `internal/db/migrations/001_initial.sql` | 1.1 |
| 1.3 | DB connection + lifecycle | `internal/db/db.go`, `internal/db/ksql.go` | 1.1, 1.2 |
| 1.4 | Import command (markdown -> SQLite) | `cmd/career-ops/cmd_import.go`, `internal/db/import.go` | 1.2, 1.3 |
| 1.5 | Export command (SQLite -> markdown) | `cmd/career-ops/cmd_export.go`, `internal/db/export.go` | 1.2, 1.3 |
| 1.6 | Add `--db` flag to root cmd | `cmd/career-ops/root.go` | 1.3 |

### Schema

```sql
applications (
    id INTEGER PRIMARY KEY,
    number INTEGER UNIQUE NOT NULL,
    date TEXT NOT NULL,
    company TEXT NOT NULL,
    company_key TEXT NOT NULL,     -- NormalizeCompanyKey output
    role TEXT NOT NULL,
    score REAL DEFAULT 0,
    score_raw TEXT DEFAULT '',
    status TEXT NOT NULL CHECK(status IN (...canonical states...)),
    has_pdf BOOLEAN DEFAULT FALSE,
    report_path TEXT DEFAULT '',
    job_url TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    archetype TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
)

pipeline_entries (url UNIQUE, source, status, added_at, processed_at)
scan_history (url + portal UNIQUE, title, company, scanned_at)
states (id PK, label, description, aliases)
evaluations (application_id FK, batch_id, raw_score, report_path)
```

FTS: `CREATE VIRTUAL TABLE applications_fts USING fts5(company, role, notes)`

Driver: `modernc.org/sqlite` (pure Go, CGO-free)

---

## Phase 2: Port Commands to SQLite

| # | Task | Current Logic -> New Logic | Depends |
|---|------|---------------------------|---------|
| 2.1 | Repository interface | Create `internal/repo/repo.go` (interface), `sqlite.go`, `markdown.go` (legacy) | Phase 1 |
| 2.2 | Port `verify` | Regex line scanning -> SQL queries with CHECK constraints | 2.1 |
| 2.3 | Port `dedup` | GroupBy + RoleMatch -> SQL grouping + Go fuzzy matching + DELETE | 2.1 |
| 2.4 | Port `merge` | TSV parse + line append -> `INSERT ON CONFLICT` | 2.1 |
| 2.5 | Port `normalize` | Line-by-line regex -> batch UPDATE with `states.Normalize()` | 2.1 |
| 2.6 | Port dashboard | `ParseApplications()` -> `SELECT` with filters | 2.1 |
| 2.7 | Port `sync-check` | File scanning -> DB + file validation | 2.1 |

Tasks 2.2-2.7 are all parallel after 2.1.

---

## Phase 3: Concurrency Layer

| # | Task | Approach | Depends |
|---|------|----------|---------|
| 3.1 | Worker pool infra | `errgroup` + `SetLimit(N)` + `mo.Result` | Phase 2 |
| 3.2 | Concurrent portal scanning | N Chrome contexts, dedup via `INSERT ON CONFLICT` | 3.1 |
| 3.3 | Concurrent batch eval | Fan-out workers for pipeline entries | 3.1 |
| 3.4 | Async dashboard loading | Goroutine pool for report summaries | 2.6 |
| 3.5 | Concurrent PDF generation | Shared Chrome browser, N render contexts | 3.1 |

Tasks 3.2-3.5 are all parallel after 3.1.

---

## Phase 4: Agent Orchestration

| # | Task | Description | Depends |
|---|------|-------------|---------|
| 4.1 | `agent-eval <url>` | Agent team: scraper -> evaluator -> CV generator | Phase 3 |
| 4.2 | `agent-batch` | Parallel batch with agent teams | 4.1 |
| 4.3 | `agent-scan` | Agent-powered JD extraction after portal scan | 3.2, 4.1 |
| 4.4 | MCP server mode | Expose career-ops as MCP tool server | Phase 2 |
| 4.5 | Story bank extraction | Mine STAR+R stories from reports | 4.1 |

4.1 and 4.4 can start in parallel. 4.2, 4.3, 4.5 after 4.1.

---

## Dependency Graph

```
Phase 0 (do first, independent)
    |
Phase 1: 1.1 -> 1.2 -> 1.3 -> [1.4 | 1.5 | 1.6]
    |
Phase 2: 2.1 -> [2.2 | 2.3 | 2.4 | 2.5 | 2.6 | 2.7]
    |
Phase 3: 3.1 -> [3.2 | 3.3 | 3.4 | 3.5]
    |
Phase 4: [4.1 | 4.4] -> [4.2 | 4.3 | 4.5]
```

## Key Decisions

- **SQLite driver:** `modernc.org/sqlite` (pure Go, CGO-free)
- **Query builder:** KSQL for type-safe, raw SQL for complex aggregation/FTS
- **Migrations:** goose
- **Repository pattern:** Interface with SQLite (primary) + markdown (legacy `--legacy`)
- **Concurrency:** `errgroup.SetLimit()` + `mo.Result`
- **Backward compat:** `--legacy` flag, `export` command always available
- **DB location:** `$CAREER_OPS_ROOT/career-ops.db` (gitignored)

## Risk Matrix

| Phase | Risk | Key Concern | Mitigation |
|-------|------|-------------|------------|
| 0 | Low | Merge conflicts | Review diffs carefully |
| 1 | Medium | Schema design | Prototype first, goose migrations for evolution |
| 2 | Low-Medium | Interface design | Derive from command requirements |
| 3 | Medium-High | Chrome resource exhaustion | Hard limits, circuit breaker |
| 4 | High | Agent orchestration | Start sequential, parallelize incrementally |
