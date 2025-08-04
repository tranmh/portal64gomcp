#!/usr/bin/env bash

# Test Coverage Analysis Script
# Analyzes test coverage and generates detailed reports

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-85}
OUTPUT_DIR="coverage"

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

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run tests with coverage
log_info "Running tests with coverage analysis..."
go test -v -race -coverprofile="$OUTPUT_DIR/coverage.out" -covermode=atomic ./...

if [ $? -ne 0 ]; then
    log_error "Tests failed"
    exit 1
fi

# Generate HTML coverage report
log_info "Generating HTML coverage report..."
go tool cover -html="$OUTPUT_DIR/coverage.out" -o "$OUTPUT_DIR/coverage.html"

# Generate detailed coverage report
log_info "Generating detailed coverage report..."
go tool cover -func="$OUTPUT_DIR/coverage.out" > "$OUTPUT_DIR/coverage_detailed.txt"

# Extract overall coverage percentage
COVERAGE=$(go tool cover -func="$OUTPUT_DIR/coverage.out" | grep total | awk '{print $3}' | sed 's/%//')

# Generate coverage summary
log_info "Generating coverage summary..."
cat > "$OUTPUT_DIR/coverage_summary.md" << EOF
# Test Coverage Report

Generated on: $(date)
Coverage Threshold: ${COVERAGE_THRESHOLD}%
Total Coverage: ${COVERAGE}%

## Overall Status
$(if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l 2>/dev/null || echo "0") )); then
    echo "✅ **PASSED** - Coverage meets threshold"
else
    echo "❌ **FAILED** - Coverage below threshold"
fi)

## Coverage by Package

EOF

# Add package-level coverage details
go tool cover -func="$OUTPUT_DIR/coverage.out" | grep -v total | while read line; do
    if [[ $line == *"github.com/svw-info/portal64gomcp/"* ]]; then
        package=$(echo $line | cut -d' ' -f1 | sed 's|github.com/svw-info/portal64gomcp/||')
        coverage_percent=$(echo $line | awk '{print $NF}')
        echo "- **$package**: $coverage_percent" >> "$OUTPUT_DIR/coverage_summary.md"
    fi
done

# Add detailed analysis
cat >> "$OUTPUT_DIR/coverage_summary.md" << EOF

## Test Files Analysis

EOF

# Analyze test files
find . -name "*_test.go" | while read test_file; do
    package_path=$(dirname "$test_file" | sed 's|^\./||')
    test_count=$(grep -c "^func Test" "$test_file" 2>/dev/null || echo "0")
    benchmark_count=$(grep -c "^func Benchmark" "$test_file" 2>/dev/null || echo "0")
    
    echo "### $package_path/$(basename "$test_file")" >> "$OUTPUT_DIR/coverage_summary.md"
    echo "- Tests: $test_count" >> "$OUTPUT_DIR/coverage_summary.md"
    echo "- Benchmarks: $benchmark_count" >> "$OUTPUT_DIR/coverage_summary.md"
    echo "" >> "$OUTPUT_DIR/coverage_summary.md"
done

# Generate coverage badge
generate_coverage_badge() {
    local coverage=$1
    local color
    
    if (( $(echo "$coverage >= 90" | bc -l 2>/dev/null || echo "0") )); then
        color="brightgreen"
    elif (( $(echo "$coverage >= 80" | bc -l 2>/dev/null || echo "0") )); then
        color="green"
    elif (( $(echo "$coverage >= 70" | bc -l 2>/dev/null || echo "0") )); then
        color="yellow"
    elif (( $(echo "$coverage >= 60" | bc -l 2>/dev/null || echo "0") )); then
        color="orange"
    else
        color="red"
    fi
    
    echo "[![Coverage](https://img.shields.io/badge/coverage-${coverage}%25-${color})](coverage/coverage.html)" > "$OUTPUT_DIR/coverage_badge.md"
}

generate_coverage_badge "$COVERAGE"

