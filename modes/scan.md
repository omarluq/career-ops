# Mode: scan — Portal Scanner (Job Discovery)

Scans configured job portals, filters by title relevance, and adds new listings to the pipeline for later evaluation.

## Recommended Execution

Run as a subagent to avoid consuming main context:

```
Agent(
    subagent_type="general-purpose",
    prompt="[contents of this file + specific data]",
    run_in_background=True
)
```

## Configuration

Read `portals.yml` which contains:
- `search_queries`: List of WebSearch queries with `site:` filters per portal (broad discovery)
- `tracked_companies`: Specific companies with `careers_url` for direct navigation
- `title_filter`: Positive/negative/seniority_boost keywords for title filtering

## Discovery Strategy (3 Levels)

### Level 1 — Direct Playwright (PRIMARY)

**For each company in `tracked_companies`:** Navigate to its `careers_url` with Playwright (`browser_navigate` + `browser_snapshot`), read ALL visible job listings, and extract title + URL for each one. This is the most reliable method because:
- Sees the page in real time (not cached Google results)
- Works with SPAs (Ashby, Lever, Workday)
- Detects new listings instantly
- Doesn't depend on Google indexing

**Every company MUST have a `careers_url` in portals.yml.** If missing, look it up once, save it, and use in future scans.

### Level 2 — Greenhouse API (COMPLEMENTARY)

For companies using Greenhouse, the JSON API (`boards-api.greenhouse.io/v1/boards/{slug}/jobs`) returns clean structured data. Use as a quick complement to Level 1 — faster than Playwright but only works with Greenhouse.

### Level 3 — WebSearch Queries (BROAD DISCOVERY)

The `search_queries` with `site:` filters cover portals broadly (all Ashby listings, all Greenhouse listings, etc.). Useful for discovering NEW companies not yet in `tracked_companies`, but results may be stale.

**Execution priority:**
1. Level 1: Playwright → all `tracked_companies` with `careers_url`
2. Level 2: API → all `tracked_companies` with `api:`
3. Level 3: WebSearch → all `search_queries` with `enabled: true`

Levels are additive — all run, results are merged and deduplicated.

## Workflow

1. **Read configuration**: `portals.yml`
2. **Read history** from DB (table `scan_history`) → previously seen URLs
3. **Read dedup sources** from DB: `repo.HasBeenScanned()` + `repo.ListApplications()`

4. **Level 1 — Playwright scan** (parallel in batches of 3-5):
   For each company in `tracked_companies` with `enabled: true` and `careers_url` defined:
   a. `browser_navigate` to the `careers_url`
   b. `browser_snapshot` to read all job listings
   c. If the page has filters/departments, navigate relevant sections
   d. For each job listing extract: `{title, url, company}`
   e. If the page paginates results, navigate additional pages
   f. Accumulate in candidate list
   g. If `careers_url` fails (404, redirect), try `scan_query` as fallback and note for URL update

5. **Level 2 — Greenhouse APIs** (parallel):
   For each company in `tracked_companies` with `api:` defined and `enabled: true`:
   a. WebFetch the API URL → JSON with job list
   b. For each job extract: `{title, url, company}`
   c. Accumulate in candidate list (dedup with Level 1)

6. **Level 3 — WebSearch queries** (parallel if possible):
   For each query in `search_queries` with `enabled: true`:
   a. Execute WebSearch with the defined `query`
   b. From each result extract: `{title, url, company}`
      - **title**: from the result title (before " @ " or " | ")
      - **url**: result URL
      - **company**: after " @ " in the title, or extract from domain/path
   c. Accumulate in candidate list (dedup with Level 1+2)

6. **Filter by title** using `title_filter` from `portals.yml`:
   - At least 1 keyword from `positive` must appear in the title (case-insensitive)
   - 0 keywords from `negative` must appear
   - `seniority_boost` keywords give priority but are not required

7. **Deduplicate** against DB:
   - table `scan_history` → exact URL already seen
   - table `applications` → normalized company + role already evaluated
   - table `pipeline` → exact URL already pending or processed

8. **For each new listing that passes filters**:
   a. Insert into DB table `pipeline` with status "pending": `{url, company, title}`
   b. Record in DB table `scan_history`: `{url, date, query_name, title, company, status: "added"}`

> **Note:** In addition to portals.yml, the system supports adapters for LinkedIn, YCombinator, Wellfound, and AI job boards via `career-ops scan --boards`

9. **Listings filtered by title**: record in `scan-history.tsv` with status `skipped_title`
10. **Duplicate listings**: record with status `skipped_dup`

## Title and Company Extraction from WebSearch Results

WebSearch results come in the format: `"Job Title @ Company"` or `"Job Title | Company"` or `"Job Title — Company"`.

Extraction patterns by portal:
- **Ashby**: `"Senior AI PM (Remote) @ EverAI"` → title: `Senior AI PM`, company: `EverAI`
- **Greenhouse**: `"AI Engineer at Anthropic"` → title: `AI Engineer`, company: `Anthropic`
- **Lever**: `"Product Manager - AI @ Temporal"` → title: `Product Manager - AI`, company: `Temporal`

Generic regex: `(.+?)(?:\s*[@|—–-]\s*|\s+at\s+)(.+?)$`

## Private URLs

If a non-publicly-accessible URL is found:
1. Save the JD to `jds/{company}-{role-slug}.md`
2. Add to pipeline.md as: `- [ ] local:jds/{company}-{role-slug}.md | {company} | {title}`

## Scan History

The `scan_history` table in the database tracks ALL seen URLs. Query via `repo.HasBeenScanned()` or `repo.ListScanHistory()`.

## Output Summary

```
Portal Scan — {YYYY-MM-DD}
━━━━━━━━━━━━━━━━━━━━━━━━━━
Queries executed: N
Listings found: N total
Filtered by title: N relevant
Duplicates: N (already evaluated or in pipeline)
New added to pipeline.md: N

  + {company} | {title} | {query_name}
  ...

→ Run /career-ops pipeline to evaluate the new listings.
```

## Managing careers_url

Every company in `tracked_companies` should have a `careers_url` — the direct URL to their jobs page. This avoids looking it up every time.

**Known patterns by platform:**
- **Ashby:** `https://jobs.ashbyhq.com/{slug}`
- **Greenhouse:** `https://job-boards.greenhouse.io/{slug}` or `https://job-boards.eu.greenhouse.io/{slug}`
- **Lever:** `https://jobs.lever.co/{slug}`
- **Custom:** The company's own URL (e.g., `https://openai.com/careers`)

**If `careers_url` doesn't exist** for a company:
1. Try the pattern for its known platform
2. If that fails, run a quick WebSearch: `"{company}" careers jobs`
3. Navigate with Playwright to confirm it works
4. **Save the found URL in portals.yml** for future scans

**If `careers_url` returns 404 or redirect:**
1. Note in the output summary
2. Try scan_query as fallback
3. Flag for manual update

## Maintaining portals.yml

- **ALWAYS save `careers_url`** when adding a new company
- Add new queries as interesting portals or roles are discovered
- Disable queries with `enabled: false` if they generate too much noise
- Adjust filtering keywords as target roles evolve
- Add companies to `tracked_companies` when you want to follow them closely
- Verify `careers_url` periodically — companies change ATS platforms
