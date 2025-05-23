.PHONY: build test clean install run help fmt vet check dev

# Variables
BINARY_NAME=local-mcp
BUILD_DIR=.
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the MCP server binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test ./...

test: build ## Test the MCP server using inspector
	@echo "Starting MCP inspector for testing..."
	@echo "Open http://localhost:5173 in your browser to test the server"
	npx @modelcontextprotocol/inspector ./$(BINARY_NAME)

run: build ## Run the MCP server directly
	@echo "Starting $(BINARY_NAME) server..."
	./$(BINARY_NAME)

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean

install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	cp $(BINARY_NAME) $(GOPATH)/bin/

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

vet: ## Run go vet on the codebase
	@echo "Running go vet..."
	go vet ./...

check: fmt vet test-unit ## Run all code quality checks and tests

dev: clean build test-unit ## Development mode - build and test

all: clean deps check build ## Run full build pipeline