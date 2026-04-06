# confluence-search-cli

Go CLI wrapping the Confluence REST API for fast search and page retrieval.

## Build & Test

```bash
make build      # go build -o confluence-search ./cmd/confluence-search/
make test       # go test ./... -v
make install    # install to ~/.local/bin/
```

## Architecture

- `cmd/confluence-search/` — Cobra CLI entry point (search, fetch, health)
- `internal/confluence/` — REST API client + JSON parsing + HTML→markdown
- `internal/cql/` — Natural language → CQL translation
- `internal/output/` — JSON/text formatting

## Key Patterns

- Stateless CLI — no caching, no persistent state
- All auth via env vars: `CONFLUENCE_URL`, `CONFLUENCE_PERSONAL_ACCESS_TOKEN`
- HTML→markdown via goquery; fallback to regex stripping
- CQL time range shorthands: 1d, 7d, 30d, 90d, 6M, 1y, 2y, 5y
