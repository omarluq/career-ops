# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Career-Ops -- AI Job Search Pipeline

## Development Commands

```bash
# Build
mise exec -- task build              # Build binary with version injection
mise exec -- task install            # Install to $GOPATH/bin

# Pipeline integrity
career-ops verify       # Health check (statuses, dupes, links)
career-ops merge        # Merge batch TSV additions into DB
career-ops dedup        # Remove duplicate entries
career-ops normalize    # Map aliases to canonical statuses
career-ops sync-check   # Validate setup consistency

# Profile management
career-ops profile show              # Display user profile
career-ops profile set <field> <val> # Update a profile field
career-ops profile enrichments       # List pending enrichments
career-ops profile apply <id>        # Apply a pending enrichment

# Data migration (legacy)
career-ops import                                         # Import markdown data into SQLite
career-ops export [--format applications|pipeline|scan-history] [--output file]  # Export SQLite to markdown

# PDF generation (requires Chrome/Chromium installed)
career-ops pdf <input.html> <output.pdf> [--format=letter|a4]

# Batch processing
career-ops batch        # (not yet implemented)

# Dashboard TUI
career-ops dashboard [--path .]

# MCP server (includes profile tools)
career-ops mcp          # Start MCP server (stdio transport)

# Development
mise exec -- task test               # Run tests
mise exec -- task lint               # golangci-lint
mise exec -- task fmt                # Auto-fix lint issues
mise exec -- task ci                 # lint + test
```

## Origin

