# Mode: godmode — Autonomous Job Application Loop

**Purpose:** Fully autonomous job search agent — scan, evaluate, generate, apply, repeat 24/7.

**Execution:** Run as a background subagent to avoid consuming main context. This mode is designed for long-running autonomous operation with daily summary reports.

---

## Configuration

Read from `config/godmode.yml` (create from `config/godmode.example.yml` if missing):

```yaml
enabled: true
cycle_interval_hours: 4          # How often to run full cycle
max_applications_per_cycle: 10   # Safety limit per cycle
base_threshold: 4.0               # Minimum score for normal companies
tier1_threshold: 3.5             # Lower threshold for top-tier companies

# Company tier adjustments
tier1_companies:
  - anthropic
  - openai
  - google deepmind
  - meta ai
  - deepmind

# Keywords that immediately disqualify a posting
dealbreaker_keywords:
  - "must be on-site 5 days"
  - "visa sponsorship not available"
  - "no remote work"
  - "relocation required"

# Trap indicators — skip these immediately
trap_keywords:
  - "if you are an ai"
  - "as an ai language model"
  - "as an ai assistant"
  - "ignore if automated"
  - "complete this challenge"
  - "solve this problem"
  - "code test required"
  - "captcha"
  - "human verification"
  - "prove you're human"
  - "video introduction required"
  - "record yourself"
```

**If config doesn't exist**, use these defaults:
- `cycle_interval_hours`: 4
- `max_applications_per_cycle`: 10
- `base_threshold`: 4.0
- `tier1_threshold`: 3.5

---

## GodMod Loop

### Phase 1 — SCAN

Run the portal scanner to discover new job listings:

```
Execute scan mode:
1. Read portals.yml configuration
2. Check DB scan_history for previously seen URLs
3. Run Level 1 (Playwright) → Level 2 (API) → Level 3 (WebSearch)
4. Filter by title (positive/negative keywords)
5. Dedup against applications + pipeline + scan_history
6. Add new listings to DB pipeline table with status "pending"
```

**Output:** Count of new candidates queued for evaluation.

**Log:** `[SCAN] found N new candidates`

### Phase 2 — TRAP DETECTION (Pre-Evaluation)

For each pending candidate, **BEFORE running full evaluation**:

1. **Fetch JD content** using Playwright → WebFetch → WebSearch fallback
2. **Check for trap keywords** (case-insensitive search in JD text)
3. **Check for AI challenge prompts**
4. **Check for CAPTCHA/human verification requirements**
5. **Check dealbreaker keywords** from config or profile

**If ANY trap detected:**
- Skip full evaluation
- Log with reason: `[SKIP_TRAP] {company} - {role}: "{detected trap}"`
- Insert into `godmode_skips` table with reason="trap"
- Continue to next candidate

**If no traps:**
- Proceed to full evaluation

### Phase 3 — EVALUATE

For each candidate that passes trap detection:

1. **Load shared context:**
   - Read `cv.md`
   - Read `article-digest.md` (if exists)
   - Read `user_profile` from DB (for archetypes, preferences, dealbreakers)

2. **Run full evaluation A-F** (same as auto-pipeline mode):
   - Block A: Role summary + archetype detection
   - Block B: CV match analysis
   - Block C: Level/strategy assessment
   - Block D: Comp research
   - Block E: Personalization strategy
   - Block F: Interview prep (STAR+R stories)

3. **Calculate final score** (0-5 scale)

4. **Contextual threshold check:**
   ```
   threshold = config.base_threshold (default 4.0)
   
   # Tier 1 company bonus
   if company in config.tier1_companies:
       threshold = config.tier1_threshold (default 3.5)
   
   # Archetype match bonus
   if detected_archetype in user_profile.archetypes:
       threshold -= 0.2
   
   # Remote match bonus
   if user_profile.remote AND job_is_remote:
       threshold -= 0.1
   
   # Dealbreaker check (hard skip)
   for dealbreaker in user_profile.dealbreakers + config.dealbreaker_keywords:
       if dealbreaker in JD:
           threshold = INFINITY (skip)
   ```

