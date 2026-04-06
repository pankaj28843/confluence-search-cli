---
name: confluence-research
description: Deep research across Confluence spaces using iterative search, page fetching, and cross-referencing. Builds comprehensive knowledge maps from internal documentation. Use for onboarding research, architecture discovery, process documentation, incident investigation, or any task requiring deep understanding of internal knowledge.
argument-hint: "<topic or question>"
---

# Confluence Deep Research

Conduct thorough research across Confluence by iteratively searching,
reading pages, following links, and synthesizing findings into a
structured report.

## Prerequisites

```
CONFLUENCE_URL=https://wiki.example.com
CONFLUENCE_PERSONAL_ACCESS_TOKEN=<your-pat>
```

## Phase 1: Discovery — Map the Knowledge Landscape

### 1a. Broad search to identify relevant spaces and pages

Start with a broad search to understand what exists:

```bash
confluence-search search "<topic>" --limit 15 --json
```

Parse the JSON to identify:
- Which **spaces** contain relevant content
- Common **labels** used for this topic
- The most recent and most relevant pages

### 1b. Narrow by space and recency

Once you identify key spaces, run focused searches:

```bash
# Search within specific spaces
confluence-search search "<refined query>" --spaces ENG,ARCH --limit 10 --json

# Find recently updated pages (likely most current)
confluence-search search "<topic>" --modified-after 90d --json

# Search by label if you found common tags
confluence-search search "<topic>" --labels architecture,design --json
```

### 1c. Title search for known document types

For standard document types, use title-only search:

```bash
confluence-search search "ADR" --titles-only --spaces ARCH --json
confluence-search search "runbook" --titles-only --spaces OPS --json
confluence-search search "RFC" --titles-only --json
```

## Phase 2: Deep Read — Extract Knowledge from Key Pages

### 2a. Fetch full content of the most relevant pages

For each promising search result, fetch the full page:

```bash
confluence-search fetch <content-id>
```

### 2b. Extract and organize

From each page, extract:
- **Decisions** — what was decided and why
- **Architecture** — system components, data flows, dependencies
- **Processes** — step-by-step procedures, checklists
- **Contacts** — team owners, escalation paths
- **Links** — references to other pages, external resources, code repos

### 2c. Follow the trail

Pages often reference other pages. When you find references to other
Confluence pages, search for them:

```bash
# If a page mentions "API Gateway Design Doc"
confluence-search search "API Gateway Design Doc" --titles-only --json
```

## Phase 3: Cross-Reference and Validate

### 3a. Check for contradictions

Search for the same topic across different spaces — teams sometimes
have conflicting documentation:

```bash
confluence-search search "deployment process" --spaces ENG --json
confluence-search search "deployment process" --spaces OPS --json
confluence-search search "deployment process" --spaces SRE --json
```

### 3b. Check recency

For critical information, verify when pages were last updated:

```bash
confluence-search search "<topic>" --modified-after 30d --json
```

Pages not updated in 1+ years may be stale.

## Phase 4: Synthesize — Build the Research Report

Produce a structured report:

```markdown
# Confluence Research: {Topic}

**Sources:** {N} pages across {M} spaces
**Last updated pages:** {date range}

## Summary
3-5 bullet points covering the key findings.

## Key Pages
| Page | Space | Last Modified | Why It Matters |
|------|-------|---------------|----------------|
| {title} | {space} | {date} | {relevance} |

## Findings

### {Finding 1}
{What was learned, citing specific pages}

### {Finding 2}
...

## Gaps and Concerns
- {Missing documentation}
- {Stale pages that need updating}
- {Contradictions between spaces}

## Recommended Actions
1. {Action item based on findings}
```

## Tips for Effective Confluence Research

- **Start broad, then narrow** — initial search reveals which spaces matter
- **Title search is faster** — use `--titles-only` when you know the doc type
- **Labels are gold** — once you find good labels, filter by them
- **Check multiple spaces** — different teams document differently
- **90-day filter catches active docs** — if nobody updated it recently, it may be stale
- **Run searches in parallel** — launch multiple Bash calls for different query variants
- **JSON output enables scripting** — parse with `jq` for batch operations
- **Content IDs are stable** — bookmark them for repeated access
