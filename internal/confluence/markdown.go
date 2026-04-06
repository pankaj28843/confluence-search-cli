package confluence

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// HTMLToMarkdown converts Confluence storage HTML to readable markdown.
func HTMLToMarkdown(html string) string {
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
		renderNode(&sb, doc.Selection)
		result = sb.String()
	}
	if result == "" {
		return basicHTMLToMarkdown(html)
	}
	return strings.TrimSpace(result)
}

// StripHTML removes HTML tags, converting highlight spans to bold.
func StripHTML(s string) string {
	s = strings.ReplaceAll(s, `<span class="search-hit">`, "**")
	s = strings.ReplaceAll(s, "</span>", "**")
	return strings.TrimSpace(htmlTagRe.ReplaceAllString(s, ""))
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
