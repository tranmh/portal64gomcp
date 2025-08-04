.PHONY: build clean test run fmt vet deps help

# Variables
BINARY_NAME=portal64-mcp
BINARY_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f | grep -v vendor/)
MAIN_PATH=./cmd/server

# Default target
all: build

# Build the binary
build: deps fmt vet
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	go build -ldflags="-w -s" -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

# Build for production (with optimizations)
build-prod: deps fmt vet
	@echo "Building $(BINARY_NAME) for production..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 go build -ldflags="-w -s -X main.version=$(shell git describe --tags --always --dirty)" -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Production build complete: $(BINARY_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)
	go clean -cache
	@echo "Clean complete"

# Run tests
test:
	@echo "Running all tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -race -short ./internal/... ./pkg/...
	@echo "Unit tests complete"

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -race ./test/integration/...
	@echo "Integration tests complete"

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with coverage threshold check
test-coverage-threshold: test-coverage
	@echo "Checking coverage threshold..."
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' | awk '{if ($$1 < 85) {print "Coverage " $$1 "% is below threshold of 85%"; exit 1} else {print "Coverage " $$1 "% meets threshold"}}'

# Run benchmarks
test-bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...
	@echo "Benchmarks complete"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...
	@echo "Race detection tests complete"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_DIR)/$(BINARY_NAME)

# Run with debug logging
run-debug: build
	@echo "Running $(BINARY_NAME) with debug logging..."
	./$(BINARY_DIR)/$(BINARY_NAME) -log-level debug

# Run with config file
run-config: build
	@echo "Running $(BINARY_NAME) with config file..."
	./$(BINARY_DIR)/$(BINARY_NAME) -config config.yaml

# Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet Go code
vet:
	@echo "Vetting code..."
	go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Generate Go modules graph
deps-graph:
	@echo "Generating dependency graph..."
	go mod graph | sed -E 's/@[^ ]+//g' | sort | uniq > deps.txt
	@echo "Dependency graph saved to deps.txt"

# Cross-compile for different platforms
build-all: deps fmt vet
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BINARY_DIR)
	# Linux amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# Linux arm64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# Windows amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	# macOS amd64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	# macOS arm64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Cross-compilation complete"

# Create a sample config file
config-sample:
	@echo "Creating sample configuration file..."
	@cat > config.yaml << 'EOF'
api:
  base_url: "http://localhost:8080"
  timeout: "30s"

mcp:
  port: 3000

logging:
  level: "info"
  format: "json"
EOF
	@echo "Sample config created: config.yaml"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-prod     - Build for production with optimizations"
	@echo "  build-all      - Cross-compile for multiple platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-coverage-threshold - Check coverage meets 85% threshold"
	@echo "  test-bench     - Run benchmarks"
	@echo "  test-race      - Run tests with race detection"
	@echo "  run            - Build and run the application"
	@echo "  run-debug      - Run with debug logging"
	@echo "  run-config     - Run with config file"
	@echo "  fmt            - Format Go code"
	@echo "  vet            - Vet Go code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  deps-update    - Update all dependencies"
	@echo "  deps-graph     - Generate dependency graph"
	@echo "  install-tools  - Install development tools"
	@echo "  config-sample  - Create sample configuration file"
	@echo "  help           - Show this help message"
