package confluence

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func parseSearchHit(raw json.RawMessage, baseURL string) (SearchResult, error) {
	var payload struct {
		Title         string `json:"title"`
		Excerpt       string `json:"excerpt"`
		URL           string `json:"url"`
		LastModified  string `json:"lastModified"`
		Content       *struct {
			ID string `json:"id"`
		} `json:"content"`
		Space *struct {
			Key string `json:"key"`
		} `json:"space"`
		ResultParentContainer *struct {
			Title string `json:"title"`
		} `json:"resultParentContainer"`
		ResultGlobalContainer *struct {
			Title string `json:"title"`
		} `json:"resultGlobalContainer"`
		Links struct {
			WebUI string `json:"webui"`
			Base  string `json:"base"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return SearchResult{}, err
	}

	// Resolve URL
	hitURL := resolveURL(payload.Links.Base, payload.Links.WebUI, baseURL)
	if hitURL == "" {
		return SearchResult{}, fmt.Errorf("missing URL")
	}

	// Clean excerpt
	excerpt := stripHTML(payload.Excerpt)

	// Extract content ID
	var contentID string
	if payload.Content != nil {
		contentID = payload.Content.ID
	}

	// Extract space key
	var spaceKey string
	if payload.Space != nil {
		spaceKey = payload.Space.Key
	}

	// Extract container title
	var containerTitle string
	if payload.ResultParentContainer != nil {
		containerTitle = payload.ResultParentContainer.Title
	} else if payload.ResultGlobalContainer != nil {
		containerTitle = payload.ResultGlobalContainer.Title
	}

	return SearchResult{
		ContentID:      contentID,
		Title:          payload.Title,
		Excerpt:        excerpt,
		URL:            hitURL,
		SpaceKey:       spaceKey,
		ContainerTitle: containerTitle,
		LastModified:   payload.LastModified,
	}, nil
}

func parseContentPage(body []byte, baseURL string) (*ContentPage, error) {
	var payload struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Body  *struct {
			Storage *struct {
				Value string `json:"value"`
			} `json:"storage"`
		} `json:"body"`
		Space *struct {
			Key string `json:"key"`
		} `json:"space"`
		Version *struct {
			Number int    `json:"number"`
			When   string `json:"when"`
		} `json:"version"`
		Metadata *struct {
			Labels *struct {
				Results []struct {
					Name string `json:"name"`
				} `json:"results"`
			} `json:"labels"`
		} `json:"metadata"`
		Ancestors []struct {
			Title string `json:"title"`
		} `json:"ancestors"`
		Links struct {
			WebUI string `json:"webui"`
			Base  string `json:"base"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse content: %w", err)
	}

	page := &ContentPage{
		ID:    payload.ID,
		Title: payload.Title,
		URL:   resolveURL(payload.Links.Base, payload.Links.WebUI, baseURL),
	}

	// HTML → markdown
	if payload.Body != nil && payload.Body.Storage != nil {
		page.Markdown = htmlToMarkdown(payload.Body.Storage.Value)
	}

	if payload.Space != nil {
		page.SpaceKey = payload.Space.Key
	}

	if payload.Version != nil {
		page.Version = payload.Version.Number
		page.LastModified = payload.Version.When
	}

	if payload.Metadata != nil && payload.Metadata.Labels != nil {
		for _, l := range payload.Metadata.Labels.Results {
			page.Labels = append(page.Labels, l.Name)
		}
	}

	for _, a := range payload.Ancestors {
		page.Ancestors = append(page.Ancestors, a.Title)
	}

	return page, nil
}

func resolveURL(base, webUI, fallbackBase string) string {
	if base != "" && webUI != "" {
		return base + webUI
	}
	if webUI != "" {
		return strings.TrimRight(fallbackBase, "/") + webUI
	}
	return ""
}

func stripHTML(s string) string {
	// Replace highlight spans with markdown bold
	s = strings.ReplaceAll(s, `<span class="search-hit">`, "**")
	s = strings.ReplaceAll(s, "</span>", "**")
	return strings.TrimSpace(htmlTagRe.ReplaceAllString(s, ""))
}

func htmlToMarkdown(html string) string {
	if html == "" {
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return basicHTMLToMarkdown(html)
	}

	var sb strings.Builder
	doc.Find("body").Each(func(_ int, s *goquery.Selection) {
		renderNode(&sb, s)
	})

	result := sb.String()
	if result == "" {
		// goquery wraps in <html><body>, try direct content
		renderNode(&sb, doc.Selection)
		result = sb.String()
	}
	if result == "" {
		return basicHTMLToMarkdown(html)
	}
	return strings.TrimSpace(result)
}

func renderNode(sb *strings.Builder, s *goquery.Selection) {
	s.Children().Each(func(_ int, child *goquery.Selection) {
		tag := goquery.NodeName(child)
		text := strings.TrimSpace(child.Text())

		switch tag {
		case "h1":
			sb.WriteString("# " + text + "\n\n")
		case "h2":
			sb.WriteString("## " + text + "\n\n")
		case "h3":
			sb.WriteString("### " + text + "\n\n")
		case "h4":
			sb.WriteString("#### " + text + "\n\n")
		case "p":
			sb.WriteString(text + "\n\n")
		case "ul":
			child.Find("li").Each(func(_ int, li *goquery.Selection) {
				sb.WriteString("- " + strings.TrimSpace(li.Text()) + "\n")
			})
			sb.WriteString("\n")
		case "ol":
			child.Find("li").Each(func(i int, li *goquery.Selection) {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(li.Text())))
			})
			sb.WriteString("\n")
		case "pre":
			code := child.Find("code").Text()
			if code == "" {
				code = text
			}
			sb.WriteString("```\n" + code + "\n```\n\n")
		case "table":
			renderTable(sb, child)
		case "a":
			href, _ := child.Attr("href")
			if href != "" {
				sb.WriteString(fmt.Sprintf("[%s](%s)", text, href))
			} else {
				sb.WriteString(text)
			}
		case "strong", "b":
			sb.WriteString("**" + text + "**")
		case "em", "i":
			sb.WriteString("*" + text + "*")
		case "code":
			sb.WriteString("`" + text + "`")
		case "br":
			sb.WriteString("\n")
		case "hr":
			sb.WriteString("\n---\n\n")
		case "div", "section", "article":
			renderNode(sb, child)
		default:
			if text != "" {
				sb.WriteString(text)
			}
		}
	})
}

