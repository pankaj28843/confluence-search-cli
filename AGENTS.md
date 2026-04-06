# Agent Instructions

This repository provides `confluence-search`, a Go CLI for searching and
fetching Confluence pages via the REST API.

## Available Tools

### confluence-search search "<query>"

Search Confluence pages. Translates natural language to CQL.

**Flags:**
- `--limit N` — max results (default: 5, max: 25)
- `--spaces ENG,OPS` — filter by space keys
- `--labels label1,label2` — filter by page labels
- `--titles-only` — search page titles only
- `--modified-after 30d` — time filter (1d/7d/30d/90d/6M/1y/2y/5y or ISO date)
- `--created-after 2025-01-01` — creation date filter
- `--dry-run` — show CQL without executing
- `--json` — JSON output

### confluence-search fetch <content-id>

Fetch full page content by numeric ID. Returns markdown with metadata.

**Flags:**
- `--json` — JSON output

### confluence-search health

Check API connectivity and token validity.

## Environment Variables

```
CONFLUENCE_URL=https://wiki.example.com           # Required
CONFLUENCE_PERSONAL_ACCESS_TOKEN=<token>           # Required
```

## Skills

See `.agents/skills/` for structured workflows:
- `confluence-search` — basic search and fetch operations
- `confluence-research` — deep multi-page research across spaces

## Install

```bash
make install
```
