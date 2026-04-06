# Shared Context -- career-ops

<!-- ============================================================
     HOW TO CUSTOMIZE THIS FILE
     ============================================================
     This file contains the shared context for all career-ops modes.
     Before using career-ops, you MUST:
     1. Fill in config/profile.yml with your personal data
     2. Create your cv.md in the project root
     3. (Optional) Create article-digest.md with your proof points
     4. Customize the sections below marked with [CUSTOMIZE]
     ============================================================ -->

## Sources of Truth (ALWAYS read before evaluating)

| File | Path | When |
|------|------|------|
| cv.md | `cv.md` (project root) | ALWAYS |
| article-digest.md | `article-digest.md` (if exists) | ALWAYS (detailed proof points) |
| profile.yml | `config/profile.yml` | ALWAYS (candidate identity and targets) |

**RULE: NEVER hardcode metrics from proof points.** Read them from cv.md + article-digest.md at evaluation time.
**RULE: For article/project metrics, article-digest.md takes precedence over cv.md** (cv.md may have older numbers).

---

## North Star -- Target Roles

The skill applies with EQUAL rigor to ALL target roles. None is primary or secondary -- any is a success if comp and growth are right:

| Archetype | Thematic axes | What they buy |
|-----------|---------------|---------------|
| **Platform / DevOps Engineer** | Kubernetes, GitOps, IaC, CI/CD, cloud cost optimization, observability | Someone who builds and scales production infrastructure with measurable impact |
| **AI / ML Infrastructure Engineer** | MLOps, LLMOps, agentic systems, RAG pipelines, MCP servers, AI observability | Someone who builds the platform layer that makes AI work in production |
| **Fullstack Engineer** | Ruby on Rails, Go, React, Hotwire, GraphQL, edge computing, performance | Someone who ships end-to-end features with strong infra instincts |
| **Solutions Architect** | System design, integrations, enterprise-ready, cross-functional | Someone who designs scalable architectures and bridges teams |
| **SRE / Cloud Engineer** | Reliability, monitoring, incident response, automation, cost optimization | Someone who keeps systems running and makes them cheaper |

### Adaptive Framing by Archetype

> **Concrete metrics: read from `cv.md` + `article-digest.md` at evaluation time. NEVER hardcode numbers here.**

| If the role is... | Emphasize about the candidate... | Proof point sources |
|-------------------|----------------------------------|---------------------|
| Platform / DevOps | ECS→EKS migrations, ArgoCD/GitOps, 45% cost reduction, 2B request eliminations via Cloudflare Workers | config/profile.yml |
| AI / ML Infrastructure | Kagent (K8s-native AI agents), AI-enhanced GitOps with Kubechecks, CC-Relay (multi-provider LLM gateway), MCP servers | config/profile.yml |
| Fullstack Engineer | G2 Buyer Intent Data product, Ruby on Rails + Go + React + Hotwire, GraphQL APIs | config/profile.yml |
| Solutions Architect | Crossplane IaC (70% faster provisioning), enterprise AI adoption at scale | config/profile.yml |
| SRE / Cloud Engineer | ARM/Spot optimization, observability, incident response, cost optimization | config/profile.yml |

### Exit Narrative (use in ALL framings)

Read from `config/profile.yml` → `narrative.exit_story`:
"Deep infrastructure engineer who has driven enterprise AI adoption from the platform layer up. Track record of measurable impact: 45% cloud cost reductions, 60% faster deployments, 2B monthly request eliminations. Active open source contributor across Go, Crystal, and JS ecosystems."

Use the candidate's exit story from `config/profile.yml` to frame ALL content:
- **In PDF Summaries:** Bridge from past to future -- "Now applying the same [skill] to [JD domain]."
- **In STAR stories:** Reference proof points from article-digest.md
- **In Draft Answers (Section G):** The transition narrative should appear in the first response.
- **When the JD asks for "entrepreneurial", "ownership", "builder", "end-to-end":** This is the #1 differentiator. Increase match weight.

### Cross-cutting Advantage

Frame profile as **"Infrastructure engineer who bridges DevOps and AI"** that adapts framing to the role:
- For Platform/DevOps: "engineer who migrates, optimizes, and scales K8s infrastructure with measurable cost/speed impact"
- For AI/ML Infra: "engineer who builds the platform layer that makes AI work — from Kagent agents to MCP servers to RAG pipelines"
- For Fullstack: "engineer with deep infra instincts who ships end-to-end — Rails, Go, React, GraphQL, edge workers"
- For SA: "engineer who designs scalable architectures with real Crossplane/IaC/GitOps implementation experience"
- For SRE: "engineer who optimizes cost (45% reduction), speed (60% faster deploys), and reliability at enterprise scale"

