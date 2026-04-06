---
name: confluence-research
description: Deep iterative research across Confluence wiki. Searches multiple spaces, fetches key pages, cross-references findings, synthesizes a report. Use for onboarding, architecture discovery, incident investigation, ADR review, or understanding internal systems.
argument-hint: "<topic or question>"
user-invocable: true
allowed-tools: "Bash, Read, Write, Agent"
---

# Confluence Deep Research

Run `confluence-search --help` for the full CLI reference.

## Workflow

### 1. Discover — find relevant spaces and pages

```bash
confluence-search search "<topic>" --limit 15 --json
```

From results, note which spaces and labels appear. Then narrow:

```bash
confluence-search search "<topic>" --spaces ENG,ARCH --json
confluence-search search "ADR" --titles-only --spaces ARCH --json
confluence-search search "<topic>" --modified-after 90d --json
```

Run multiple searches in parallel for different query variants.

### 2. Read — fetch key pages

```bash
confluence-search fetch <content-id>
```

Extract: decisions, architecture, processes, contacts, links to other pages.
When a page references another, search for it by title:

```bash
confluence-search search "referenced page title" --titles-only --json
```

### 3. Cross-reference — check for conflicts and staleness

Search the same topic across different spaces:

```bash
confluence-search search "release process" --spaces ENG --json
confluence-search search "release process" --spaces OPS --json
```

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
{Details citing pages}

## Gaps
- Missing docs, stale pages, contradictions
```

## Tips

- Start broad, then narrow by space/label
- `--titles-only` for known doc types (ADR, RFC, runbook)
- `--modified-after 90d` finds actively maintained pages
- JSON output + parallel searches for speed
