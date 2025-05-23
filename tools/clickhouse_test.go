package tools

import (
	"testing"
)

// Tests for environment-based ClickHouse configuration are handled separately
// since the new implementation uses environment variables instead of explicit parameters

func TestIsQuerySafe(t *testing.T) {
	unsafeQueries := []string{
		"DROP TABLE test",
		"DELETE FROM users",
		"INSERT INTO logs VALUES (1, 'test')",
		"UPDATE users SET name = 'hacker'",
		"CREATE TABLE malicious (id Int32)",
		"ALTER TABLE users ADD COLUMN evil String",
	}

	for _, query := range unsafeQueries {
		t.Run(query, func(t *testing.T) {
			if isQuerySafe(query) {
				t.Errorf("Expected query to be unsafe: %s", query)
			}
		})
	}
}

func TestIsQuerySafe_SafeQueries(t *testing.T) {
	safeQueries := []string{
		"SELECT 1",
		"select * from users limit 10",
		"SHOW TABLES",
		"show databases",
		"DESCRIBE users",
		"describe table_name",
	}

	for _, query := range safeQueries {
		t.Run(query, func(t *testing.T) {
			if !isQuerySafe(query) {
				t.Errorf("Expected query to be safe: %s", query)
			}
		})
	}
}

func TestParseClickHouseLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"default", nil, defaultCHLimit},
		{"valid limit", float64(50), 50},
		{"too small", float64(0), 1},
		{"too large", float64(2000), maxCHLimit},
		{"negative", float64(-10), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseClickHouseLimit(tt.input)
			if result != tt.expected {
				t.Errorf("parseClickHouseLimit(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
