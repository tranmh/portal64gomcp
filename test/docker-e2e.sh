#!/bin/bash

# Portal64 MCP Server E2E Test Docker Helper
# Simplifies running e2e tests with Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.e2e.yml"
PROJECT_NAME="portal64-e2e"

# Functions
print_header() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check Docker and Docker Compose
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    
    print_success "Docker and Docker Compose are available"
}

# Build images
build_images() {
    print_info "Building Docker images..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME build --no-cache
    print_success "Images built successfully"
}

# Run full e2e test suite
run_full_tests() {
    print_header "Running Full E2E Test Suite"
    
    # Clean up any existing containers
    cleanup
    
    # Start server and run tests
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up --abort-on-container-exit portal64-server e2e-tests
    
    # Check test results
    if [ $? -eq 0 ]; then
        print_success "Full E2E test suite completed successfully"
    else
        print_error "Full E2E test suite failed"
        show_logs
        exit 1
    fi
}

# Run specific test category
run_category_tests() {
    local category=$1
    print_header "Running E2E Tests - Category: $category"
    
    # Clean up any existing containers
    cleanup
    
    # Set environment variable and run tests
    export TEST_CATEGORY=$category
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME --profile category-tests up --abort-on-container-exit portal64-server e2e-tests-category
    
    if [ $? -eq 0 ]; then
        print_success "E2E tests for category '$category' completed successfully"
    else
        print_error "E2E tests for category '$category' failed"
        show_logs
        exit 1
    fi
}

# Run performance tests
run_performance_tests() {
    print_header "Running Performance Tests and Benchmarks"
    
    # Clean up any existing containers
    cleanup
    
    # Run performance tests
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME --profile performance up --abort-on-container-exit portal64-server performance-tests
    
    if [ $? -eq 0 ]; then
        print_success "Performance tests completed successfully"
    else
        print_error "Performance tests failed"
        show_logs
        exit 1
    fi
}

# Run with mock server (offline testing)
run_mock_tests() {
    print_header "Running E2E Tests with Mock Server"
    
    # Clean up any existing containers
    cleanup
    
    # Run with mock server
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME --profile mock up --abort-on-container-exit mock-server
    
    print_info "Mock server is running on http://localhost:8081"
    print_info "You can now run tests against the mock server manually"
}

# Start results viewer
start_results_viewer() {
    print_header "Starting Test Results Viewer"
    
    # Ensure test results directory exists
    mkdir -p test-results
    
    # Start results viewer
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME --profile viewer up -d results-viewer
    
    print_success "Results viewer started at http://localhost:8082"
    print_info "Test results will be available at the above URL"
}

# Show logs
show_logs() {
    print_info "Showing container logs..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs --tail=50
}

# Clean up containers and volumes
cleanup() {
    print_info "Cleaning up existing containers..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v --remove-orphans 2>/dev/null || true
}

# Development mode
dev_mode() {
    print_header "Starting Development Mode"
    
    # Clean up
    cleanup
    
    # Start in development mode with live reloading
    docker-compose -f $COMPOSE_FILE -f docker-compose.dev.yml -p $PROJECT_NAME up --build
}

# Show status
show_status() {
    print_header "Docker Container Status"
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
}

# Show help
show_help() {
cat << EOF
Portal64 MCP Server E2E Test Docker Helper

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    build                   Build Docker images
    test                    Run full e2e test suite
    test-category CATEGORY  Run specific test category
    performance             Run performance tests and benchmarks
    mock                    Start mock server for offline testing
    viewer                  Start test results viewer
    dev                     Start development mode with live reloading
    logs                    Show container logs
    status                  Show container status
    cleanup                 Clean up containers and volumes
    help                    Show this help message

TEST CATEGORIES:
    administrative          Administrative tools tests
    search                 Search tools tests
    detail                 Detail tools tests
    analysis               Analysis tools tests
    protocol               MCP protocol tests
    error_scenarios        Error handling tests
    performance            Performance tests

EXAMPLES:
    $0 build                           # Build images
    $0 test                            # Run all tests
    $0 test-category search            # Run search tests only
    $0 performance                     # Run performance tests
    $0 mock                            # Start mock server
    $0 viewer                          # Start results viewer
    $0 dev                             # Development mode
    $0 cleanup                         # Clean up everything

ENVIRONMENT VARIABLES:
    TEST_CATEGORY          Test category to run (used with test-category)
    LOG_LEVEL             Log level (debug, info, warn, error)
    BASE_URL              Server base URL (default: http://portal64-server:8080)

For more information, see test/README.md
EOF
}

# Main execution
main() {
    # Check prerequisites
    check_docker
    
    # Parse command
    case "${1:-help}" in
        build)
            build_images
            ;;
        test|tests)
            run_full_tests
            ;;
        test-category)
            if [ -z "$2" ]; then
                print_error "Test category is required"
                echo "Available categories: administrative, search, detail, analysis, protocol, error_scenarios, performance"
                exit 1
            fi
            run_category_tests "$2"
            ;;
        performance|perf|benchmark)
            run_performance_tests
            ;;
        mock)
            run_mock_tests
            ;;
        viewer|results)
            start_results_viewer
            ;;
        dev|development)
            dev_mode
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        cleanup|clean)
            cleanup
            print_success "Cleanup completed"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"