# Check for uncovered critical files
log_info "Checking for uncovered critical files..."
CRITICAL_FILES=(
    "internal/config/config.go"
    "internal/api/client.go"
    "internal/mcp/server.go"
    "internal/mcp/protocol.go"
)

UNCOVERED_CRITICAL=()
for file in "${CRITICAL_FILES[@]}"; do
    if [ -f "$file" ]; then
        file_coverage=$(go tool cover -func="$OUTPUT_DIR/coverage.out" | grep "$file" | awk '{print $3}' | sed 's/%//' || echo "0")
        if (( $(echo "$file_coverage < 70" | bc -l 2>/dev/null || echo "1") )); then
            UNCOVERED_CRITICAL+=("$file ($file_coverage%)")
        fi
    fi
done

if [ ${#UNCOVERED_CRITICAL[@]} -gt 0 ]; then
    log_warning "Critical files with low coverage:"
    for file in "${UNCOVERED_CRITICAL[@]}"; do
        log_warning "  - $file"
    done
fi

# Generate test recommendations
log_info "Generating test recommendations..."
cat > "$OUTPUT_DIR/test_recommendations.md" << EOF
# Test Coverage Recommendations

## High Priority
- Add tests for files with < 70% coverage
- Focus on error handling scenarios
- Add integration tests for critical paths

## Medium Priority
- Add benchmark tests for performance-critical functions
- Add property-based tests for complex logic
- Add mutation tests to verify test quality

## Low Priority
- Add fuzzing tests for input validation
- Add stress tests for concurrent operations
- Add chaos engineering tests

## Files Needing Attention
EOF

# Find files with low coverage
go tool cover -func="$OUTPUT_DIR/coverage.out" | grep -v total | while read line; do
    if [[ $line == *".go"* ]]; then
        filename=$(echo $line | awk '{print $1}')
        coverage_percent=$(echo $line | awk '{print $3}' | sed 's/%//')
        
        if (( $(echo "$coverage_percent < 70" | bc -l 2>/dev/null || echo "1") )); then
            echo "- $filename ($coverage_percent%)" >> "$OUTPUT_DIR/test_recommendations.md"
        fi
    fi
done

# Display results
echo
log_info "Coverage Analysis Complete"
echo "=========================="
log_info "Total Coverage: $COVERAGE%"
log_info "Threshold: $COVERAGE_THRESHOLD%"

if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l 2>/dev/null || echo "0") )); then
    log_success "Coverage meets threshold! ✅"
    exit_code=0
else
    log_error "Coverage below threshold! ❌"
    exit_code=1
fi

echo
log_info "Reports generated in $OUTPUT_DIR/:"
log_info "  - coverage.html (visual report)"
log_info "  - coverage_summary.md (markdown summary)"
log_info "  - coverage_detailed.txt (detailed text report)"
log_info "  - test_recommendations.md (improvement suggestions)"
log_info "  - coverage_badge.md (badge for README)"

# Show top uncovered functions
echo
log_info "Top uncovered functions:"
go tool cover -func="$OUTPUT_DIR/coverage.out" | grep -v "100.0%" | grep -v "total:" | sort -k3 -n | head -10 | while read line; do
    func_name=$(echo $line | awk '{print $2}')
    coverage_percent=$(echo $line | awk '{print $3}')
    echo "  - $func_name ($coverage_percent)"
done

# Generate JSON report for CI/CD
cat > "$OUTPUT_DIR/coverage.json" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "coverage": $COVERAGE,
    "threshold": $COVERAGE_THRESHOLD,
    "status": "$(if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l 2>/dev/null || echo "0") )); then echo "PASS"; else echo "FAIL"; fi)",
    "files_analyzed": $(go tool cover -func="$OUTPUT_DIR/coverage.out" | grep -v total | wc -l),
    "critical_files_low_coverage": ${#UNCOVERED_CRITICAL[@]}
}
EOF

log_info "JSON report: $OUTPUT_DIR/coverage.json"

exit $exit_code
