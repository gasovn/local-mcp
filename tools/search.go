package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
)

const (
	defaultSearchLimit = 10
	maxSearchLimit     = 20
	requestTimeout     = 30 * time.Second
	userAgent          = "local-mcp/1.0"
	titleMaxLength     = 60
)

// SearchResult represents a single search result.
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// SearchResponse contains the complete search response.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Query   string         `json:"query"`
	Total   int            `json:"total"`
}

// DuckDuckGoResponse represents the API response from DuckDuckGo.
type duckDuckGoResponse struct {
	Abstract       string `json:"Abstract"`
	AbstractText   string `json:"AbstractText"`
	AbstractSource string `json:"AbstractSource"`
	AbstractURL    string `json:"AbstractURL"`
	RelatedTopics  []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
	Results []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"Results"`
}

// NewSearchTool creates a new web search tool using DuckDuckGo.
func NewSearchTool() fxctx.Tool {
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "search-web",
			Description: ptr("Search the web using DuckDuckGo. Returns a list of search results with titles, URLs, and descriptions."),
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"query": {
						"type":        "string",
						"description": "The search query to execute",
					},
					"limit": {
						"type":        "integer",
						"description": "Maximum number of results to return (default: 10, max: 20)",
						"minimum":     1,
						"maximum":     maxSearchLimit,
						"default":     defaultSearchLimit,
					},
				},
				Required: []string{"query"},
			},
		},
		searchHandler,
	)
}

func searchHandler(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return errorResult("Query parameter is required and must be a non-empty string")
	}

	limit := parseLimit(args["limit"])

	results, err := performSearch(ctx, query, limit)
	if err != nil {
		return errorResult(fmt.Sprintf("Search failed: %v", err))
	}

	if len(results.Results) == 0 {
		return successResult(fmt.Sprintf("No results found for query: %s", query))
	}

	return formatSearchResults(results)
}

func parseLimit(limitArg interface{}) int {
	limit := defaultSearchLimit
	if l, ok := limitArg.(float64); ok {
		limit = int(l)
		if limit < 1 {
			limit = 1
		}
		if limit > maxSearchLimit {
			limit = maxSearchLimit
		}
	}
	return limit
}

func performSearch(ctx context.Context, query string, limit int) (*SearchResponse, error) {
	client := &http.Client{Timeout: requestTimeout}

	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ddgResponse duckDuckGoResponse
	if err := json.Unmarshal(body, &ddgResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	results := extractSearchResults(&ddgResponse, limit)
	if len(results) == 0 {
		return createFallbackResponse(query, searchURL), nil
	}

	return &SearchResponse{
		Results: results,
		Query:   query,
		Total:   len(results),
	}, nil
}

func extractSearchResults(ddgResponse *duckDuckGoResponse, limit int) []SearchResult {
	var results []SearchResult

	// Add abstract result if available
	if ddgResponse.AbstractText != "" && ddgResponse.AbstractURL != "" {
		results = append(results, SearchResult{
			Title:       ddgResponse.AbstractSource,
			URL:         ddgResponse.AbstractURL,
			Description: ddgResponse.AbstractText,
		})
	}

	// Add direct results
	for _, result := range ddgResponse.Results {
		if len(results) >= limit || result.Text == "" || result.FirstURL == "" {
			break
		}
		results = append(results, SearchResult{
			Title:       extractTitle(result.Text),
			URL:         result.FirstURL,
			Description: result.Text,
		})
	}

	// Add related topics if we need more results
	for _, topic := range ddgResponse.RelatedTopics {
		if len(results) >= limit || topic.Text == "" || topic.FirstURL == "" {
			break
		}
		results = append(results, SearchResult{
			Title:       extractTitle(topic.Text),
			URL:         topic.FirstURL,
			Description: topic.Text,
		})
	}

	return results
}

func createFallbackResponse(query, searchURL string) *SearchResponse {
	return &SearchResponse{
		Results: []SearchResult{
			{
				Title:       "DuckDuckGo Search",
				URL:         searchURL,
				Description: fmt.Sprintf("No instant answers found for '%s'. Please visit DuckDuckGo directly for web search results.", query),
			},
		},
		Query: query,
		Total: 1,
	}
}

func extractTitle(text string) string {
	text = strings.TrimSpace(text)

	// Extract title from text - take first part before dash or first sentence
	if idx := strings.Index(text, " - "); idx > 0 {
		return strings.TrimSpace(text[:idx])
	}

	if idx := strings.Index(text, ". "); idx > 0 && idx < 100 {
		return strings.TrimSpace(text[:idx])
	}

	if len(text) > titleMaxLength {
		return strings.TrimSpace(text[:titleMaxLength]) + "..."
	}

	return text
}

func formatSearchResults(results *SearchResponse) *mcp.CallToolResult {
	var content []interface{}

	// Add summary
	content = append(content, mcp.TextContent{
		Type: "text",
		Text: fmt.Sprintf("Search results for '%s' (%d results):\n", results.Query, len(results.Results)),
	})

	// Add each result
	for i, result := range results.Results {
		resultText := fmt.Sprintf("%d. **%s**\n   URL: %s\n   %s\n",
			i+1, result.Title, result.URL, result.Description)

		content = append(content, mcp.TextContent{
			Type: "text",
			Text: resultText,
		})
	}

	return &mcp.CallToolResult{
		IsError: ptr(false),
		Content: content,
	}
}
