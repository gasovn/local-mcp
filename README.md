# Local MCP Server

A Model Context Protocol (MCP) server providing web search and ClickHouse database tools.

## Features

- **Web Search**: Search using DuckDuckGo API
- **ClickHouse Integration**: Execute safe SQL queries against ClickHouse databases
- **Environment Configuration**: Configure connection through environment variables

## Installation

### Prerequisites

- Go 1.24+
- ClickHouse server (optional, for database features)

### Build

```bash
git clone <repository-url>
cd local-mcp
make build
```

## Configuration

Configure as MCP server in your editor settings:

```json
{
  "context_servers": {
    "local-mcp": {
      "command": "/path/to/local-mcp",
      "args": [],
      "env": {
        "CLICKHOUSE_HOST": "localhost",
        "CLICKHOUSE_PORT": "9000", 
        "CLICKHOUSE_DATABASE": "default",
        "CLICKHOUSE_USERNAME": "default",
        "CLICKHOUSE_PASSWORD": "",
        "CLICKHOUSE_SECURE": "false"
      }
    }
  }
}
```

## Available Tools

### search-web
Search the web using DuckDuckGo.

Parameters:
- `query` (required): Search query string
- `limit` (optional): Max results (1-20, default: 10)

### ClickHouse Tools

#### clickhouse-env-query
Execute SQL queries using environment configuration.

Parameters:
- `query` (required): SQL query (SELECT/SHOW/DESCRIBE only)
- `limit` (optional): Max rows (1-1000, default: 100)

#### clickhouse-env-schemas
List available databases.

#### clickhouse-env-tables
List tables in a database.

Parameters:
- `database` (optional): Database name (uses env default if not specified)

## Security

- Only read-only SQL operations allowed (SELECT, SHOW, DESCRIBE)
- Query results are limited to prevent resource exhaustion
- Connection parameters validated

## Development

```bash
# Run tests
make test-unit

# Format code
make fmt

# Run checks
make check

# Build and test
make dev
```

## License

MIT