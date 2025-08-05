#!/bin/bash

# E2E Test Runner for Portal64 MCP Server
# Implements the test execution strategy from docs/e2e-test-strategy.md

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8888"
TEST_RESULTS_DIR="test-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="${TEST_RESULTS_DIR}/e2e_test_results_${TIMESTAMP}.txt"

# Test categories
declare -a TEST_CATEGORIES=(
    "administrative"
    "search"
    "detail"
    "analysis"
    "protocol"
    "error_scenarios"
    "performance"
)

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

# Check if server is running
check_server() {
    print_info "Checking if Portal64 MCP Server is running on ${BASE_URL}..."
    
    if curl -s --connect-timeout 5 "${BASE_URL}/health" > /dev/null 2>&1; then
        print_success "Server is running and responsive"
        return 0
    else
        print_error "Server is not running or not responding at ${BASE_URL}"
        print_info "Please start the Portal64 MCP Server on localhost:8888 before running tests"
        return 1
    fi
}

# Setup test environment
setup_test_environment() {
    print_info "Setting up test environment..."
    
    # Create test results directory
    mkdir -p "${TEST_RESULTS_DIR}"
    
    # Initialize results file
    cat > "${RESULTS_FILE}" << EOF
Portal64 MCP Server E2E Test Results
====================================
Test Run: ${TIMESTAMP}
Base URL: ${BASE_URL}
Test Strategy: docs/e2e-test-strategy.md

EOF
    
    print_success "Test environment setup complete"
}

# Run specific test category
run_test_category() {
    local category=$1
    local description=$2
    
    print_header "Running ${description}"
    
    case $category in
        "administrative")
            run_go_test "TestPortal64MCP_E2E_AllTools/1.*Administrative.*Tools.*Tests"
            ;;
        "search")
            run_go_test "TestPortal64MCP_E2E_AllTools/2.*Search.*Tools.*Tests"
            ;;
        "detail")
            run_go_test "TestPortal64MCP_E2E_AllTools/3.*Detail.*Tools.*Tests"
            ;;
        "analysis")
            run_go_test "TestPortal64MCP_E2E_AllTools/4.*Analysis.*Tools.*Tests"
            ;;
        "protocol")
            run_go_test "TestPortal64MCP_E2E_AllTools/5.*MCP.*Protocol.*Tests"
            ;;
        "error_scenarios")
            run_go_test "TestPortal64MCP_E2E_ErrorScenarios"
            ;;
        "performance")
            run_go_test "TestPortal64MCP_E2E_Performance"
            ;;
    esac
}

# Run Go tests with proper formatting
run_go_test() {
    local test_pattern=$1
    local start_time=$(date +%s)
    
    echo "Running test pattern: $test_pattern" | tee -a "${RESULTS_FILE}"
    echo "----------------------------------------"
    
    if go test -v -timeout 300s ./test/integration -run "$test_pattern" 2>&1 | tee -a "${RESULTS_FILE}"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_success "Test category completed successfully in ${duration}s"
        echo "✓ Test category completed successfully in ${duration}s" >> "${RESULTS_FILE}"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_error "Test category failed after ${duration}s"
        echo "✗ Test category failed after ${duration}s" >> "${RESULTS_FILE}"
        return 1
    fi
}

# Run benchmarks
run_benchmarks() {
    print_header "Running Performance Benchmarks"
    
    echo "Performance Benchmarks" >> "${RESULTS_FILE}"
    echo "=====================" >> "${RESULTS_FILE}"
    
    go test -bench=. -benchmem ./test/integration 2>&1 | tee -a "${RESULTS_FILE}"
    
    if [ $? -eq 0 ]; then
        print_success "Benchmarks completed successfully"
    else
        print_warning "Some benchmarks may have failed or been skipped"
    fi
}

