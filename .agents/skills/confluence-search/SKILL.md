---
name: confluence-search
description: Search and fetch Confluence wiki pages via the confluence-search CLI. Use when the user asks about internal docs, wiki pages, runbooks, ADRs, RFCs, onboarding guides, or any Confluence-hosted knowledge.
argument-hint: "<query>"
user-invocable: true
allowed-tools: "Bash, Read"
---

# Confluence Search

Run `confluence-search --help` and `confluence-search search --help` to get
the full flag reference. The CLI is self-documenting.

## Setup

Requires `CONFLUENCE_URL` and `CONFLUENCE_PERSONAL_ACCESS_TOKEN` env vars.
Verify: `confluence-search health`

## Examples

```bash
# Search pages
confluence-search search "deployment process"
confluence-search search "API gateway" --spaces ENG,OPS --limit 10
confluence-search search "runbook" --labels oncall --titles-only
confluence-search search "incident" --modified-after 30d

# Get content IDs from JSON output
confluence-search search "deployment" --json

# Fetch full page by content ID
confluence-search fetch 98765

# Preview CQL without calling API
confluence-search search "migration guide" --dry-run
```