5. **Decision:**
   - If `score >= threshold`: **PASS** → proceed to generate phase
   - Else: **SKIP_THRESHOLD**
   - Log: `[EVALUATE] {company} - {role}: {score}/5 - {PASS|SKIP_THRESHOLD}`
   - Insert into `godmode_skips` table if skipped

### Phase 4 — GENERATE

For each candidate that passed threshold:

1. **Generate ATS PDF:**
   - Read `templates/cv-template.html`
   - Inject JD keywords into CV content
   - Run `career-ops pdf` (chromedp → output/{num}-{slug}.pdf)

2. **Generate cover letter** (if form supports it):
   - Map JD quotes to proof points
   - 1 page max, same visual design as CV
   - Save to `output/{num}-{slug}-cover.pdf`

3. **Prepare SubmissionJob:**
   ```go
   {
       Company: "{company}",
       Role: "{role}",
       JDURL: "{original_url}",
       CVPDF: "{pdf_path}",
       CoverLetter: "{cover_letter_text}",
       Portal: {detected ATS type},
       FormData: {
           // Pre-filled from profile
           "full_name": user_profile.name,
           "email": user_profile.email,
           "location": user_profile.location,
           // ... more fields as needed
       }
   }
   ```

**Log:** `[GENERATE] {company} - {role}: PDF ready`

### Phase 5 — APPLY

1. **Batch prepare** all SubmissionJobs from generate phase

2. **Safety check:**
   - If count > `config.max_applications_per_cycle`: truncate and log warning

3. **Run Applicator.SubmitBatch:**
   ```go
   applicator := New(repo, concurrency=3, rateLimit=30*time.Second)
   results, err := applicator.SubmitBatch(ctx, jobs, func(job, result) {
       log("[APPLY] %s - %s: %s", job.Company, job.Role, result.Status)
   })
   ```

4. **Handle results:**
   - `submitted`: Update DB applications table with status="Applied", link to PDF
   - `failed`: Log error, insert into `godmode_skips` with reason="apply_failed"
   - `skipped`: Log reason, insert into `godmode_skips`

5. **Post-apply tracking:**
   - Record submission in DB `submissions` table
   - Update pipeline entry status to "processed"

**Log:** `[APPLY] {company} - {role}: {submitted|failed|skipped}`

### Phase 6 — DAILY REPORT

Generate summary report of the past 24 hours (or since last report):

1. **Query logs** from `logs/godmode-{date}.log`

2. **Compile metrics:**
   - Jobs scanned
   - New candidates
   - Evaluated (avg score)
   - Passed threshold
   - Applied successfully
   - Skipped (traps)
   - Skipped (threshold)
   - Failed (with error summary)

3. **Generate report** at `reports/godmode-summary-{YYYY-MM-DD}.md`:
   ```markdown
   # GodMod Daily Report — {YYYY-MM-DD}
   
   ## Summary
   - Jobs scanned: {N}
   - New candidates: {N}
   - Evaluated: {N} (avg score: X.X/5)
   - Passed threshold: {N}
   - Applied: {N}
   - Skipped (traps): {N}
   - Skipped (threshold): {N}
   - Failed: {N}
   
   ## Applied Successfully
   | Company | Role | Score | Archetype | Status |
   |---------|------|-------|-----------|--------|
   | ... | ... | ... | ... | submitted |
   
   ## Skipped (Traps)
   | Company | Reason | JD URL |
   |---------|--------|--------|
   | ... | "AI language model" prompt | ... |
   
   ## Skipped (Threshold)
   | Company | Role | Score | Threshold |
   |---------|------|-------|-----------|
   | ... | ... | 3.8 | 4.0 |
   
   ## Errors
   | Company | Error |
   |---------|-------|
   | ... | timeout during form fill |
   ```

4. **Log:** `[DAILY_REPORT] generated: reports/godmode-summary-{date}.md`

### Phase 7 — SLEEP

Wait `config.cycle_interval_hours` before next cycle.

**Log:** `[SLEEP] {N} hours until next cycle (next run at {timestamp})`

**Resume capability:** If interrupted, godmode resumes from last completed phase on next run.

---

## Log Format

All godmode operations log to `logs/godmode-{YYYY-MM-DD}.log`:

