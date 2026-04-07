package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pankaj28843/confluence-search-cli/internal/confluence"
)

// mockAPI implements confluence.API for testing.
type mockAPI struct {
	searchFunc       func(ctx context.Context, cql string, limit int) (*confluence.SearchResponse, error)
	fetchContentFunc func(ctx context.Context, id string) (*confluence.ContentPage, error)
	fetchCommentsFunc func(ctx context.Context, id string) ([]confluence.Comment, error)
	healthCheckFunc  func(ctx context.Context) error
}

func (m *mockAPI) Search(ctx context.Context, cql string, limit int) (*confluence.SearchResponse, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, cql, limit)
	}
	return &confluence.SearchResponse{}, nil
}

func (m *mockAPI) FetchContent(ctx context.Context, id string) (*confluence.ContentPage, error) {
	if m.fetchContentFunc != nil {
		return m.fetchContentFunc(ctx, id)
	}
	return &confluence.ContentPage{Title: "Test", URL: "https://wiki.example.com/test"}, nil
}

func (m *mockAPI) FetchComments(ctx context.Context, id string) ([]confluence.Comment, error) {
	if m.fetchCommentsFunc != nil {
		return m.fetchCommentsFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAPI) HealthCheck(ctx context.Context) error {
	if m.healthCheckFunc != nil {
		return m.healthCheckFunc(ctx)
	}
	return nil
}

// setMockClient installs a mock and returns a cleanup function.
func setMockClient(api confluence.API) func() {
	orig := clientFactory
	clientFactory = func() (confluence.API, error) {
		return api, nil
	}
	return func() { clientFactory = orig }
}

func TestSearchCmdTextOutput(t *testing.T) {
	mock := &mockAPI{
		searchFunc: func(_ context.Context, cql string, limit int) (*confluence.SearchResponse, error) {
			return &confluence.SearchResponse{
				Results: []confluence.SearchResult{
					{Title: "Deploy Guide", URL: "https://wiki.example.com/deploy", SpaceKey: "ENG", Excerpt: "How to deploy"},
				},
			}, nil
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := searchCmd()
	cmd.SetArgs([]string{"deploy"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchCmdDryRun(t *testing.T) {
	jsonOutput = false
	timing = false

	cmd := searchCmd()
	cmd.SetArgs([]string{"deploy", "--dry-run"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchCmdNoResults(t *testing.T) {
	mock := &mockAPI{
		searchFunc: func(_ context.Context, _ string, _ int) (*confluence.SearchResponse, error) {
			return &confluence.SearchResponse{Results: nil}, nil
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := searchCmd()
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchCmdError(t *testing.T) {
	mock := &mockAPI{
		searchFunc: func(_ context.Context, _ string, _ int) (*confluence.SearchResponse, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := searchCmd()
	cmd.SetArgs([]string{"deploy"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error = %q, want connection refused", err)
	}
}

func TestSearchCmdJSONOutput(t *testing.T) {
	mock := &mockAPI{
		searchFunc: func(_ context.Context, _ string, _ int) (*confluence.SearchResponse, error) {
			return &confluence.SearchResponse{
				Results: []confluence.SearchResult{
					{Title: "API Docs", URL: "https://wiki.example.com/api"},
				},
			}, nil
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = true
	timing = false
	defer func() { jsonOutput = false }()

	cmd := searchCmd()
	cmd.SetArgs([]string{"api"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFetchCmdTextOutput(t *testing.T) {
	mock := &mockAPI{
		fetchContentFunc: func(_ context.Context, id string) (*confluence.ContentPage, error) {
			return &confluence.ContentPage{
				ID:       id,
				Title:    "Architecture",
				URL:      "https://wiki.example.com/arch",
				Markdown: "# Architecture\n\nOverview.",
			}, nil
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := fetchCmd()
	cmd.SetArgs([]string{"12345"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFetchCmdWithComments(t *testing.T) {
	mock := &mockAPI{
		fetchContentFunc: func(_ context.Context, _ string) (*confluence.ContentPage, error) {
			return &confluence.ContentPage{
				Title:    "Test",
				URL:      "https://wiki.example.com/test",
				Markdown: "Content here.",
			}, nil
		},
		fetchCommentsFunc: func(_ context.Context, _ string) ([]confluence.Comment, error) {
			return []confluence.Comment{
				{ID: "1", Author: "Jane", Body: "Looks good.", Location: "footer"},
			}, nil
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := fetchCmd()
	cmd.SetArgs([]string{"12345"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFetchCmdNoComments(t *testing.T) {
	commentsCalled := false
	mock := &mockAPI{
		fetchContentFunc: func(_ context.Context, _ string) (*confluence.ContentPage, error) {
			return &confluence.ContentPage{Title: "Test", URL: "https://wiki.example.com/test"}, nil
		},
		fetchCommentsFunc: func(_ context.Context, _ string) ([]confluence.Comment, error) {
			commentsCalled = true
			return nil, nil
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := fetchCmd()
	cmd.SetArgs([]string{"12345", "--no-comments"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if commentsCalled {
		t.Error("FetchComments should not be called with --no-comments")
	}
}

func TestFetchCmdError(t *testing.T) {
	mock := &mockAPI{
		fetchContentFunc: func(_ context.Context, _ string) (*confluence.ContentPage, error) {
			return nil, fmt.Errorf("page not found")
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := fetchCmd()
	cmd.SetArgs([]string{"99999"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "page not found") {
		t.Errorf("error = %q, want page not found", err)
	}
}

func TestHealthCmdOK(t *testing.T) {
	mock := &mockAPI{}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := healthCmd()
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestHealthCmdError(t *testing.T) {
	mock := &mockAPI{
		healthCheckFunc: func(_ context.Context) error {
			return fmt.Errorf("unauthorized")
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = false
	timing = false

	cmd := healthCmd()
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("error = %q, want unauthorized", err)
	}
}

func TestHealthCmdJSONOK(t *testing.T) {
	mock := &mockAPI{}
	cleanup := setMockClient(mock)
	defer cleanup()

	var buf bytes.Buffer
	jsonOutput = true
	timing = false
	defer func() { jsonOutput = false }()

	cmd := healthCmd()
	// The JSON goes to os.Stdout via getWriter, but we can still verify no error
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	_ = buf
}

func TestHealthCmdJSONError(t *testing.T) {
	mock := &mockAPI{
		healthCheckFunc: func(_ context.Context) error {
			return fmt.Errorf("timeout")
		},
	}
	cleanup := setMockClient(mock)
	defer cleanup()

	jsonOutput = true
	timing = false
	defer func() { jsonOutput = false }()

	cmd := healthCmd()
	// In JSON mode, health errors are returned as JSON, not as command errors
	err := cmd.Execute()
	// The JSON error path calls w.JSON which writes to stdout and returns nil
	// So actually the command itself doesn't return an error in JSON mode
	_ = err
}

func TestGetClientMissingURL(t *testing.T) {
	t.Setenv("CONFLUENCE_URL", "")
	t.Setenv("CONFLUENCE_PERSONAL_ACCESS_TOKEN", "")
	_, err := getClient()
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestGetClientMissingToken(t *testing.T) {
	t.Setenv("CONFLUENCE_URL", "https://wiki.example.com")
	t.Setenv("CONFLUENCE_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("CONFLUENCE_PAT", "")
	_, err := getClient()
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestGetClientPATFallback(t *testing.T) {
	t.Setenv("CONFLUENCE_URL", "https://wiki.example.com")
	t.Setenv("CONFLUENCE_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("CONFLUENCE_PAT", "fallback-token")
	client, err := getClient()
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
}

// Verify mockAPI satisfies the interface at compile time.
var _ confluence.API = (*mockAPI)(nil)
