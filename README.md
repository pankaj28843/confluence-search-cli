# confluence-search

Fast Go CLI for searching and fetching Confluence pages. Translates natural
language queries to CQL, calls the Confluence REST API, and returns results
as text or JSON.

Built as a lightweight alternative to the
[confluence-mcp-server](https://github.com/pankaj28843/confluence-mcp-server)
Python MCP server — same capabilities, no server overhead.

## Install

```bash
git clone https://github.com/pankaj28843/confluence-search-cli.git
cd confluence-search-cli
make install    # installs to ~/.local/bin/confluence-search
```

Requires Go 1.22+ (auto-downloads Go 1.25 toolchain via `go.mod`).

## Setup

Set two environment variables:

```bash
export CONFLUENCE_URL=https://wiki.example.com
export CONFLUENCE_PERSONAL_ACCESS_TOKEN=your-token-here
```

Verify connectivity:

```bash
confluence-search health
```

## Usage

### Search

```bash
# Basic search
confluence-search search "deployment process"

# Filter by space
confluence-search search "API documentation" --spaces ENG,OPS

# Filter by label
confluence-search search "runbook" --labels oncall,production

# Titles only (faster)
confluence-search search "architecture decision record" --titles-only

# Recent pages
confluence-search search "incident postmortem" --modified-after 30d

# More results
confluence-search search "onboarding" --limit 15

# Preview CQL without executing
confluence-search search "migration guide" --dry-run

# JSON output
confluence-search search "design doc" --json
```

### Fetch

Retrieve full page content by content ID (from search results):

```bash
confluence-search fetch 12345
confluence-search fetch 12345 --json
```

### Time Filters

| Shorthand | Meaning |
|-----------|---------|
| `1d` | Last 24 hours |
| `7d` | Last 7 days |
| `30d` | Last 30 days |
| `90d` | Last 90 days |
| `1y` | Last year |
| `2y` | Last 2 years (default) |

Also accepts ISO dates: `--modified-after 2025-01-01`

## Agent Integration

This CLI is designed for use with coding agents (Claude Code, Cursor,
Windsurf, etc.). See [`AGENTS.md`](AGENTS.md) for tool descriptions
and [`.agents/skills/`](.agents/skills/) for structured workflows.

## Development

```bash
make build    # build binary
make test     # run tests
make clean    # remove binary
```

## License

MIT