func renderTable(sb *strings.Builder, table *goquery.Selection) {
	var headers []string
	table.Find("thead tr th, thead tr td").Each(func(_ int, th *goquery.Selection) {
		headers = append(headers, strings.TrimSpace(th.Text()))
	})

	if len(headers) > 0 {
		sb.WriteString("| " + strings.Join(headers, " | ") + " |\n")
		sep := make([]string, len(headers))
		for i := range sep {
			sep[i] = "---"
		}
		sb.WriteString("| " + strings.Join(sep, " | ") + " |\n")
	}

	table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
		var cells []string
		tr.Find("td, th").Each(func(_ int, td *goquery.Selection) {
			cells = append(cells, strings.TrimSpace(td.Text()))
		})
		if len(cells) > 0 {
			sb.WriteString("| " + strings.Join(cells, " | ") + " |\n")
		}
	})
	sb.WriteString("\n")
}

func basicHTMLToMarkdown(html string) string {
	s := html
	s = strings.ReplaceAll(s, "<h1>", "# ")
	s = strings.ReplaceAll(s, "</h1>", "\n\n")
	s = strings.ReplaceAll(s, "<h2>", "## ")
	s = strings.ReplaceAll(s, "</h2>", "\n\n")
	s = strings.ReplaceAll(s, "<b>", "**")
	s = strings.ReplaceAll(s, "</b>", "**")
	s = strings.ReplaceAll(s, "<strong>", "**")
	s = strings.ReplaceAll(s, "</strong>", "**")
	return strings.TrimSpace(htmlTagRe.ReplaceAllString(s, ""))
}

// URLEncode encodes a string for use in URLs.
func URLEncode(s string) string {
	return url.QueryEscape(s)
}
