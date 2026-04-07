package confluence

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdownHeadings(t *testing.T) {
	html := "<h1>Title</h1><h2>Subtitle</h2><p>Body text.</p>"
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "# Title") {
		t.Errorf("expected h1 heading, got: %s", md)
	}
	if !strings.Contains(md, "## Subtitle") {
		t.Errorf("expected h2 heading, got: %s", md)
	}
	if !strings.Contains(md, "Body text.") {
		t.Errorf("expected body, got: %s", md)
	}
}

func TestHTMLToMarkdownCodeBlock(t *testing.T) {
	html := `<pre><code>func main() {}</code></pre>`
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "```") {
		t.Errorf("expected code fence, got: %s", md)
	}
	if !strings.Contains(md, "func main()") {
		t.Errorf("expected code content, got: %s", md)
	}
}

func TestHTMLToMarkdownList(t *testing.T) {
	html := "<ul><li>Item A</li><li>Item B</li></ul>"
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "- Item A") {
		t.Errorf("expected bullet list, got: %s", md)
	}
}

func TestHTMLToMarkdownEmpty(t *testing.T) {
	if HTMLToMarkdown("") != "" {
		t.Error("expected empty output for empty input")
	}
}

func TestHTMLToMarkdownFallback(t *testing.T) {
	// Plain text with tags should still work via fallback
	html := "<b>bold</b> and <strong>strong</strong>"
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "**bold**") {
		t.Errorf("expected bold markdown, got: %s", md)
	}
}

func TestStripHTML(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{`plain text`, "plain text"},
		{`<b>bold</b>`, "bold"},
		{`<span class="search-hit">term</span>`, "**term**"},
		{`<p>paragraph</p>`, "paragraph"},
		{"", ""},
	}
	for _, tc := range cases {
		got := StripHTML(tc.input)
		if got != tc.want {
			t.Errorf("StripHTML(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatMarkdownWithComments(t *testing.T) {
	page := &ContentPage{
		Title:    "Test Page",
		URL:      "https://wiki.example.com/pages/123",
		SpaceKey: "ENG",
		Markdown: "# Content\n\nBody text.",
		Comments: []Comment{
			{ID: "1", Author: "Jane", Date: "2026-01-15T10:30:00Z", Body: "Looks good.", Location: "footer"},
			{ID: "2", Author: "Alice", Date: "2026-02-01T09:00:00Z", Body: "Outdated section.", Location: "inline", InlineOriginalSelection: "deployment steps", Resolved: true},
		},
	}
	md := page.FormatMarkdown()
	if !strings.Contains(md, "## Comments (2)") {
		t.Error("expected comments header")
	}
	if !strings.Contains(md, "### Footer Comments") {
		t.Error("expected footer section")
	}
	if !strings.Contains(md, "### Inline Comments") {
		t.Error("expected inline section")
	}
	if !strings.Contains(md, "**Jane** (2026-01-15)") {
		t.Error("expected footer comment author/date")
	}
	if !strings.Contains(md, "> \"deployment steps\"") {
		t.Error("expected inline original selection quote")
	}
	if !strings.Contains(md, "[resolved]") {
		t.Error("expected resolved label")
	}
}

func TestContentPageFormatMarkdown(t *testing.T) {
	page := &ContentPage{
		Title:        "Test Page",
		URL:          "https://wiki.example.com/pages/123",
		SpaceKey:     "ENG",
		Version:      5,
		LastModified: "2026-01-01T00:00:00Z",
		Labels:       []string{"architecture", "rfc"},
		Ancestors:    []string{"Root", "Engineering"},
		Markdown:     "# Content\n\nBody text.",
	}
	md := page.FormatMarkdown()
	if !strings.Contains(md, "# Test Page") {
		t.Error("expected title")
	}
	if !strings.Contains(md, "**Space:** ENG") {
		t.Error("expected space")
	}
	if !strings.Contains(md, "**Labels:** architecture, rfc") {
		t.Error("expected labels")
	}
	if !strings.Contains(md, "Root > Engineering") {
		t.Error("expected ancestors")
	}
	if !strings.Contains(md, "Body text.") {
		t.Error("expected body content")
	}
}
