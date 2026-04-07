# Career-Ops

> A Go port of [santifer/career-ops](https://github.com/santifer/career-ops) -- the AI-powered job search pipeline built on Claude Code. Full credit to [Santiago](https://santifer.io) for the original system, modes, scoring logic, and pipeline design. This fork replaces Node.js with a single Go binary.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Claude Code](https://img.shields.io/badge/Claude_Code-000?style=flat&logo=anthropic&logoColor=white)
![Bubble Tea](https://img.shields.io/badge/Bubble_Tea-FF75B5?style=flat&logo=go&logoColor=white)
![chromedp](https://img.shields.io/badge/chromedp-4285F4?style=flat&logo=googlechrome&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-003B57?style=flat&logo=sqlite&logoColor=white)
![golangci--lint](https://img.shields.io/badge/golangci--lint-00ACD7?style=flat&logo=go&logoColor=white)
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

---

## What Changed in This Fork

The original career-ops runs on Node.js with 6 separate `.mjs` scripts and Playwright for PDF generation. This fork consolidates everything into a single statically-compiled Go binary:

| Original (Node.js) | This Fork (Go) |
|---------------------|----------------|
| 6 `.mjs` scripts | Single `career-ops` binary |
| Playwright (headless browser) | chromedp (Chrome DevTools Protocol) |
| `npm install` + `npx playwright install` | `go install` or `task build` |
| Separate `dashboard/` directory | `career-ops dashboard` subcommand |
| npm scripts / package.json | go-task (Taskfile.yml) |
| Node.js runtime required | Zero runtime dependencies |
| Markdown tables for data | SQLite database (modernc.org/sqlite, pure Go) |
| No MCP support | MCP server (stdio transport) for AI agent integration |

Everything else -- modes, templates, scoring, portals, batch processing -- remains identical to upstream.

## Features

| Feature | Description |
|---------|-------------|
| **Auto-Pipeline** | Paste a URL, get a full evaluation + PDF + tracker entry |
| **6-Block Evaluation** | Role summary, CV match, level strategy, comp research, personalization, interview prep (STAR+R) |
| **Interview Story Bank** | Accumulates STAR+Reflection stories across evaluations |
| **Negotiation Scripts** | Salary negotiation frameworks, geographic discount pushback, competing offer leverage |
| **ATS PDF Generation** | Keyword-injected CVs via chromedp with Space Grotesk + DM Sans design |
| **Portal Scanner** | 45+ companies pre-configured across Ashby, Greenhouse, Lever, Wellfound |
| **Batch Processing** | Parallel evaluation with `claude -p` workers |
| **Dashboard TUI** | Bubble Tea terminal UI to browse, filter, and sort your pipeline |
| **Pipeline Integrity** | Automated merge, dedup, status normalization, health checks |
| **SQLite Storage** | Application data, pipeline, scan history stored in SQLite with FTS5 search |
| **MCP Server** | Model Context Protocol server for AI agent tool integration |
| **Import/Export** | Bidirectional migration between markdown and SQLite formats |

## Quick Start

### Prerequisites

- Go 1.24+ (or use [mise](https://mise.jdx.dev/) to manage versions)
- Google Chrome or Chromium (for PDF generation via chromedp)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) installed

### Install

```bash
# Option 1: go install
go install github.com/omarluq/career-ops/cmd/career-ops@latest

# Option 2: Clone and build
git clone https://github.com/omarluq/career-ops.git
cd career-ops
task build          # Binary lands in ./bin/career-ops

# Option 3: With mise (installs Go, task, linter, etc.)
git clone https://github.com/omarluq/career-ops.git
cd career-ops
mise install        # Installs all tool versions
task build
```

### Configure

```bash
cp config/profile.example.yml config/profile.yml  # Edit with your details
cp templates/portals.example.yml portals.yml       # Customize companies
```

Create `cv.md` in the project root with your CV in markdown, then open Claude Code:

```bash
claude   # Claude auto-detects career-ops and enters onboarding if needed
```

> **The system is designed to be customized by Claude itself.** Modes, archetypes, scoring weights, negotiation scripts -- just ask Claude to change them.

## CLI Commands

All pipeline tools are subcommands of the single binary:

```bash
career-ops verify       # Pipeline health check (statuses, dupes, links)
career-ops merge        # Merge batch TSV additions into applications.md
career-ops dedup        # Remove duplicate tracker entries
career-ops normalize    # Map status aliases to canonical statuses
career-ops sync-check   # Validate setup consistency
career-ops pdf <in> <out> [--format=letter|a4]  # Generate PDF from HTML
career-ops batch        # Run batch processing
career-ops dashboard    # Launch the TUI pipeline viewer
career-ops import       # Import markdown data into SQLite
career-ops export       # Export SQLite data to markdown
career-ops mcp          # Start MCP server (stdio transport)
```

Or use go-task for development shortcuts:

```bash
mise exec -- task build       # Build binary to ./bin/
mise exec -- task install     # Install to $GOPATH/bin
mise exec -- task test        # Run tests with race detector
mise exec -- task lint        # Run golangci-lint
mise exec -- task fmt         # Format and auto-fix
mise exec -- task ci          # Full CI pipeline (fmt + lint + test + build)
mise exec -- task dashboard   # Build and launch TUI
```

## Usage with Claude Code

Career-ops is driven by Claude Code with skill modes:

```
/career-ops                  Show all available commands
/career-ops {paste a JD}     Full auto-pipeline (evaluate + PDF + tracker)
/career-ops scan             Scan portals for new offers
/career-ops pdf              Generate ATS-optimized CV
/career-ops batch            Batch evaluate multiple offers
/career-ops tracker          View application status
/career-ops apply            Fill application forms with AI
/career-ops pipeline         Process pending URLs
/career-ops contacto         LinkedIn outreach message
/career-ops deep             Deep company research
/career-ops training         Evaluate a course/cert
/career-ops project          Evaluate a portfolio project
```

Or just paste a job URL or description directly -- career-ops auto-detects it and runs the full pipeline.

## How It Works

```
You paste a job URL or description
        |
        v
+------------------+
|  Archetype       |  Classifies role type against your target archetypes
|  Detection       |
+--------+---------+
         |
+--------v---------+
|  A-F Evaluation   |  Match, gaps, comp research, STAR stories
|  (reads cv.md)    |
+--------+---------+
         |
    +----+----+
    v    v    v
 Report  PDF  Tracker
  .md   .pdf   .tsv
```

## Pre-configured Portals

The scanner comes with **45+ companies** ready to scan and **19 search queries** across major job boards. Copy `templates/portals.example.yml` to `portals.yml` and add your own:

**AI Labs:** Anthropic, OpenAI, Mistral, Cohere, LangChain, Pinecone
**Voice AI:** ElevenLabs, PolyAI, Parloa, Hume AI, Deepgram, Vapi, Bland AI
**AI Platforms:** Retool, Airtable, Vercel, Temporal, Glean, Arize AI
**Contact Center:** Ada, LivePerson, Sierra, Decagon, Talkdesk, Genesys
**Enterprise:** Salesforce, Twilio, Gong, Dialpad
**LLMOps:** Langfuse, Weights & Biases, Lindy, Cognigy, Speechmatics
**Automation:** n8n, Zapier, Make.com
**European:** Factorial, Attio, Tinybird, Clarity AI, Travelperk

**Job boards searched:** Ashby, Greenhouse, Lever, Wellfound, Workable, RemoteFront

## Project Structure

```
career-ops/
├── cmd/career-ops/             # CLI entrypoint and subcommands
│   ├── main.go
│   ├── root.go                 # Cobra root command
│   ├── cmd_verify.go
│   ├── cmd_merge.go
│   ├── cmd_dedup.go
│   ├── cmd_normalize.go
│   ├── cmd_sync_check.go
│   ├── cmd_pdf.go
│   ├── cmd_batch.go
│   └── cmd_dashboard.go
├── internal/                   # Library packages
│   ├── closer/                 # Deferred close error handling
│   ├── db/                     # SQLite database layer
│   │   ├── application.go      # Application entity (model + CRUD)
│   │   ├── pipeline_entry.go   # Pipeline entity
│   │   ├── scan_record.go      # Scan history entity
│   │   ├── evaluation.go       # Evaluation entity
│   │   ├── validation.go       # Input validation
│   │   └── migrations/         # Goose SQL migrations
│   ├── mcp/                    # MCP server (tools + resources)
│   ├── model/                  # Shared data models
│   ├── repo/                   # Repository interface + SQLite impl
│   ├── scanner/                # Concurrent portal scanner
│   ├── states/                 # Canonical status management
│   ├── tracker/                # Markdown parsing, TSV, fuzzy matching
│   ├── ui/                     # Bubble Tea dashboard
│   │   ├── screens/            # Pipeline and viewer screens
│   │   └── theme/              # Catppuccin Mocha theme
│   ├── vinfo/                  # Version info (injected at build)
│   └── worker/                 # Generic worker pool + fan-out
├── modes/                      # 14 Claude skill modes
│   ├── _shared.md
│   ├── oferta.md
│   ├── pdf.md
│   ├── scan.md
│   ├── batch.md
│   └── ...
├── templates/
│   ├── cv-template.html        # ATS-optimized CV template
│   ├── portals.example.yml     # Scanner config template
│   └── states.yml              # Canonical statuses
├── config/
│   └── profile.example.yml     # Profile template
├── batch/
│   ├── batch-prompt.md         # Self-contained worker prompt
│   └── batch-runner.sh         # Orchestrator script
├── fonts/                      # Space Grotesk + DM Sans
├── data/                       # Tracking data (gitignored)
├── reports/                    # Evaluation reports (gitignored)
├── output/                     # Generated PDFs (gitignored)
├── Taskfile.yml                # go-task build automation
├── .mise.toml                  # Tool version management
├── .golangci.yml               # Linter configuration
├── go.mod
├── go.sum
└── CLAUDE.md                   # Agent instructions
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.24 |
| **CLI Framework** | Cobra |
| **Configuration** | Viper + YAML |
| **PDF Generation** | chromedp (Chrome DevTools Protocol) |
| **Dashboard TUI** | Bubble Tea + Lipgloss (Catppuccin Mocha) |
| **Build Automation** | go-task |
| **Tool Management** | mise |
| **Linting** | golangci-lint v2 |
| **Agent** | Claude Code with custom skill modes |
| **Data** | SQLite (modernc.org/sqlite, pure Go, CGO-free) + ksql + goose migrations |
| **Functional Go** | samber/lo (collections) + samber/oops (errors) + samber/mo (monads) |
| **MCP** | mark3labs/mcp-go v0.47 (stdio transport) |

## Ethical Use

**This system is designed for quality, not quantity.** The goal is to help you find and apply to roles where there is a genuine match -- not to spam companies with mass applications.

- Never submit an application without reviewing it first
- If a score is below 3.0/5, the system explicitly recommends skipping
- A well-targeted application to 5 companies beats a generic blast to 50
- Every application a human reads costs someone's attention -- only send what is worth reading

## Upstream

This is a Go port of [santifer/career-ops](https://github.com/santifer/career-ops). The original system was built by [Santiago](https://santifer.io), who used it to evaluate 740+ offers, generate 100+ tailored CVs, and land a Head of Applied AI role. The portfolio that pairs with this system is also open source: [cv-santiago](https://github.com/santifer/cv-santiago).

## License

MIT
