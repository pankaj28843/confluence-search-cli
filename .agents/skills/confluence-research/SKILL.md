---
name: confluence-research
description: Deep iterative research across Confluence wiki. Searches multiple spaces, fetches key pages, cross-references findings, synthesizes a report. Use for onboarding, architecture discovery, incident investigation, ADR review, or understanding internal systems.
argument-hint: "<topic or question>"
user-invocable: true
---

# Confluence Deep Research

**First: run `confluence-search --help` to see all current commands and flags.**
The CLI evolves; `--help` is always the source of truth.

## Workflow

### 1. Discover — broad search, then narrow

```bash
# Start broad with JSON to get content IDs and space keys
confluence-search search "<topic>" --limit 15 --json

# Narrow by space, labels, or recency
confluence-search search "<topic>" --spaces ENG,ARCH --json
confluence-search search "ADR" --titles-only --spaces ARCH --json
confluence-search search "<topic>" --modified-after 90d --json
```

Run multiple searches in parallel for different query variants.

### 2. Read — fetch key pages

```bash
# Full page with comments (comments often have critical context)
confluence-search fetch <content-id>

# Content only, skip comments (faster when you just need the body)
confluence-search fetch <content-id> --no-comments
```

Comments on ADRs, RFCs, and runbooks frequently contain decisions,
objections, and clarifications not in the page body. **Default to
including comments** unless you are only extracting structured content.

When a page references another page, search for it by title:

```bash
confluence-search search "referenced page title" --titles-only --json
```

### 3. Cross-reference — check for conflicts and staleness

Search the same topic across different spaces:

```bash
confluence-search search "release process" --spaces ENG --json
confluence-search search "release process" --spaces OPS --json
```

Compare version dates (`last_modified` in JSON) to identify stale docs.

### 4. Synthesize — write a report

```markdown
# Research: {Topic}

## Summary
- Key findings (3-5 bullets)

## Key Pages
| Page | Space | Modified | Relevance |
|------|-------|----------|-----------|

## Findings
### {Finding}
{Details citing pages by title and URL}

## Gaps
- Missing docs, stale pages, contradictions
```

## Tips

- `--json` for all searches — gives `content_id`, `space_key`, `last_modified`
- `--titles-only` for known doc types (ADR, RFC, runbook)
- `--modified-after 90d` finds actively maintained pages
- `--no-comments` when fetching many pages for content scanning (saves time)
- `--timing` to monitor API latency if responses feel slow
- Run parallel searches across spaces to maximize coverage
