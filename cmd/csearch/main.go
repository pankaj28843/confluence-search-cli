package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pankaj28843/confluence-search-cli/internal/confluence"
	"github.com/pankaj28843/confluence-search-cli/internal/cql"
	"github.com/pankaj28843/confluence-search-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	timing     bool
	version    = "dev"
)

func main() {
	root := &cobra.Command{
		Use:   "csearch",
		Short: "Fast Confluence search from the command line",
		Long: `csearch - Confluence Search CLI

Search Confluence pages and fetch content directly from the terminal.
Translates natural language queries to CQL and calls the Confluence REST API.

Requires environment variables:
  CONFLUENCE_URL                 Base URL (e.g., https://wiki.example.com)
  CONFLUENCE_PERSONAL_ACCESS_TOKEN  Personal access token for auth

Workflow:
  csearch search "deployment process"           Search pages
  csearch search "API docs" --spaces ENG,OPS    Filter by space
  csearch fetch 12345                           Fetch page content
  csearch health                                Check API connectivity`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON (machine-readable)")
	root.PersistentFlags().BoolVar(&timing, "timing", false, "Show execution time on stderr")
	root.Version = version

	root.AddCommand(searchCmd())
	root.AddCommand(fetchCmd())
	root.AddCommand(healthCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func getClient() (*confluence.Client, error) {
	baseURL := os.Getenv("CONFLUENCE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("CONFLUENCE_URL environment variable is required")
	}
	token := os.Getenv("CONFLUENCE_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		token = os.Getenv("CONFLUENCE_PAT")
	}
	if token == "" {
		return nil, fmt.Errorf("CONFLUENCE_PERSONAL_ACCESS_TOKEN environment variable is required")
	}
	return confluence.NewClient(baseURL, token), nil
}

func getWriter() *output.Writer {
	return output.New(jsonOutput, timing)
}

func searchCmd() *cobra.Command {
	var (
		limit         int
		spaces        []string
		labels        []string
		titlesOnly    bool
		modifiedAfter string
		createdAfter  string
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Confluence pages",
		Long: `Search Confluence pages using natural language queries translated to CQL.

By default returns pages modified within the last 2 years. Archived and
trashed pages are always excluded.

Examples:
  csearch search "deployment process"
  csearch search "API documentation" --spaces ENG,OPS
  csearch search "release notes" --labels release,changelog
  csearch search "onboarding" --titles-only
  csearch search "architecture" --modified-after 30d
  csearch search "design doc" --dry-run    # Show CQL without executing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := getWriter()
			defer w.Finish()

			query := args[0]
			opts := cql.TranslateOptions{
				Spaces:        spaces,
				Labels:        labels,
				TitlesOnly:    titlesOnly,
				ModifiedAfter: modifiedAfter,
				CreatedAfter:  createdAfter,
			}

			cqlStr, err := cql.Translate(query, opts)
			if err != nil {
				return err
			}

			if dryRun {
				if w.Format == output.FormatJSON {
					return w.JSON(map[string]string{"cql": cqlStr})
				}
				w.Text("CQL: %s\n", cqlStr)
				return nil
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			resp, err := client.Search(cqlStr, limit)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if w.Format == output.FormatJSON {
				return w.JSON(resp)
			}

			if len(resp.Results) == 0 {
				w.Text("No results found for %q\n", query)
				return nil
			}

			for i, r := range resp.Results {
				if i > 0 {
					w.Text("\n")
				}
				w.Text("%s\n", r.Title)
				w.Text("  %s\n", r.URL)
				if r.SpaceKey != "" {
					w.Text("  Space: %s\n", r.SpaceKey)
				}
				if r.Excerpt != "" {
					for _, line := range output.WordWrap(r.Excerpt, 76) {
						w.Text("  %s\n", line)
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 5, "Maximum number of results (1-25)")
	cmd.Flags().StringSliceVar(&spaces, "spaces", nil, "Filter by space keys (comma-separated)")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "Filter by labels (comma-separated)")
	cmd.Flags().BoolVar(&titlesOnly, "titles-only", false, "Search titles only")
	cmd.Flags().StringVar(&modifiedAfter, "modified-after", "2y", "Time filter: 1d, 7d, 30d, 90d, 6M, 1y, 2y, 5y, or ISO date")
	cmd.Flags().StringVar(&createdAfter, "created-after", "", "Created after filter (same format as modified-after)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show CQL query without executing")
	return cmd
}

func fetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch <content-id>",
		Short: "Fetch a Confluence page by content ID",
		Long: `Fetch the full content of a Confluence page by its numeric content ID.

The content ID is returned in search results. The page body is converted
from Confluence storage HTML to readable markdown.

Examples:
  csearch fetch 12345
  csearch fetch 12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := getWriter()
			defer w.Finish()

			client, err := getClient()
			if err != nil {
				return err
			}

			page, err := client.FetchContent(args[0])
			if err != nil {
				return fmt.Errorf("fetch failed: %w", err)
			}

			if w.Format == output.FormatJSON {
				return w.JSON(page)
			}

			// Render as markdown document
			w.Text("# %s\n\n", page.Title)
			w.Text("**URL:** %s\n", page.URL)
			if page.SpaceKey != "" {
				w.Text("**Space:** %s\n", page.SpaceKey)
			}
			if page.Version > 0 {
				w.Text("**Version:** %d\n", page.Version)
			}
			if page.LastModified != "" {
				w.Text("**Last modified:** %s\n", page.LastModified)
			}
			if len(page.Labels) > 0 {
				w.Text("**Labels:** %s\n", strings.Join(page.Labels, ", "))
			}
			if len(page.Ancestors) > 0 {
				w.Text("**Ancestors:** %s\n", strings.Join(page.Ancestors, " > "))
			}
			w.Text("\n---\n\n")
			if page.Markdown != "" {
				w.Text("%s\n", page.Markdown)
			}
			return nil
		},
	}
}

func healthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check Confluence API connectivity",
		Long: `Verify that the Confluence REST API is reachable and the
authentication token is valid.

Examples:
  csearch health
  csearch health --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := getWriter()
			defer w.Finish()

			client, err := getClient()
			if err != nil {
				return err
			}

			err = client.HealthCheck()
			if err != nil {
				if w.Format == output.FormatJSON {
					return w.JSON(map[string]string{"status": "error", "error": err.Error()})
				}
				return fmt.Errorf("health check failed: %w", err)
			}

			if w.Format == output.FormatJSON {
				return w.JSON(map[string]string{"status": "ok"})
			}
			w.Text("Confluence API: OK\n")
			return nil
		},
	}
}
