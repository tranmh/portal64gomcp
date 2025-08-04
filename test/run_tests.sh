#!/bin/bash

# Portal64 MCP Server Test Runner
# This script runs the complete test suite following the test strategy

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=85
TEST_TIMEOUT="10m"
VERBOSE=${VERBOSE:-false}

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
    
    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
    
    log_success "Prerequisites check completed"
}

# Run code formatting and linting
run_code_quality_checks() {
    log_info "Running code quality checks..."
    
    # Format code
    log_info "Formatting code..."
    go fmt ./...
    
    # Vet code
    log_info "Vetting code..."
    go vet ./...
    
    # Run linter if available
    if command -v golangci-lint &> /dev/null; then
        log_info "Running golangci-lint..."
        golangci-lint run --timeout=5m
    else
        log_warning "Skipping golangci-lint (not installed)"
    fi
    
    log_success "Code quality checks completed"
}

# Download dependencies
setup_dependencies() {
    log_info "Setting up dependencies..."
    go mod download
    go mod tidy
    log_success "Dependencies setup completed"
}

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."
    
    local test_flags="-v -race -short -timeout=${TEST_TIMEOUT}"
    if [ "$VERBOSE" = "true" ]; then
        test_flags="$test_flags -v"
    fi
    
    if go test $test_flags ./internal/... ./pkg/...; then
        log_success "Unit tests passed"
        return 0
    else
        log_error "Unit tests failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    
    local test_flags="-v -race -timeout=${TEST_TIMEOUT}"
    if [ "$VERBOSE" = "true" ]; then
        test_flags="$test_flags -v"
    fi
    
    if go test $test_flags ./test/integration/...; then
        log_success "Integration tests passed"
        return 0
    else
        log_error "Integration tests failed"
        return 1
    fi
}

# Run all tests with coverage
run_tests_with_coverage() {
    log_info "Running all tests with coverage..."
    
    local test_flags="-v -race -timeout=${TEST_TIMEOUT} -coverprofile=coverage.out -covermode=atomic"
    if [ "$VERBOSE" = "true" ]; then
        test_flags="$test_flags -v"
    fi
    
    if go test $test_flags ./...; then
        log_success "All tests passed"
        
        # Generate coverage report
        log_info "Generating coverage report..."
        go tool cover -html=coverage.out -o coverage.html
        
        # Check coverage threshold
        check_coverage_threshold
        
        return 0
    else
        log_error "Some tests failed"
        return 1
    fi
}

# Check coverage threshold
check_coverage_threshold() {
    log_info "Checking coverage threshold..."
    
    local coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    
    if (( $(echo "$coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
        log_success "Coverage $coverage% meets threshold of $COVERAGE_THRESHOLD%"
    else
        log_error "Coverage $coverage% is below threshold of $COVERAGE_THRESHOLD%"
        return 1
    fi
}

# Run benchmarks
run_benchmarks() {
    log_info "Running benchmarks..."
    
    if go test -bench=. -benchmem -timeout=${TEST_TIMEOUT} ./...; then
        log_success "Benchmarks completed"
        return 0
    else
        log_error "Benchmarks failed"
        return 1
    fi
}

# Run race detection tests
run_race_tests() {
    log_info "Running race detection tests..."
    
    if go test -race -timeout=${TEST_TIMEOUT} ./...; then
        log_success "Race detection tests passed"
        return 0
    else
        log_error "Race condition detected"
        return 1
    fi
}

# Build the application
build_application() {
    log_info "Building application..."
    
    if make build; then
        log_success "Application built successfully"
        return 0
    else
        log_error "Application build failed"
        return 1
    fi
}

# Run smoke tests on built binary
run_smoke_tests() {
    log_info "Running smoke tests..."
    
    if [ ! -f "bin/portal64-mcp" ]; then
        log_error "Binary not found. Run build first."
        return 1
    fi
    
    # Test help command
    if ./bin/portal64-mcp -h > /dev/null 2>&1; then
        log_success "Binary help command works"
    else
        log_error "Binary help command failed"
        return 1
    fi
    
    log_success "Smoke tests passed"
    return 0
}

# Clean up test artifacts
cleanup() {
    log_info "Cleaning up test artifacts..."
    rm -f coverage.out coverage.html
    go clean -testcache
    log_success "Cleanup completed"
}

# Show test results summary
show_summary() {
    log_info "Test Summary:"
    echo "=============="
    
    if [ -f coverage.html ]; then
        log_info "Coverage report: coverage.html"
    fi
    
    if [ -f coverage.out ]; then
        local coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        log_info "Total coverage: $coverage"
    fi
    
    log_success "All tests completed successfully!"
}

# Main execution
main() {
    local test_type=${1:-"all"}
    local exit_code=0
    
    log_info "Starting Portal64 MCP Server test suite..."
    log_info "Test type: $test_type"
    
    case $test_type in
        "quick")
            check_prerequisites
            setup_dependencies
            run_code_quality_checks || exit_code=1
            run_unit_tests || exit_code=1
            ;;
        "unit")
            check_prerequisites
            setup_dependencies
            run_unit_tests || exit_code=1
            ;;
        "integration")
            check_prerequisites
            setup_dependencies
            run_integration_tests || exit_code=1
            ;;
        "coverage")
            check_prerequisites
            setup_dependencies
            run_code_quality_checks || exit_code=1
            run_tests_with_coverage || exit_code=1
            ;;
        "bench")
            check_prerequisites
            setup_dependencies
            run_benchmarks || exit_code=1
            ;;
        "race")
            check_prerequisites
            setup_dependencies
            run_race_tests || exit_code=1
            ;;
        "build")
            check_prerequisites
            setup_dependencies
            run_code_quality_checks || exit_code=1
            build_application || exit_code=1
            run_smoke_tests || exit_code=1
            ;;
        "ci")
            # Full CI pipeline
            check_prerequisites
            setup_dependencies
            run_code_quality_checks || exit_code=1
            run_tests_with_coverage || exit_code=1
            build_application || exit_code=1
            run_smoke_tests || exit_code=1
            ;;
        "all"|*)
            # Complete test suite
            check_prerequisites
            setup_dependencies
            run_code_quality_checks || exit_code=1
            run_tests_with_coverage || exit_code=1
            run_benchmarks || exit_code=1
            build_application || exit_code=1
            run_smoke_tests || exit_code=1
            ;;
    esac
    
    if [ $exit_code -eq 0 ]; then
        show_summary
        log_success "Test suite completed successfully!"
    else
        log_error "Test suite failed with exit code $exit_code"
    fi
    
    return $exit_code
}

# Handle script arguments
usage() {
    echo "Usage: $0 [test_type]"
    echo ""
    echo "Test types:"
    echo "  quick       - Run quick tests (code quality + unit tests)"
    echo "  unit        - Run unit tests only"
    echo "  integration - Run integration tests only"
    echo "  coverage    - Run tests with coverage report"
    echo "  bench       - Run benchmarks"
    echo "  race        - Run race detection tests"
    echo "  build       - Build and smoke test"
    echo "  ci          - Full CI pipeline"
    echo "  all         - Complete test suite (default)"
    echo ""
    echo "Environment variables:"
    echo "  VERBOSE=true  - Enable verbose output"
    echo "  COVERAGE_THRESHOLD=85 - Set coverage threshold percentage"
}

# Handle command line arguments
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    exit 0
fi

# Trap to ensure cleanup happens
trap cleanup EXIT

# Run main function
main "$@"
