package confluence

import (
	"encoding/json"
	"testing"
)

func TestParseSearchHit(t *testing.T) {
	raw := json.RawMessage(`{
		"title": "Deploy Guide",
		"excerpt": "How to <span class=\"search-hit\">deploy</span>",
		"lastModified": "2026-01-01T00:00:00Z",
		"content": {"id": "12345"},
		"space": {"key": "ENG"},
		"resultParentContainer": {"title": "Operations"},
		"_links": {"webui": "/display/ENG/Deploy+Guide", "base": "https://wiki.example.com"}
	}`)

	hit, err := parseSearchHit(raw, "https://wiki.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if hit.Title != "Deploy Guide" {
		t.Errorf("title = %q", hit.Title)
	}
	if hit.ContentID != "12345" {
		t.Errorf("content_id = %q", hit.ContentID)
	}
	if hit.SpaceKey != "ENG" {
		t.Errorf("space_key = %q", hit.SpaceKey)
	}
	if hit.URL != "https://wiki.example.com/display/ENG/Deploy+Guide" {
		t.Errorf("url = %q", hit.URL)
	}
	if hit.ContainerTitle != "Operations" {
		t.Errorf("container = %q", hit.ContainerTitle)
	}
	// Excerpt should have HTML stripped, highlights converted to bold
	if hit.Excerpt != "How to **deploy**" {
		t.Errorf("excerpt = %q", hit.Excerpt)
	}
}

func TestParseSearchHitTopLevelID(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "67890",
		"title": "Top Level ID",
		"excerpt": "",
		"_links": {"webui": "/display/ENG/Top", "base": "https://wiki.example.com"}
	}`)
	hit, err := parseSearchHit(raw, "https://wiki.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if hit.ContentID != "67890" {
		t.Errorf("content_id = %q, want 67890", hit.ContentID)
	}
}

func TestParseSearchHitMissingURL(t *testing.T) {
	raw := json.RawMessage(`{"title": "No URL", "_links": {}}`)
	_, err := parseSearchHit(raw, "https://wiki.example.com")
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestParseContentPage(t *testing.T) {
	body := []byte(`{
		"id": "99",
		"title": "Architecture",
		"body": {"storage": {"value": "<h1>Overview</h1><p>System design.</p>"}},
		"space": {"key": "ARCH"},
		"version": {"number": 3, "when": "2026-01-01T00:00:00Z"},
		"metadata": {"labels": {"results": [{"name": "rfc"}, {"name": "approved"}]}},
		"ancestors": [{"title": "Root"}, {"title": "Engineering"}],
		"_links": {"webui": "/display/ARCH/Architecture", "base": "https://wiki.example.com"}
	}`)

	page, err := parseContentPage(body, "https://wiki.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if page.ID != "99" {
		t.Errorf("id = %q", page.ID)
	}
	if page.SpaceKey != "ARCH" {
		t.Errorf("space = %q", page.SpaceKey)
	}
	if page.Version != 3 {
		t.Errorf("version = %d", page.Version)
	}
	if len(page.Labels) != 2 || page.Labels[0] != "rfc" {
		t.Errorf("labels = %v", page.Labels)
	}
	if len(page.Ancestors) != 2 || page.Ancestors[0] != "Root" {
		t.Errorf("ancestors = %v", page.Ancestors)
	}
	// Markdown should be converted from HTML
	if page.Markdown == "" {
		t.Error("expected markdown content")
	}
}

func TestParseComments(t *testing.T) {
	body := []byte(`{
		"results": [
			{
				"id": "1001",
				"version": {"by": {"displayName": "Jane Doe"}, "when": "2026-01-15T10:30:00Z"},
				"body": {"view": {"value": "<p>This needs updating.</p>"}},
				"extensions": {"location": "footer"}
			},
			{
				"id": "1002",
				"version": {"by": {"displayName": "Bob Smith"}, "when": "2026-01-20T14:00:00Z"},
				"body": {"view": {"value": "<p>Done, updated in v5.</p>"}},
				"extensions": {"location": "footer"}
			},
			{
				"id": "1003",
				"version": {"by": {"displayName": "Alice Chen"}, "when": "2026-02-01T09:00:00Z"},
				"body": {"view": {"value": "<p>This section is outdated.</p>"}},
				"extensions": {
					"location": "inline",
					"inlineProperties": {"originalSelection": "the deployment steps"},
					"resolution": {"status": "resolved"}
				}
			}
		],
		"start": 0, "limit": 25, "size": 3
	}`)

	comments, err := parseComments(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 3 {
		t.Fatalf("expected 3 comments, got %d", len(comments))
	}

	// Footer comment
	if comments[0].Author != "Jane Doe" {
		t.Errorf("comment[0] author = %q", comments[0].Author)
	}
	if comments[0].Location != "footer" {
		t.Errorf("comment[0] location = %q", comments[0].Location)
	}
	if comments[0].Body == "" {
		t.Error("comment[0] body is empty")
	}

	// Inline comment with resolution
	if comments[2].Location != "inline" {
		t.Errorf("comment[2] location = %q", comments[2].Location)
	}
	if comments[2].InlineOriginalSelection != "the deployment steps" {
		t.Errorf("comment[2] originalSelection = %q", comments[2].InlineOriginalSelection)
	}
	if !comments[2].Resolved {
		t.Error("comment[2] should be resolved")
	}
}

func TestParseCommentsEmpty(t *testing.T) {
	body := []byte(`{"results": [], "start": 0, "limit": 25, "size": 0}`)
	comments, err := parseComments(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}

func TestResolveURL(t *testing.T) {
	cases := []struct {
		base, webUI, fallback, want string
	}{
		{"https://wiki.com", "/page", "https://other.com", "https://wiki.com/page"},
		{"", "/page", "https://other.com", "https://other.com/page"},
		{"", "", "https://other.com", ""},
	}
	for _, tc := range cases {
		got := resolveURL(tc.base, tc.webUI, tc.fallback)
		if got != tc.want {
			t.Errorf("resolveURL(%q, %q, %q) = %q, want %q", tc.base, tc.webUI, tc.fallback, got, tc.want)
		}
	}
}
