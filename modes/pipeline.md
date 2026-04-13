# Mode: pipeline — URL Inbox (Second Brain)

Processes accumulated job listing URLs from the database (pipeline table). The user adds URLs whenever they want and then runs `/career-ops pipeline` to process them all.

## Workflow

1. **List pending pipeline** via `career-ops` CLI or `repo.ListPipeline(ctx, "pending")`
2. **For each pending URL**:
   a. Calculate next sequential `REPORT_NUM` (read `reports/`, take highest number + 1)
   b. **Extract JD** using Playwright (browser_navigate + browser_snapshot) → WebFetch → WebSearch
   c. If the URL is not accessible → mark as `- [!]` with note and continue
   d. **Run full auto-pipeline**: Evaluation A-F → Report .md → PDF (if score >= 3.0) → Tracker
   e. **Mark as processed** in DB: update pipeline entry status to "processed" with result `#NNN | Company | Role | Score/5 | PDF ✅/❌`
3. **If there are 3+ pending URLs**, launch agents in parallel (Agent tool with `run_in_background`) to maximize speed.
4. **When finished**, display summary table:

```
| # | Company | Role | Score | PDF | Recommended Action |
```

## Smart JD Detection from URL

1. **Playwright (preferred):** `browser_navigate` + `browser_snapshot`. Works with all SPAs.
2. **WebFetch (fallback):** For static pages or when Playwright is not available.
3. **WebSearch (last resort):** Search secondary portals that index the JD.

**Special cases:**
- **LinkedIn**: May require login → mark `[!]` and ask the user to paste the text
- **PDF**: If the URL points to a PDF, read it directly with the Read tool
- **`local:` prefix**: Read the local file. Example: `local:jds/linkedin-pm-ai.md` → read `jds/linkedin-pm-ai.md`

## Automatic Numbering

1. List all files in `reports/`
2. Extract the number from the prefix (e.g., `142-medispend...` → 142)
3. New number = max found + 1

## Source Sync

Before processing any URL, verify sync:
```bash
career-ops sync-check
```
If there's a desynchronization, warn the user before continuing.

> **Note:** Use `career-ops export --format pipeline` to export pipeline to markdown if needed.
