---
name: confluence-search
description: Search and fetch Confluence pages using the confluence-search CLI. Translates natural language queries to CQL, filters by space/label/date, and retrieves full page content as markdown. Use when the user asks about internal docs, processes, runbooks, or any Confluence-hosted knowledge.
argument-hint: "<query> [--spaces SPACE1,SPACE2] [--labels label1] [--limit N]"
---

# Confluence Search

Search and retrieve Confluence pages directly from the terminal using the
`confluence-search` CLI. No MCP server overhead — instant CQL translation
and REST API calls.

## Prerequisites

The `confluence-search` binary must be installed and these environment
variables set:

```
CONFLUENCE_URL=https://wiki.example.com
CONFLUENCE_PERSONAL_ACCESS_TOKEN=<your-pat>
```

## Quick Search

```bash
confluence-search search "<query>"
```

Returns the top 5 results with title, URL, space, and highlighted excerpt.

## Filtered Search

```bash
# Filter by space
confluence-search search "deployment" --spaces ENG,OPS

# Filter by label
confluence-search search "runbook" --labels oncall,production

# Search titles only (faster, more precise)
confluence-search search "API gateway" --titles-only

# More results
confluence-search search "onboarding" --limit 15

# Recent pages only
confluence-search search "incident" --modified-after 30d
```

## Fetch Full Page Content

After finding a page via search, fetch its full content by content ID:

```bash
confluence-search fetch <content-id>
```

The content ID appears in search results (JSON mode). This returns the full
page body converted to markdown with metadata (space, version, labels,
ancestor breadcrumb).

## JSON Output for Scripting

Add `--json` to any command for machine-readable output:

```bash
confluence-search search "architecture decision" --json
confluence-search fetch 12345 --json
```

## Dry Run (Preview CQL)

Preview the generated CQL without executing:

```bash
confluence-search search "migration guide" --dry-run
# Output: text ~ "migration guide" AND (type = "page" OR type = "blogpost") AND ...
```

## Time Filters

Use shorthand time ranges for the `--modified-after` and `--created-after` flags:

| Shorthand | Meaning |
|-----------|---------|
| `1d` | Last 24 hours |
| `7d` | Last 7 days |
| `30d` | Last 30 days |
| `90d` | Last 90 days |
| `6M` | Last 6 months |
| `1y` | Last year |
| `2y` | Last 2 years (default) |
| `5y` | Last 5 years |
| `2025-01-01` | Since specific date |

## Typical Workflow

```bash
# 1. Find relevant pages
confluence-search search "deployment checklist" --spaces ENG

# 2. Get content IDs from JSON output
confluence-search search "deployment checklist" --spaces ENG --json

# 3. Fetch the full page
confluence-search fetch 98765
```

## Health Check

Verify API connectivity and token validity:

```bash
confluence-search health
```
