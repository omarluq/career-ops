# Architecture

## System Overview

```
                    ┌─────────────────────────────────┐
                    │         Claude Code Agent        │
                    │   (reads CLAUDE.md + modes/*.md) │
                    └──────────┬──────────────────────┘
                               │
            ┌──────────────────┼──────────────────────┐
            │                  │                       │
     ┌──────▼──────┐   ┌──────▼──────┐   ┌───────────▼────────┐
     │ Single Eval  │   │ Portal Scan │   │   Batch Process    │
     │ (auto-pipe)  │   │  (scan.md)  │   │   (batch-runner)   │
     └──────┬──────┘   └──────┬──────┘   └───────────┬────────┘
            │                  │                       │
            │           ┌──────▼──────┐          ┌────▼─────┐
            │           │  SQLite DB  │          │ N workers│
            │           │  (pipeline) │          │ (claude -p)
            │           └─────────────┘          └────┬─────┘
            │                                          │
     ┌──────▼──────────────────────────────────────────▼──────┐
     │                    Output Pipeline                      │
     │  ┌──────────┐  ┌────────────┐  ┌───────────────────┐  │
     │  │ Report.md│  │  PDF (HTML  │  │ SQLite upsert     │  │
     │  │ (A-F eval)│  │  → chromedp)│  │ (repo layer)      │  │
     │  └──────────┘  └────────────┘  └───────────────────┘  │
     └────────────────────────────────────────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   SQLite database    │
                    │  (canonical store)   │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │     MCP Server       │
                    │  (tools + resources) │
                    └─────────────────────┘
```

## Evaluation Flow (Single Offer)

1. **Input**: User pastes JD text or URL
2. **Extract**: chromedp/WebFetch extracts JD from URL
3. **Classify**: Detect archetype (1 of 6 types)
4. **Evaluate**: 6 blocks (A-F):
   - A: Role summary
   - B: CV match (gaps + mitigation)
   - C: Level strategy
   - D: Comp research (WebSearch)
   - E: CV personalization plan
   - F: Interview prep (STAR stories)
5. **Score**: Weighted average across 10 dimensions (1-5)
6. **Report**: Save as `reports/{num}-{company}-{date}.md`
7. **PDF**: Generate ATS-optimized CV via chromedp
8. **Track**: Upsert into SQLite via the repo layer (+ optional TSV for batch merge)

## Batch Processing

The batch system processes multiple offers in parallel:

```
batch-input.tsv    →  batch-runner.sh  →  N × claude -p workers
(id, url, source)     (orchestrator)       (self-contained prompt)
                           │
                    batch-state.tsv
                    (tracks progress)
```

Each worker is a headless Claude instance (`claude -p`) that receives the full `batch-prompt.md` as context. Workers produce:
- Report .md
- PDF
- Tracker TSV line

The orchestrator manages parallelism, state, retries, and resume.

## Data Flow

```
cv.md                      →  Evaluation context
article-digest.md          →  Proof points for matching
config/profile.yml         →  Candidate identity
portals.yml                →  Scanner configuration
internal/states/           →  Canonical status values
templates/cv-template.html →  PDF generation template
```

Data flows through SQLite via the repository layer (`internal/repo`). The MCP server
exposes this data externally through:

- **Tools**: `search`, `list`, `get`, `update_status`, `add_to_pipeline`, `pipeline_status`
- **Resources**: `applications://list`, `metrics://pipeline`

An import/export bridge allows migration between the legacy markdown tracker and SQLite:

```
data/applications.md  ←──export──  SQLite  ──import──→  data/applications.md
```

## File Naming Conventions

- Reports: `{###}-{company-slug}-{YYYY-MM-DD}.md` (3-digit zero-padded)
- PDFs: `cv-candidate-{company-slug}-{YYYY-MM-DD}.pdf`
- Tracker TSVs: `batch/tracker-additions/{id}.tsv`

## Pipeline Integrity

CLI commands maintain data consistency:

| Command | Purpose |
|---------|---------|
| `career-ops verify` | Health check: statuses, duplicates, links |
| `career-ops dedup` | Removes duplicate entries by company+role |
| `career-ops normalize` | Maps status aliases to canonical values |
| `career-ops merge` | Merges batch TSV additions |
| `career-ops import` | Import markdown data into SQLite |
| `career-ops export` | Export SQLite to markdown |
| `career-ops sync-check` | Validates setup consistency |

## Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/db` | SQLite entity layer (Application, PipelineEntry, ScanRecord, Evaluation) |
| `internal/repo` | Repository interface + SQLite implementation |
| `internal/mcp` | MCP server with 6 tools + 2 resources |
| `internal/worker` | Generic Pool[T,R], FanOut, RunBatch with errgroup |
| `internal/scanner` | Concurrent portal scanner with chromedp |
| `internal/tracker` | Markdown table parsing + metrics computation |
| `internal/states` | Canonical status management + normalization |
| `internal/closer` | Deferred close error handling (Guard pattern) |
| `internal/model` | Shared domain models |
| `internal/ui` | Bubble Tea TUI (screens + theme) |
| `internal/vinfo` | Build version injection |

## Dashboard TUI

The `career-ops dashboard` subcommand launches a Bubble Tea TUI that visualizes the pipeline:

- Filter tabs: All, Evaluada, Aplicado, Entrevista, Top >=4, No Aplicar
- Sort modes: Score, Date, Company, Status
- Grouped/flat view
- Lazy-loaded report previews
- Inline status picker
