package confluence

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseSearchHit(raw json.RawMessage, baseURL string) (SearchResult, error) {
	var payload struct {
		Title        string `json:"title"`
		Excerpt      string `json:"excerpt"`
		LastModified string `json:"lastModified"`
		Content      *struct {
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

	hitURL := resolveURL(payload.Links.Base, payload.Links.WebUI, baseURL)
	if hitURL == "" {
		return SearchResult{}, fmt.Errorf("missing URL")
	}

	var contentID string
	if payload.Content != nil {
		contentID = payload.Content.ID
	}

	var spaceKey string
	if payload.Space != nil {
		spaceKey = payload.Space.Key
	}

	var containerTitle string
	if payload.ResultParentContainer != nil {
		containerTitle = payload.ResultParentContainer.Title
	} else if payload.ResultGlobalContainer != nil {
		containerTitle = payload.ResultGlobalContainer.Title
	}

	return SearchResult{
		ContentID:      contentID,
		Title:          payload.Title,
		Excerpt:        StripHTML(payload.Excerpt),
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

	if payload.Body != nil && payload.Body.Storage != nil {
		page.Markdown = HTMLToMarkdown(payload.Body.Storage.Value)
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
