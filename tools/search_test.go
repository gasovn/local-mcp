package tools

import (
	"context"
	"testing"
)

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Title with dash separator",
			input:    "Go Programming Best Practices - A Complete Guide",
			expected: "Go Programming Best Practices",
		},
		{
			name:     "Title with period separator",
			input:    "Effective Go programming. Learn the fundamentals of Go development.",
			expected: "Effective Go programming",
		},
		{
			name:     "Long title truncation",
			input:    "This is a very long title that should be truncated because it exceeds the maximum length limit",
			expected: "This is a very long title that should be truncated because i...",
		},
		{
			name:     "Short title",
			input:    "Go Tutorial",
			expected: "Go Tutorial",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Title with whitespace",
			input:    "  Go Programming Guide  ",
			expected: "Go Programming Guide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTitle(tt.input)
			if result != tt.expected {
				t.Errorf("extractTitle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"default nil", nil, defaultSearchLimit},
		{"valid limit", float64(5), 5},
		{"too small", float64(0), 1},
		{"too large", float64(25), maxSearchLimit},
		{"negative", float64(-5), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLimit(tt.input)
			if result != tt.expected {
				t.Errorf("parseLimit(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSearchResultValidation(t *testing.T) {
	result := SearchResult{
		Title:       "Test Title",
		URL:         "https://example.com",
		Description: "Test description",
	}

	if result.Title == "" {
		t.Error("SearchResult should have a title")
	}
	if result.URL == "" {
		t.Error("SearchResult should have a URL")
	}
	if result.Description == "" {
		t.Error("SearchResult should have a description")
	}
}

func TestSearchResponseValidation(t *testing.T) {
	response := SearchResponse{
		Results: []SearchResult{
			{
				Title:       "Test Result",
				URL:         "https://example.com",
				Description: "Test description",
			},
		},
		Query: "test query",
		Total: 1,
	}

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}
	if response.Query != "test query" {
		t.Errorf("Expected query 'test query', got %q", response.Query)
	}
	if response.Total != 1 {
		t.Errorf("Expected total 1, got %d", response.Total)
	}
}

func TestSearchHandlerValidation(t *testing.T) {
	ctx := context.Background()

	// Test missing query parameter
	result := searchHandler(ctx, map[string]interface{}{})
	if result.IsError == nil || !*result.IsError {
		t.Error("Expected error for missing query parameter")
	}

	// Test empty query
	result = searchHandler(ctx, map[string]interface{}{
		"query": "",
	})
	if result.IsError == nil || !*result.IsError {
		t.Error("Expected error for empty query")
	}

	// Test invalid query type
	result = searchHandler(ctx, map[string]interface{}{
		"query": 123,
	})
	if result.IsError == nil || !*result.IsError {
		t.Error("Expected error for invalid query type")
	}

	// Test whitespace-only query
	result = searchHandler(ctx, map[string]interface{}{
		"query": "   ",
	})
	if result.IsError == nil || !*result.IsError {
		t.Error("Expected error for whitespace-only query")
	}
}

func TestNewSearchTool(t *testing.T) {
	tool := NewSearchTool()

	if tool == nil {
		t.Fatal("NewSearchTool() returned nil")
	}

	// Test that the tool can be created without panicking
	// Additional validation would require accessing internal fields
	// which may not be available depending on the foxy-contexts implementation
}
