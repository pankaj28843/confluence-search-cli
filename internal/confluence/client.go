package confluence

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a lightweight Confluence REST API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	UserAgent  string
}

// NewClient creates a client from environment config.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		UserAgent: "confluence-search-cli/0.1",
	}
}

// SearchResult represents a single search hit from Confluence.
type SearchResult struct {
	ContentID      string `json:"content_id,omitempty"`
	Title          string `json:"title"`
	Excerpt        string `json:"excerpt"`
	URL            string `json:"url"`
	SpaceKey       string `json:"space_key,omitempty"`
	ContainerTitle string `json:"container_title,omitempty"`
	LastModified   string `json:"last_modified,omitempty"`
}

// SearchResponse wraps search results.
type SearchResponse struct {
	Results   []SearchResult `json:"results"`
	NextStart *int           `json:"next_start,omitempty"`
}

// Comment represents a single Confluence page comment (inline or footer).
type Comment struct {
	ID                      string `json:"id"`
	Author                  string `json:"author"`
	Date                    string `json:"date"`
	Body                    string `json:"body"`
	Location                string `json:"location"`                           // "inline", "footer", or "resolved"
	InlineOriginalSelection string `json:"inline_original_selection,omitempty"` // text the inline comment is attached to
	Resolved                bool   `json:"resolved,omitempty"`
}

// ContentPage represents a fetched Confluence page.
type ContentPage struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	Markdown     string    `json:"markdown"`
	SpaceKey     string    `json:"space_key,omitempty"`
	Version      int       `json:"version,omitempty"`
	LastModified string    `json:"last_modified,omitempty"`
	Labels       []string  `json:"labels,omitempty"`
	Ancestors    []string  `json:"ancestors,omitempty"`
	Comments     []Comment `json:"comments,omitempty"`
}

// FormatMarkdown renders the page as a self-contained markdown document
// with metadata header. Matches the Python MCP server's output format.
func (p *ContentPage) FormatMarkdown() string {
	var sb strings.Builder
	sb.WriteString("# " + p.Title + "\n\n")
	sb.WriteString("**URL:** " + p.URL + "\n")
	if p.SpaceKey != "" {
		sb.WriteString("**Space:** " + p.SpaceKey + "\n")
	}
	if p.Version > 0 {
		sb.WriteString(fmt.Sprintf("**Version:** %d\n", p.Version))
	}
	if p.LastModified != "" {
		sb.WriteString("**Last modified:** " + p.LastModified + "\n")
	}
	if len(p.Labels) > 0 {
		sb.WriteString("**Labels:** " + strings.Join(p.Labels, ", ") + "\n")
	}
	if len(p.Ancestors) > 0 {
		sb.WriteString("**Ancestors:** " + strings.Join(p.Ancestors, " > ") + "\n")
	}
	sb.WriteString("\n---\n\n")
	if p.Markdown != "" {
		sb.WriteString(p.Markdown)
	}
	if len(p.Comments) > 0 {
		sb.WriteString("\n\n---\n\n")
		sb.WriteString(fmt.Sprintf("## Comments (%d)\n\n", len(p.Comments)))

		var footer, inline []Comment
		for _, c := range p.Comments {
			if c.Location == "inline" {
				inline = append(inline, c)
			} else {
				footer = append(footer, c)
			}
		}

		if len(footer) > 0 {
			sb.WriteString("### Footer Comments\n\n")
			for _, c := range footer {
				writeComment(&sb, c)
			}
		}
		if len(inline) > 0 {
			sb.WriteString("### Inline Comments\n\n")
			for _, c := range inline {
				if c.InlineOriginalSelection != "" {
					sb.WriteString("> \"" + c.InlineOriginalSelection + "\"\n\n")
				}
				writeComment(&sb, c)
			}
		}
	}
	return sb.String()
}

func writeComment(sb *strings.Builder, c Comment) {
	date := c.Date
	if len(date) >= 10 {
		date = date[:10]
	}
	header := fmt.Sprintf("**%s** (%s)", c.Author, date)
	if c.Resolved {
		header += " [resolved]"
	}
	sb.WriteString(header + ":\n")
	sb.WriteString(c.Body + "\n\n")
}

// Search executes a CQL search against Confluence REST API.
func (c *Client) Search(cql string, limit int) (*SearchResponse, error) {
	params := url.Values{
		"cql":     {cql},
		"limit":   {fmt.Sprintf("%d", limit)},
		"excerpt": {"highlight"},
	}

	body, err := c.get("/content/search", params)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var raw struct {
		Results []json.RawMessage `json:"results"`
		Links   struct {
			Next string `json:"next"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	resp := &SearchResponse{}
	for _, rawHit := range raw.Results {
		hit, err := parseSearchHit(rawHit, c.BaseURL)
		if err != nil {
			continue
		}
		resp.Results = append(resp.Results, hit)
	}
	return resp, nil
}

// FetchContent retrieves a page by content ID with full body.
func (c *Client) FetchContent(contentID string) (*ContentPage, error) {
	params := url.Values{
		"expand": {"body.storage,metadata.labels,version,space,ancestors"},
	}

	body, err := c.get("/content/"+url.PathEscape(contentID), params)
	if err != nil {
		return nil, fmt.Errorf("fetch content: %w", err)
	}

	return parseContentPage(body, c.BaseURL)
}

// FetchComments retrieves comments (inline + footer + resolved) for a page.
func (c *Client) FetchComments(contentID string) ([]Comment, error) {
	params := url.Values{
		"expand":   {"body.view,version,extensions.inlineProperties,extensions.resolution"},
		"location": {"footer", "inline", "resolved"},
		"limit":    {"100"},
	}

	body, err := c.get("/content/"+url.PathEscape(contentID)+"/child/comment", params)
	if err != nil {
		return nil, fmt.Errorf("fetch comments: %w", err)
	}

	return parseComments(body)
}

// HealthCheck performs a minimal API call to verify connectivity.
func (c *Client) HealthCheck() error {
	params := url.Values{
		"cql":   {"type=page"},
		"limit": {"1"},
	}
	_, err := c.get("/content/search", params)
	return err
}

func (c *Client) get(path string, params url.Values) ([]byte, error) {
	u := fmt.Sprintf("%s/rest/api%s?%s", c.BaseURL, path, params.Encode())

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return body, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
