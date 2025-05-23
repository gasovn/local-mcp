package tools

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

const (
	defaultCHHost     = "localhost"
	defaultCHPort     = 9000
	defaultCHDatabase = "default"
	defaultCHUsername = "default"
	defaultCHPassword = ""
	defaultCHLimit    = 100
	maxCHLimit        = 1000
	chTimeout         = 30 * time.Second
	maxConnections    = 5
	connLifetime      = 10 * time.Minute
)

// ClickHouseConfig holds the connection configuration for ClickHouse.
type ClickHouseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Secure   bool   `json:"secure"`
}

func isQuerySafe(query string) bool {
	trimmedQuery := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(trimmedQuery, "SELECT") ||
		strings.HasPrefix(trimmedQuery, "SHOW") ||
		strings.HasPrefix(trimmedQuery, "DESCRIBE")
}

func parseClickHouseLimit(limitArg interface{}) int {
	limit := defaultCHLimit
	if l, ok := limitArg.(float64); ok {
		limit = int(l)
		if limit < 1 {
			limit = 1
		}
		if limit > maxCHLimit {
			limit = maxCHLimit
		}
	}
	return limit
}

func connectToClickHouse(ctx context.Context, config ClickHouseConfig) (driver.Conn, error) {
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "local-mcp", Version: "1.0.0"},
			},
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      chTimeout,
		MaxOpenConns:     maxConnections,
		MaxIdleConns:     maxConnections,
		ConnMaxLifetime:  connLifetime,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		BlockBufferSize:  10,
	}

	if config.Secure {
		options.TLS = &tls.Config{}
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return conn, nil
}

func executeQuery(ctx context.Context, conn driver.Conn, query string, limit int) (string, error) {
	// Add LIMIT clause if not present in SELECT queries
	if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(query)), "SELECT") &&
		!strings.Contains(strings.ToUpper(query), "LIMIT") {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	return formatQueryResults(rows, limit)
}

func formatQueryResults(rows driver.Rows, limit int) (string, error) {
	columnTypes := rows.ColumnTypes()
	columnNames := make([]string, len(columnTypes))
	for i, col := range columnTypes {
		columnNames[i] = col.Name()
	}

	var result strings.Builder
	result.WriteString("Query Results:\n\n")

	// Write header
	result.WriteString(strings.Join(columnNames, " | "))
	result.WriteString("\n")
	result.WriteString(strings.Repeat("-", len(strings.Join(columnNames, " | "))))
	result.WriteString("\n")

	rowCount := 0
	for rows.Next() {
		if rowCount >= limit {
			break
		}

		values := createValueSlice(columnTypes)
		if err := rows.Scan(values...); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		stringValues := convertValuesToStrings(values)
		result.WriteString(strings.Join(stringValues, " | "))
		result.WriteString("\n")
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	if rowCount == 0 {
		result.WriteString("No rows returned.\n")
	} else {
		result.WriteString(fmt.Sprintf("\nTotal rows: %d", rowCount))
		if rowCount >= limit {
			result.WriteString(fmt.Sprintf(" (limited to %d)", limit))
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

func createValueSlice(columnTypes []driver.ColumnType) []interface{} {
	values := make([]interface{}, len(columnTypes))
	for i, colType := range columnTypes {
		switch colType.DatabaseTypeName() {
		case "UInt8":
			values[i] = new(uint8)
		case "UInt16":
			values[i] = new(uint16)
		case "UInt32":
			values[i] = new(uint32)
		case "UInt64":
			values[i] = new(uint64)
		case "Int8":
			values[i] = new(int8)
		case "Int16":
			values[i] = new(int16)
		case "Int32":
			values[i] = new(int32)
		case "Int64":
			values[i] = new(int64)
		case "Float32":
			values[i] = new(float32)
		case "Float64":
			values[i] = new(float64)
		case "String", "FixedString":
			values[i] = new(string)
		case "Date", "DateTime", "DateTime64":
			values[i] = new(time.Time)
		default:
			values[i] = new(string)
		}
	}
	return values
}

func convertValuesToStrings(values []interface{}) []string {
	stringValues := make([]string, len(values))
	for i, val := range values {
		if val == nil {
			stringValues[i] = "NULL"
			continue
		}

		switch v := val.(type) {
		case *uint8:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *uint16:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *uint32:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *uint64:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *int8:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *int16:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *int32:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *int64:
			stringValues[i] = fmt.Sprintf("%d", *v)
		case *float32:
			stringValues[i] = fmt.Sprintf("%g", *v)
		case *float64:
			stringValues[i] = fmt.Sprintf("%g", *v)
		case *string:
			stringValues[i] = *v
		case *time.Time:
			stringValues[i] = v.Format("2006-01-02 15:04:05")
		default:
			stringValues[i] = fmt.Sprintf("%v", val)
		}
	}
	return stringValues
}
