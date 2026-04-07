---
name: confluence-search
description: Search and fetch Confluence wiki pages via the confluence-search CLI. Use when the user asks about internal docs, wiki pages, runbooks, ADRs, RFCs, onboarding guides, or any Confluence-hosted knowledge.
argument-hint: "<query>"
user-invocable: true
---

# Confluence Search

**First: run `confluence-search search --help` and `confluence-search fetch --help`
to see all current flags.** The CLI evolves; `--help` is the source of truth.

## Setup

Requires `CONFLUENCE_URL` and `CONFLUENCE_PERSONAL_ACCESS_TOKEN` env vars.
Verify connectivity: `confluence-search health`

## Search

```bash
# Basic search — returns title, URL, space, excerpt
confluence-search search "deployment process"

# Narrow by space, labels, recency
confluence-search search "API gateway" --spaces ENG,OPS --limit 10
confluence-search search "runbook" --labels oncall --titles-only
confluence-search search "incident" --modified-after 30d

# JSON output — use this to get content IDs for fetch
confluence-search search "deployment" --json

# Preview the CQL query without hitting the API
confluence-search search "migration guide" --dry-run
```

## Fetch

Fetches full page content as markdown. **Comments (inline + footer) are
included by default** — use `--no-comments` to skip them and save an API call.

```bash
# Full page with comments
confluence-search fetch 98765

# Page without comments (faster, less noise)
confluence-search fetch 98765 --no-comments

# JSON output — includes structured comments array
confluence-search fetch 98765 --json
```

## Global flags

All commands accept `--json` (machine-readable output) and `--timing`
(prints elapsed ms to stderr).

## Workflow tip

Search returns page URLs but content IDs may be absent in text mode.
Use `--json` to reliably get `content_id`, then pipe into fetch:

```bash
confluence-search search "topic" --json   # note content_id from results
confluence-search fetch <content-id>      # get full page + comments
```
