# Mode: tracker — Application Tracker

Reads and displays applications from the SQLite database.

**Tracker format:**
```markdown
| # | Date | Company | Role | Score | Status | PDF | Report |
```

Possible statuses: `Evaluated` → `Applied` → `Responded` → `Contact` → `Interview` → `Offer` / `Rejected` / `Discarded` / `SKIP`

- `Applied` = the candidate submitted their application
- `Responded` = A recruiter/company reached out and the candidate replied (inbound)
- `Contact` = The candidate proactively reached out to someone at the company (outbound, e.g., LinkedIn power move)

If the user asks to update a status, use `career-ops` CLI or MCP tools to update statuses.

> **Note:** Use `career-ops export --format applications` to export to markdown if needed.

Also show statistics:
- Total applications
- By status
- Average score
- % with PDF generated
- % with report generated