# Generate test report
generate_report() {
    print_header "Generating Test Report"
    
    local report_file="${TEST_RESULTS_DIR}/test_report_${TIMESTAMP}.md"
    
    cat > "${report_file}" << EOF
# Portal64 MCP Server E2E Test Report

**Test Run:** ${TIMESTAMP}  
**Base URL:** ${BASE_URL}  
**Test Strategy:** docs/e2e-test-strategy.md

## Test Execution Summary

The tests follow the execution order specified in the e2e test strategy:

1. **Administrative Tests First** - Verify API health and basic connectivity
2. **Search Tools** - Test all search functionalities  
3. **Detail Tools** - Test profile and detail retrieval
4. **Analysis Tools** - Test statistical and historical data
5. **MCP Protocol** - Test protocol compliance and resource access
6. **Error Scenarios** - Test error handling and edge cases
7. **Performance Tests** - Verify response times and server stability

## Test Data Used

As specified in the e2e test strategy, all tests use predefined data:

- **Players**: Query "Minh Cuong", Player ID "C0327-297"
- **Clubs**: Query "Altbach", Club ID "C0327"  
- **Tournaments**: Query "Ulm", Tournament ID "C350-C01-SMU"
- **Date Range**: 2023-2024 season

## Detailed Results

See full test output in: \`${RESULTS_FILE}\`

## Success Criteria

### Functional Success
- [ ] All 50+ test cases pass
- [ ] Non-empty results for all specified test data
- [ ] Proper error handling for invalid inputs
- [ ] Consistent response formats

### Performance Success
- [ ] Response times under 5 seconds for all calls
- [ ] Server remains stable under test load
- [ ] Memory usage stays within acceptable limits

### Protocol Compliance Success
- [ ] Full MCP protocol compliance
- [ ] Proper tool and resource discovery
- [ ] Correct error response formatting
- [ ] Resource URI handling works correctly

## Recommendations

Based on test results, consider the following improvements:

1. **Performance Optimization** - Focus on any endpoints exceeding 5s response time
2. **Error Handling** - Improve error messages for failed test cases
3. **Data Consistency** - Ensure test data remains available and consistent
4. **Monitoring** - Implement continuous monitoring for performance regression

## Next Steps

1. Review failed test cases and fix underlying issues
2. Update test data if database changes
3. Schedule regular e2e test execution
4. Monitor performance baselines over time

---
*Generated by Portal64 MCP E2E Test Runner*
EOF

    print_success "Test report generated: ${report_file}"
    print_info "Full test output available: ${RESULTS_FILE}"
}

# Main execution
main() {
    local start_time=$(date +%s)
    local failed_categories=0
    
    print_header "Portal64 MCP Server E2E Test Suite"
    print_info "Following test strategy from docs/e2e-test-strategy.md"
    
    # Check prerequisites
    if ! check_server; then
        exit 1
    fi
    
    # Setup environment
    setup_test_environment
    
    # Test execution order as specified in strategy
    print_info "Executing tests in strategic order..."
    
    # 1. Administrative Tests First
    if ! run_test_category "administrative" "Administrative Tools Tests"; then
        ((failed_categories++))
    fi
    
    # 2. Search Tools
    if ! run_test_category "search" "Search Tools Tests"; then
        ((failed_categories++))
    fi
    
    # 3. Detail Tools  
    if ! run_test_category "detail" "Detail Tools Tests"; then
        ((failed_categories++))
    fi
    
    # 4. Analysis Tools
    if ! run_test_category "analysis" "Analysis Tools Tests"; then
        ((failed_categories++))
    fi
    
    # 5. MCP Protocol Tests
    if ! run_test_category "protocol" "MCP Protocol Tests"; then
        ((failed_categories++))
    fi
    
    # 6. Error Scenario Testing
    if ! run_test_category "error_scenarios" "Error Scenario Tests"; then
        ((failed_categories++))
    fi
    
    # 7. Performance Testing
    if ! run_test_category "performance" "Performance Tests"; then
        ((failed_categories++))
    fi
    
    # Run benchmarks
    run_benchmarks
    
    # Generate report
    generate_report
    
    # Final summary
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
    print_header "Test Execution Complete"
    
    if [ $failed_categories -eq 0 ]; then
        print_success "All test categories passed! Total time: ${total_duration}s"
        echo "" >> "${RESULTS_FILE}"
        echo "FINAL RESULT: SUCCESS - All test categories passed" >> "${RESULTS_FILE}"
        exit 0
    else
        print_error "${failed_categories} test categories failed. Total time: ${total_duration}s"
        echo "" >> "${RESULTS_FILE}"
        echo "FINAL RESULT: FAILURE - ${failed_categories} test categories failed" >> "${RESULTS_FILE}"
        exit 1
    fi
}

# Help function
show_help() {
    cat << EOF
Portal64 MCP Server E2E Test Runner

USAGE:
    ./run_e2e_tests.sh [OPTIONS]

OPTIONS:
    -h, --help          Show this help message
    -c, --category      Run specific test category only
    -b, --benchmarks    Run benchmarks only
    -r, --report        Generate report from existing results
    --no-server-check   Skip server availability check

CATEGORIES:
    administrative      Administrative tools tests
    search             Search tools tests  
    detail             Detail tools tests
    analysis           Analysis tools tests
    protocol           MCP protocol tests
    error_scenarios    Error handling tests
    performance        Performance tests

EXAMPLES:
    ./run_e2e_tests.sh                           # Run all tests
    ./run_e2e_tests.sh -c search                 # Run search tests only
    ./run_e2e_tests.sh -b                        # Run benchmarks only
    ./run_e2e_tests.sh --no-server-check         # Skip server check

For more information, see docs/e2e-test-strategy.md
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -c|--category)
            CATEGORY="$2"
            shift 2
            ;;
        -b|--benchmarks)
            BENCHMARKS_ONLY=true
            shift
            ;;
        -r|--report)
            REPORT_ONLY=true
            shift
            ;;
        --no-server-check)
            SKIP_SERVER_CHECK=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Execute based on arguments
if [[ "$BENCHMARKS_ONLY" == "true" ]]; then
    setup_test_environment
    run_benchmarks
elif [[ "$REPORT_ONLY" == "true" ]]; then
    generate_report
elif [[ -n "$CATEGORY" ]]; then
    if [[ "$SKIP_SERVER_CHECK" != "true" ]] && ! check_server; then
        exit 1
    fi
    setup_test_environment
    case $CATEGORY in
        administrative|search|detail|analysis|protocol|error_scenarios|performance)
            run_test_category "$CATEGORY" "$(echo $CATEGORY | tr '_' ' ' | sed 's/\b\w/\U&/g') Tests"
            ;;
        *)
            print_error "Invalid category: $CATEGORY"
            show_help
            exit 1
            ;;
    esac
else
    main
fi
