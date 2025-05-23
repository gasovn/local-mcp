package tools

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
)

const (
	envCHHost     = "CLICKHOUSE_HOST"
	envCHPort     = "CLICKHOUSE_PORT"
	envCHDatabase = "CLICKHOUSE_DATABASE"
	envCHUsername = "CLICKHOUSE_USERNAME"
	envCHPassword = "CLICKHOUSE_PASSWORD"
	envCHSecure   = "CLICKHOUSE_SECURE"
)

// NewClickHouseQueryTool creates a ClickHouse query tool that uses environment variables.
func NewClickHouseQueryTool() fxctx.Tool {
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "clickhouse-query",
			Description: ptr("Execute SQL queries against ClickHouse database using connection parameters from environment variables (configured in Zed settings)"),
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"query": {
						"type":        "string",
						"description": "SQL query to execute against ClickHouse",
					},
					"limit": {
						"type":        "integer",
						"description": "Maximum number of rows to return (default: 100, max: 1000)",
						"minimum":     1,
						"maximum":     maxCHLimit,
						"default":     defaultCHLimit,
					},
				},
				Required: []string{"query"},
			},
		},
		clickHouseQueryHandler,
	)
}

// NewClickHouseSchemasTool creates a tool to list ClickHouse schemas using environment configuration.
func NewClickHouseSchemasTool() fxctx.Tool {
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "clickhouse-schemas",
			Description: ptr("List available databases in ClickHouse instance using environment configuration"),
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]map[string]interface{}{},
				Required:   []string{},
			},
		},
		clickHouseSchemasHandler,
	)
}

// NewClickHouseTablesTool creates a tool to list tables using environment configuration.
func NewClickHouseTablesTool() fxctx.Tool {
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "clickhouse-tables",
			Description: ptr("List tables in a ClickHouse database using environment configuration"),
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"database": {
						"type":        "string",
						"description": "Database name to list tables from (optional, uses env CLICKHOUSE_DATABASE if not specified)",
					},
				},
				Required: []string{},
			},
		},
		clickHouseTablesHandler,
	)
}

func clickHouseQueryHandler(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return errorResult("Query parameter is required and must be a non-empty string")
	}

	if !isQuerySafe(query) {
		return errorResult("Only SELECT, SHOW, and DESCRIBE queries are allowed for security reasons")
	}

	config := getClickHouseConfigFromEnv()
	if config == nil {
		return errorResult("ClickHouse configuration not found in environment variables. Please check your settings.")
	}

	limit := parseClickHouseLimit(args["limit"])

	conn, err := connectToClickHouse(ctx, *config)
	if err != nil {
		return errorResult("Failed to connect to ClickHouse: " + err.Error() + "\nPlease verify your connection settings.")
	}
	defer conn.Close()

	results, err := executeQuery(ctx, conn, query, limit)
	if err != nil {
		return errorResult("Query execution failed: " + err.Error())
	}

	return successResult(results)
}

func clickHouseSchemasHandler(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
	config := getClickHouseConfigFromEnv()
	if config == nil {
		return errorResult("ClickHouse configuration not found in environment variables. Please check your settings.")
	}

	conn, err := connectToClickHouse(ctx, *config)
	if err != nil {
		return errorResult("Failed to connect to ClickHouse: " + err.Error() + "\nPlease verify your connection settings.")
	}
	defer conn.Close()

	results, err := executeQuery(ctx, conn, "SHOW DATABASES", maxCHLimit)
	if err != nil {
		return errorResult("Failed to list databases: " + err.Error())
	}

	return successResult(results)
}

func clickHouseTablesHandler(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
	config := getClickHouseConfigFromEnv()
	if config == nil {
		return errorResult("ClickHouse configuration not found in environment variables. Please check your settings.")
	}

	database := config.Database
	if db, ok := args["database"].(string); ok && db != "" {
		database = db
	}

	conn, err := connectToClickHouse(ctx, *config)
	if err != nil {
		return errorResult("Failed to connect to ClickHouse: " + err.Error() + "\nPlease verify your connection settings.")
	}
	defer conn.Close()

	query := "SHOW TABLES FROM " + database
	results, err := executeQuery(ctx, conn, query, maxCHLimit)
	if err != nil {
		return errorResult("Failed to list tables from database '" + database + "': " + err.Error())
	}

	return successResult(results)
}

func getClickHouseConfigFromEnv() *ClickHouseConfig {
	host := os.Getenv(envCHHost)
	if host == "" {
		return nil
	}

	port := parseEnvInt(envCHPort, defaultCHPort)
	database := getEnvOrDefault(envCHDatabase, defaultCHDatabase)
	username := getEnvOrDefault(envCHUsername, defaultCHUsername)
	password := os.Getenv(envCHPassword)
	secure := parseEnvBool(envCHSecure)

	return &ClickHouseConfig{
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
		Secure:   secure,
	}
}

func parseEnvInt(envVar string, defaultValue int) int {
	str := os.Getenv(envVar)
	if str == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvOrDefault(envVar, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseEnvBool(envVar string) bool {
	str := os.Getenv(envVar)
	return strings.ToLower(str) == "true"
}