Convert "builder" into a professional signal, not a "hobby maker". Real proof points make this credible.

### Portfolio as Proof Point (use in high-value applications)

<!-- [CUSTOMIZE] If you have a live demo, dashboard, or public project, configure it here.
     Example:
     dashboard:
       url: "https://yoursite.dev/demo"
       password: "demo-2026"
       when_to_share: "LLMOps, AI Platform, observability roles"
     Read from config/profile.yml → narrative.proof_points and narrative.dashboard -->

If the candidate has a live demo/dashboard (check profile.yml), offer access in applications for relevant roles.

### Comp Intelligence

<!-- [CUSTOMIZE] Research comp ranges for YOUR target roles and update these ranges -->

**General guidance:**
- Use WebSearch for current market data (Glassdoor, Levels.fyi, Blind)
- Frame by role title, not by skills -- titles determine comp bands
- Contractor rates are typically 30-50% higher than employee base to account for benefits
- Geographic arbitrage works for remote roles: lower CoL = better net

### Negotiation Scripts

<!-- [CUSTOMIZE] Adapt these to your situation -->

**Salary expectations (general framework):**
> "Based on market data for Senior/Staff Platform and DevOps roles, I'm targeting $180K-$220K base. I'm flexible on structure — what matters is the total package and the opportunity."

**Geographic discount pushback:**
> "The roles I'm competitive for are output-based, not location-based. My track record doesn't change based on postal code."

**When offered below target:**
> "I'm comparing with opportunities in the [higher range]. I'm drawn to [company] because of [reason]. Can we explore [target]?"

### Location Policy

<!-- [CUSTOMIZE] Adapt to your situation. Read from config/profile.yml → location -->

**In forms:**
- Binary "can you be on-site?" questions: follow your actual availability from profile.yml
- In free-text fields: specify your timezone overlap and availability

**In evaluations (scoring):**
- Remote dimension for hybrid outside your country: score **3.0** (not 1.0)
- Only score 1.0 if JD explicitly says "must be on-site 4-5 days/week, no exceptions"

### Time-to-offer priority
- Working demo + metrics > perfection
- Apply sooner > learn more
- 80/20 approach, timebox everything

---

## Global Rules

### NEVER

1. Invent experience or metrics
2. Modify cv.md or portfolio files
3. Submit applications on behalf of the candidate
4. Share phone number in generated messages
5. Recommend comp below market rate
6. Generate a PDF without reading the JD first
7. Use corporate-speak
8. Ignore the tracker (every evaluated offer gets registered)

### ALWAYS

0. **Cover letter:** If the form has an option to attach or write a cover letter, ALWAYS include one. Generate PDF with the same visual design as the CV. Content: JD quotes mapped to proof points, links to relevant case studies. 1 page max.
1. Read cv.md and article-digest.md (if exists) before evaluating any offer
1b. **First evaluation of each session:** Run `node cv-sync-check.mjs` with Bash. If it reports warnings, notify the candidate before continuing
2. Detect the role archetype and adapt framing
3. Cite exact lines from CV when matching
4. Use WebSearch for comp and company data
5. Register in tracker after evaluating
6. Generate content in the language of the JD (EN default)
7. Be direct and actionable -- no fluff
8. When generating English text (PDF summaries, bullets, LinkedIn messages, STAR stories): native tech English, not translated. Short sentences, action verbs, no unnecessary passive voice.
8b. **Case study URLs in PDF Professional Summary:** If the PDF mentions case studies or demos, URLs MUST appear in the first paragraph (Professional Summary). The recruiter may only read the summary. All URLs with `white-space: nowrap` in HTML.
9. **Tracker additions as TSV** -- NEVER edit applications.md to add new entries. Write TSV in `batch/tracker-additions/` and `merge-tracker.mjs` handles the merge.
10. **Include `**URL:**` in every report header** -- between Score and PDF.

### Tools

| Tool | Use |
|------|-----|
| WebSearch | Comp research, trends, company culture, LinkedIn contacts, fallback for JDs |
| WebFetch | Fallback for extracting JDs from static pages |
| Playwright | Verify if offers are still active (browser_navigate + browser_snapshot), extract JDs from SPAs. **CRITICAL: NEVER launch 2+ agents with Playwright in parallel -- they share a single browser instance.** |
| Read | cv.md, article-digest.md, cv-template.html |
| Write | Temporary HTML for PDF, applications.md, reports .md |
| Edit | Update tracker |
| Bash | `node generate-pdf.mjs` |
