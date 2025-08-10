# Makefile for Portal64 MCP Server with SSL support

.PHONY: help build run test clean ssl-certs ssl-clean dev prod docker-build

# Default target
help:
	@echo "Portal64 MCP Server - Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo ""
	@echo "SSL targets:"
	@echo "  ssl-certs   - Generate development SSL certificates"
	@echo "  ssl-clean   - Remove generated SSL certificates"
	@echo "  ssl-info    - Show SSL certificate information"
	@echo ""
	@echo "Run targets:"
	@echo "  dev         - Run in development mode (SSL disabled)"
	@echo "  prod        - Run in production mode (SSL enabled)"
	@echo "  run         - Run with default configuration"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build - Build Docker image"

# Build configuration
BINARY_NAME := portal64-mcp
BUILD_DIR := bin
CERT_DIR := certs
GO_FILES := $(shell find . -name '*.go' -type f)

# SSL configuration
SSL_COUNTRY := US
SSL_STATE := ""
SSL_CITY := ""
SSL_ORG := "Portal64 MCP Server"
SSL_HOSTS := localhost,127.0.0.1,::1

# Build the binary
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Generate development SSL certificates
ssl-certs: $(CERT_DIR)/server.crt $(CERT_DIR)/server.key

$(CERT_DIR)/server.crt $(CERT_DIR)/server.key:
	@echo "Generating SSL certificates for development..."
	@mkdir -p $(CERT_DIR)
	@openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout $(CERT_DIR)/server.key \
		-out $(CERT_DIR)/server.crt \
		-config <(printf "[req]\n" \
			"distinguished_name = req_distinguished_name\n" \
			"req_extensions = v3_req\n" \
			"prompt = no\n" \
			"[req_distinguished_name]\n" \
			"C = $(SSL_COUNTRY)\n" \
			"ST = $(SSL_STATE)\n" \
			"L = $(SSL_CITY)\n" \
			"O = $(SSL_ORG)\n" \
			"CN = localhost\n" \
			"[v3_req]\n" \
			"keyUsage = nonRepudiation, digitalSignature, keyEncipherment\n" \
			"subjectAltName = @alt_names\n" \
			"[alt_names]\n" \
			"DNS.1 = localhost\n" \
			"IP.1 = 127.0.0.1\n" \
			"IP.2 = ::1\n")
	@chmod 600 $(CERT_DIR)/server.key
	@chmod 644 $(CERT_DIR)/server.crt
	@echo "SSL certificates generated:"
	@echo "  Certificate: $(CERT_DIR)/server.crt"
	@echo "  Private Key: $(CERT_DIR)/server.key"

# Remove SSL certificates
ssl-clean:
	@echo "Removing SSL certificates..."
	rm -rf $(CERT_DIR)

# Show SSL certificate information
ssl-info:
	@if [ -f "$(CERT_DIR)/server.crt" ]; then \
		echo "SSL Certificate Information:"; \
		openssl x509 -in $(CERT_DIR)/server.crt -text -noout | grep -A 5 -B 5 "Subject:\|Validity\|DNS:\|IP Address:"; \
	else \
		echo "No SSL certificate found. Run 'make ssl-certs' to generate one."; \
	fi

# Development mode (SSL disabled)
dev: build
	@echo "Starting Portal64 MCP Server in development mode (SSL disabled)..."
	ENV=development $(BUILD_DIR)/$(BINARY_NAME) -log-level debug

# Production mode (SSL enabled)
prod: build ssl-certs
	@echo "Starting Portal64 MCP Server in production mode (SSL enabled)..."
	ENV=production $(BUILD_DIR)/$(BINARY_NAME)

# Run with default configuration
run: build
	@echo "Starting Portal64 MCP Server..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Development with SSL enabled (for testing)
dev-ssl: build ssl-certs
	@echo "Starting Portal64 MCP Server in development mode with SSL enabled..."
	ENV=development MCP_SSL_ENABLED=true $(BUILD_DIR)/$(BINARY_NAME) -log-level debug

# Test SSL connection
test-ssl:
	@echo "Testing SSL connection..."
	@if curl -k -s https://localhost:8888/health > /dev/null; then \
		echo "✓ SSL connection successful"; \
		curl -k -s https://localhost:8888/api/v1/ssl/info | jq .; \
	else \
		echo "✗ SSL connection failed"; \
		exit 1; \
	fi

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

# Full development setup
setup: deps ssl-certs
	@echo "Development setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  make dev      - Run in development mode"
	@echo "  make dev-ssl  - Run in development mode with SSL"
	@echo "  make prod     - Run in production mode"