This system was built and used by [santifer](https://santifer.io) to evaluate 740+ job offers, generate 100+ tailored CVs, and land a Head of Applied AI role. The archetypes, scoring logic, negotiation scripts, and proof point structure all reflect his specific career search in AI/automation roles.

The portfolio that goes with this system is also open source: [cv-santiago](https://github.com/santifer/cv-santiago).

**It will work out of the box, but it's designed to be made yours.** If the archetypes don't match your career, the modes are in the wrong language, or the scoring doesn't fit your priorities -- just ask. You (Claude) can edit any file in this system. The user says "change the archetypes to data engineering roles" and you do it. That's the whole point.

## What is career-ops

AI-powered job search automation built on Claude Code: pipeline tracking, offer evaluation, CV generation, portal scanning, batch processing.

### Main Files

| File | Function |
|------|----------|
| `career-ops.db` | **SQLite database — single source of truth** for applications, pipeline, scan history, profile, enrichments, submissions |
| `portals.yml` | Query and company config for scanner |
| `templates/cv-template.html` | HTML template for CVs |
| `cv.md` | Canonical CV (content, not data) |
| `article-digest.md` | Compact proof points from portfolio (optional) |
| `interview-prep/story-bank.md` | Accumulated STAR+R stories across evaluations |
| `reports/` | Evaluation reports (format: `{###}-{company-slug}-{YYYY-MM-DD}.md`) |
| `internal/db/` | SQLite database layer (entity files, migrations, validation) |
| `internal/repo/` | Repository interface + SQLite implementation |
| `internal/mcp/` | MCP server (tools + resources + profile tools) |
| `internal/jobboard/` | **Unified job board abstraction** — Board interface (search + apply), Registry, rate limiter |
| `internal/jobboard/boards/` | **100 board implementations** — ATS, aggregators, startup, remote/AI, freelance/intl |
| `internal/scanner/` | Portal scanner + legacy board adapters |
| `internal/applicator/` | High-throughput application submission pipeline |
| `internal/worker/` | Generic worker pool, fan-out, batch processing |
| `internal/closer/` | Deferred close error handling |
| `internal/ui/screens/` | TUI screens — pipeline table, kanban board, profile view, file viewer |

### Data Architecture

**SQLite is the single source of truth.** All application tracking, pipeline management, scan history, user profile, and enrichment data lives in `career-ops.db`. The legacy markdown files (`data/applications.md`, `data/pipeline.md`, `data/scan-history.tsv`) are no longer read at runtime — use `career-ops import` to migrate existing data and `career-ops export` to generate markdown snapshots.

### Job Board Architecture

The `internal/jobboard/` package provides a unified `Board` interface for both job discovery and application submission:

```go
type Board interface {
    Meta() BoardMeta          // static info: name, slug, category, capabilities
    Search(ctx, query) ([]SearchResult, error)  // discover jobs
    Apply(ctx, app) (ApplyResult, error)        // submit application
    HealthCheck(ctx) error                       // verify reachable
}
```

**Registry** holds all 100 boards, supports lookup by slug/category/capability, and `SearchAll()` fans out queries in parallel via errgroup.

**Board categories:** ATS (20), Aggregator (20), Startup (20), Remote/AI/Niche (20), Freelance/International (20).

**Capabilities:** `search`, `apply`, `api` (structured API), `scrape` (chromedp required).

```go
reg := jobboard.NewRegistry()
boards.RegisterAll(reg)                    // registers all 100 boards
results, _ := reg.SearchAll(ctx, query)    // parallel fan-out search
```

### TUI Dashboard Views

The dashboard (`career-ops dashboard`) has multiple view modes:

| View | File | Description |
|------|------|-------------|
| Pipeline (table) | `screens/pipeline.go` | Table view with tabs, sorting, filtering |
| Kanban (cards) | `screens/kanban.go` | Card-based kanban board grouped by status |
| Profile | `screens/profile.go` | User profile with collapsible sections |
| File viewer | `screens/viewer.go` | Report/file viewer with syntax highlighting |

### First Run — Onboarding (IMPORTANT)

**Before doing ANYTHING else, check if the system is set up.** Run these checks silently every time a session starts:

1. Does `cv.md` exist?
2. Does the user profile exist in the DB? (run `career-ops profile show` — if all fields are empty, profile needs setup)
3. Does `portals.yml` exist (not just templates/portals.example.yml)?

**If ANY of these is missing, enter onboarding mode.** Do NOT proceed with evaluations, scans, or any other mode until the basics are in place. Guide the user step by step:

#### Step 1: CV (required)
If `cv.md` is missing, ask:
> "I don't have your CV yet. You can either:
> 1. Paste your CV here and I'll convert it to markdown
> 2. Paste your LinkedIn URL and I'll extract the key info
> 3. Tell me about your experience and I'll draft a CV for you
>
> Which do you prefer?"

Create `cv.md` from whatever they provide. Make it clean markdown with standard sections (Summary, Experience, Projects, Education, Skills).

#### Step 2: Profile (required)
If the DB profile is empty (all fields blank), ask:
> "I need a few details to personalize the system:
> - Your full name and email
> - Your location and timezone
> - What roles are you targeting? (e.g., 'Senior Backend Engineer', 'AI Product Manager')
> - Your salary target range
>
> I'll set everything up for you."

Save their answers to the DB via `career-ops profile set <field> <value>` or the MCP `profile_update` tool. For archetypes, map their target roles to the closest matches and update `modes/_shared.md` if needed.

#### Step 3: Portals (recommended)
If `portals.yml` is missing:
> "I'll set up the job scanner with 45+ pre-configured companies. Want me to customize the search keywords for your target roles?"

Copy `templates/portals.example.yml` → `portals.yml`. If they gave target roles in Step 2, update `title_filter.positive` to match.

#### Step 4: Database
The SQLite database is auto-created and migrated on first use — no manual setup needed. All application tracking, pipeline, scan history, and profile data is stored in `career-ops.db`.

#### Step 5: Get to Know the User
Once the basics are in place, proactively learn about the user to personalize evaluations and CV generation:
> "Before we start evaluating offers, I'd like to understand you better. This helps me write stronger applications and spot the right opportunities:
>
> 1. **Superpowers:** What are your unique strengths? What do you do better than most people in your field?
> 2. **Preferences:** Remote, hybrid, or on-site? Company size preference (startup, mid-size, enterprise)? Any industry preferences?
> 3. **Deal-breakers:** Anything that's an automatic no? (e.g., mandatory relocation, specific technologies you refuse to work with, minimum comp, visa requirements)
> 4. **Key accomplishments:** What are 2-3 achievements you're most proud of? These become your go-to proof points in applications."

Store their answers in the DB via `career-ops profile set` or MCP `profile_update` / `profile_enrich` tools. Use this context in every evaluation and CV generation going forward.

#### Step 6: Ready
Once all files exist, confirm:
> "You're all set! You can now:
> - Paste a job URL to evaluate it
> - Run `/career-ops scan` to search portals
> - Run `/career-ops` to see all commands
>
> Everything is customizable — just ask me to change anything.
>
> Tip: Having a personal portfolio dramatically improves your job search. If you don't have one yet, the author's portfolio is also open source: github.com/santifer/cv-santiago — feel free to fork it and make it yours."

Then suggest automation:
> "Want me to scan for new offers automatically? I can set up a recurring scan every few days so you don't miss anything. Just say 'scan every 3 days' and I'll configure it."

If the user accepts, use the `/loop` or `/schedule` skill (if available) to set up a recurring `/career-ops scan`. If those aren't available, suggest adding a cron job or remind them to run `/career-ops scan` periodically.

### Personalization

This system is designed to be customized by YOU (Claude). When the user asks you to change archetypes, translate modes, adjust scoring, add companies, or modify negotiation scripts -- do it directly. You read the same files you use, so you know exactly what to edit.

**Common customization requests:**
- "Change the archetypes to [backend/frontend/data/devops] roles" → edit `modes/_shared.md`
- "Translate the modes to English" → edit all files in `modes/`
- "Add these companies to my portals" → edit `portals.yml`
- "Update my profile" → `career-ops profile set <field> <value>` or MCP `profile_update`
- "Change the CV template design" → edit `templates/cv-template.html`
- "Adjust the scoring weights" → edit `modes/_shared.md` and `batch/batch-prompt.md`

### Skill Modes

| If the user... | Mode |
|----------------|------|
| Pastes JD or URL | auto-pipeline (evaluate + report + PDF + tracker) |
| Asks to evaluate offer | `evaluate` |
| Asks to compare offers | `compare` |
| Wants LinkedIn outreach | `outreach` |
| Asks for company research | `deep` |
| Wants to generate CV/PDF | `pdf` |
| Evaluates a course/cert | `training` |
| Evaluates portfolio project | `project` |
| Asks about application status | `tracker` |
| Fills out application form | `apply` |
| Searches for new offers | `scan` |
| Processes pending URLs | `pipeline` |
| Batch processes offers | `batch` |
| Manages their profile | `career-ops profile` CLI |
| Wants profile enrichment | MCP `profile_enrich` tool |

### CV Source of Truth

- `cv.md` in project root is the canonical CV
- `article-digest.md` has detailed proof points (optional)
- **NEVER hardcode metrics** -- read them from these files at evaluation time

---

## Ethical Use -- CRITICAL

**This system is designed for quality, not quantity.** The goal is to help the user find and apply to roles where there is a genuine match -- not to spam companies with mass applications.

- **NEVER submit an application without the user reviewing it first.** Fill forms, draft answers, generate PDFs -- but always STOP before clicking Submit/Send/Apply. The user makes the final call.
- **Discourage low-fit applications.** If a score is below 4.0/5, explicitly tell the user this is a weak match and recommend skipping unless they have a specific reason.
- **Quality over speed.** A well-targeted application to 5 companies beats a generic blast to 50. Guide the user toward fewer, better applications.
- **Respect recruiters' time.** Every application a human reads costs someone's attention. Only send what's worth reading.

---

## Offer Verification -- MANDATORY

**NEVER trust WebSearch/WebFetch to verify if an offer is still active.** ALWAYS use chromedp:
1. `chromedp.Navigate` to the URL
2. `chromedp.OuterHTML` to read content
3. Only footer/navbar without JD = closed. Title + description + Apply = active.

---

## Stack and Conventions

- Go (cobra CLI, chromedp for browser automation, bubbletea for TUI, **SQLite via modernc.org/sqlite**), YAML (config), HTML/CSS (template)
- **SQLite** is the single source of truth — all data stored in `career-ops.db`
- **ksql** (vingarcia/ksql) for entity CRUD, raw SQL for FTS5/aggregation
- **goose** (pressly/goose/v3) for migrations in `internal/db/migrations/`
- **samber/lo** for all collection operations (Map, Filter, GroupBy, ForEach, etc.) — NEVER manual for loops
- **samber/oops** for all error wrapping — NEVER fmt.Errorf
- **samber/mo** for monads (Result, Option) in worker pool returns
- **mcp-go** (mark3labs/mcp-go v0.47) for MCP server
- CLI source in `cmd/career-ops/`, library code in `internal/`
- Output in `output/` (gitignored), Reports in `reports/`
- JDs in `jds/` (referenced as `local:jds/{file}`)
- Batch in `batch/` (gitignored except scripts and prompt)
- Report numbering: sequential 3-digit zero-padded, max existing + 1
- **RULE: After each batch of evaluations, run `career-ops merge`** to merge tracker additions and avoid duplications.
- **RULE: NEVER create duplicate entries in DB if company+role already exists.** Use `repo.UpsertApplication` to update.

### Go Conventions

- **Collections**: Use `samber/lo` for ALL collection operations. Never write manual for loops.
- **Errors**: Use `samber/oops` for ALL error wrapping. Never use `fmt.Errorf`.
- **Monads**: Use `samber/mo` (Result, Option) for worker pool returns.
- **Validation**: Manual validation with `validationError` collector in `internal/db/validation.go`.
- **Entity pattern**: Each entity (Application, PipelineEntry, ScanRecord, Evaluation, UserProfile, ProfileEnrichment, Submission) gets its own file in `internal/db/` with model struct, ksql table definition, and CRUD methods.
- **Migrations**: Each migration in its own `.sql` file under `internal/db/migrations/` with goose up/down annotations.
- **Lint**: Zero tolerance -- no `//nolint` directives, no `.golangci.yml` exclusion rules. Fix the actual code.
- **Task runner**: Always use `mise exec -- task` prefix (never bare `task`).

### TSV Format for Tracker Additions

Write one TSV file per evaluation to `batch/tracker-additions/{num}-{company-slug}.tsv`. Single line, 9 tab-separated columns:

```
{num}\t{date}\t{company}\t{role}\t{status}\t{score}/5\t{pdf_emoji}\t[{num}](reports/{num}-{slug}-{date}.md)\t{note}
```

**Column order (IMPORTANT -- status BEFORE score):**
1. `num` -- sequential number (integer)
2. `date` -- YYYY-MM-DD
3. `company` -- short company name
4. `role` -- job title
5. `status` -- canonical status (e.g., `Evaluated`)
6. `score` -- format `X.X/5` (e.g., `4.2/5`)
7. `pdf` -- `✅` or `❌`
8. `report` -- markdown link `[num](reports/...)`
9. `notes` -- one-line summary

**Note:** In applications.md, score comes BEFORE status. The merge script handles this column swap automatically.

### Pipeline Integrity

1. **New entries**: Write TSV in `batch/tracker-additions/` and `career-ops merge` upserts into DB.
2. **Status updates**: Use `repo.UpsertApplication` or MCP `update_status` tool.
3. All reports MUST include `**URL:**` in the header (between Score and PDF).
4. All statuses MUST be canonical (see `templates/states.yml`).
5. Health check: `career-ops verify`
6. Normalize statuses: `career-ops normalize`
7. Dedup: `career-ops dedup`
8. Import legacy markdown into SQLite: `career-ops import`
9. Export SQLite to markdown: `career-ops export`

### Batch Processing -- Headless Fallback

> When running in headless mode (`claude -p`), chromedp may not be available. Use WebFetch as a fallback for JD extraction, but flag the result as "unconfirmed" since WebFetch may not render JavaScript-heavy pages.

### MCP Tools

The MCP server exposes these tools for AI-driven operations:

| Tool | Purpose |
|------|---------|
| `list_applications` | List tracked applications with optional filters |
| `search_applications` | Free-text search across applications |
| `add_to_pipeline` | Add a URL to the processing pipeline |
| `update_status` | Update an application's status |
| `get_metrics` | Pipeline statistics and metrics |
| `evaluate_offer` | Trigger offer evaluation |
| `profile_get` | Read the user's career profile |
| `profile_update` | Update a single profile field |
| `profile_enrich` | Record an enrichment from external source |
| `profile_enrichments` | List pending enrichments |
| `profile_apply_enrichment` | Apply a pending enrichment |

### Profile Enrichment

The AI can continuously improve the user profile from multiple sources:
- **Conversations**: Extract preferences, skills, deal-breakers from chat
- **GitHub**: Analyze repos for tech stack, contributions, proof points
- **LinkedIn**: Extract experience, endorsements, connections
- **Blog/Articles**: Mine portfolio content for proof points
- **Job applications**: Learn from feedback and interview outcomes

Use MCP `profile_enrich` tool to record enrichments, then `profile_apply_enrichment` to apply them after review.

### Canonical States

**Source of truth:** `templates/states.yml`

| State | When to use |
|-------|-------------|
| `Evaluated` | Report completed, pending decision |
| `Applied` | Application sent |
| `Responded` | Company responded |
| `Interview` | In interview process |
| `Offer` | Offer received |
| `Rejected` | Rejected by company |
| `Discarded` | Discarded by candidate or offer closed |
| `SKIP` | Doesn't fit, don't apply |

**RULES:**
- No markdown bold (`**`) in status field
- No dates in status field (use the date column)
- No extra text (use the notes column)