```
[2026-04-13 14:23:01] GODMOD started (cycle {N})
[2026-04-13 14:23:02] CONFIG: base_threshold=4.0, tier1_threshold=3.5, max_per_cycle=10
[2026-04-13 14:23:15] SCAN: started
[2026-04-13 14:24:30] SCAN: found 12 new candidates
[2026-04-13 14:24:31] TRAP_CHECK: checking 12 candidates
[2026-04-13 14:24:35] TRAP_CHECK: TrapCorp - AI Eng SKIP ("if you are an ai")
[2026-04-13 14:24:40] EVALUATE: Anthropic - AI Engineer: started
[2026-04-13 14:25:10] EVALUATE: Anthropic - AI Engineer: 4.5/5 PASS
[2026-04-13 14:25:11] GENERATE: Anthropic - AI Engineer: PDF generated
[2026-04-13 14:25:45] APPLY: Anthropic - AI Engineer: submitted
[2026-04-13 14:26:00] EVALUATE: StartupXYZ - AI Eng: 3.2/5 SKIP_THRESHOLD
[2026-04-13 14:26:30] EVALUATE: OpenAI - Research Eng: 4.8/5 PASS
[2026-04-13 14:27:00] GENERATE: OpenAI - Research Eng: PDF generated
[2026-04-13 14:27:30] APPLY: OpenAI - Research Eng: failed (timeout)
[2026-04-13 18:00:00] DAILY_REPORT: generated
[2026-04-13 18:00:01] SLEEP: 4 hours (next: 2026-04-13 22:00:01)
[2026-04-13 18:00:02] GODMOD cycle {N} complete
```

---

## CLI Usage

```
/career-ops godmode [--once] [--report-only]
```

**Flags:**
- `--once`: Run one cycle and exit (for testing/debugging)
- `--report-only`: Generate daily report without running cycle

**Examples:**
```
/career-ops godmode              # Continuous loop (24/7 mode)
/career-ops godmode --once       # Single run for testing
/career-ops godmode --report-only # Just today's report
```

---

## Database Tables (New)

**Table: `godmode_runs`**
- id (PK)
- started_at
- completed_at
- cycle_number
- phase (scan/trap_check/evaluate/generate/apply/report/sleep)
- candidates_found
- candidates_evaluated
- candidates_passed
- applications_submitted
- applications_failed

**Table: `godmode_skips`**
- id (PK)
- company
- role
- url
- reason (trap/threshold/dealbreaker/apply_failed)
- details (text)
- skipped_at

---

## Safety Features

1. **Rate limiting:** Per-portal delays via Applicator (default 30s between same portal)
2. **Max applications per cycle:** Configurable safety limit (default 10)
3. **Trap detection:** Pre-evaluation filter saves tokens and avoids spam
4. **Dealbreaker respect:** Honors user preferences from profile
5. **Resume capability:** Interrupted runs resume from last safe state
6. **Daily reports:** Full visibility into what godmode did
7. **No auto-submit without review:** All applications tracked, user can review any time

---

## Integration with Existing Components

| Phase | Uses |
|-------|-------|
| Scan | `modes/scan.md` + `internal/scanner/` |
| Trap Detection | (new) |
| Evaluate | `modes/auto-pipeline.md` + `modes/evaluate.md` |
| Generate | `modes/pdf.md` + chromedp |
| Apply | `internal/applicator/` |
| Track | MCP tools + SQLite DB |

GodMod is an **orchestrator** — it coordinates existing modes and components, not a rewrite.

---

## User Control

At any time, user can:
- Check status: `/career-ops tracker` (see all godmode applications)
- Stop godmode: Ctrl+C (graceful shutdown after current phase)
- Review reports: `reports/godmode-summary-{date}.md`
- Check skips: Query `godmode_skips` table via MCP `search_applications`
- Adjust config: Edit `config/godmode.yml` (picked up next cycle)
- Pause: Set `enabled: false` in config

---

## Exit Conditions

GodMod stops if:
- User interrupts (Ctrl+C)
- Config `enabled: false`
- Critical error (DB connection loss, Chrome crash)
- Daily report generation fails (logs error, exits gracefully)

On shutdown, godmode:
1. Logs shutdown reason
2. Completes current phase if possible
3. Generates partial daily report if interrupted mid-cycle
