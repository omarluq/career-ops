# Setup Guide

## Prerequisites

- [Claude Code](https://claude.ai/code) installed and configured
- Go 1.24+ (or [mise](https://mise.jdx.dev/) to manage it automatically)
- Chrome or Chromium (for PDF generation via chromedp)

## Quick Start

### 1. Clone and build

```bash
git clone https://github.com/omarluq/career-ops.git
cd career-ops
mise install           # Or manually install Go 1.24+
mise exec -- task build
```

This compiles the `career-ops` binary with version injection and places it in `$GOPATH/bin`.

### 2. Configure your profile

```bash
cp config/profile.example.yml config/profile.yml
```

Edit `config/profile.yml` with your personal details: name, email, target roles, narrative, proof points.

### 3. Add your CV

Create `cv.md` in the project root with your full CV in markdown format. This is the source of truth for all evaluations and PDFs.

(Optional) Create `article-digest.md` with proof points from your portfolio projects/articles.

### 4. Configure portals

```bash
cp templates/portals.example.yml portals.yml
```

Edit `portals.yml`:
- Update `title_filter.positive` with keywords matching your target roles
- Add companies you want to track in `tracked_companies`
- Customize `search_queries` for your preferred job boards

### 5. Import existing data (optional)

If you are migrating from a previous markdown-based tracker:

```bash
career-ops import
```

This imports `data/applications.md` into the SQLite database.

### 6. Start using

Open Claude Code in this directory:

```bash
claude
```

Then paste a job offer URL or description. Career-ops will automatically evaluate it, generate a report, create a tailored PDF, and track it.

## Available Commands

| Action | How |
|--------|-----|
| Evaluate an offer | Paste a URL or JD text |
| Search for offers | `/career-ops scan` |
| Process pending URLs | `/career-ops pipeline` |
| Generate a PDF | `career-ops pdf <input.html> <output.pdf>` |
| Batch evaluate | `/career-ops batch` |
| Check tracker status | `/career-ops tracker` |
| Fill application form | `/career-ops apply` |
| Import markdown to SQLite | `career-ops import` |
| Export SQLite to markdown | `career-ops export` |
| Start MCP server | `career-ops mcp` |
| Open dashboard TUI | `career-ops dashboard` |

## Verify Setup

```bash
career-ops verify       # Check pipeline integrity
career-ops sync-check   # Validate configuration consistency
```
