# Agent Instructions

`confluence-search` is a Go CLI for searching and fetching Confluence pages.

Run `confluence-search --help` for the full command reference.

## Quick Start

```bash
confluence-search search "deployment process"          # search pages
confluence-search search "runbook" --spaces OPS --json  # filter + JSON
confluence-search fetch 12345                           # fetch page by ID
confluence-search health                                # check connectivity
```

## Environment

```
CONFLUENCE_URL=https://wiki.example.com
CONFLUENCE_PERSONAL_ACCESS_TOKEN=<token>
```

## Skills

See `.agents/skills/` for structured workflows:
- `confluence-search` — search and fetch operations
- `confluence-research` — deep multi-page research across spaces
