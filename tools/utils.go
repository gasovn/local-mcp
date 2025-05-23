package tools

import (
	"github.com/strowk/foxy-contexts/pkg/mcp"
)

// errorResult creates an error result for MCP tools.
func errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: ptr(true),
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: message,
			},
		},
	}
}

// successResult creates a success result for MCP tools.
func successResult(content string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: ptr(false),
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}
