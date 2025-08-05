# Portal64 MCP Server E2E Tests

This directory contains comprehensive end-to-end (e2e) tests for the Portal64 MCP Server, implementing the test strategy outlined in [docs/e2e-test-strategy.md](../docs/e2e-test-strategy.md).

## Overview

The e2e test suite validates all MCP functions through REST API calls against a running Portal64 MCP Server instance on `localhost:8888`. The tests use specific test data that ensures non-empty query results for reliable testing.

## Test Architecture

### Test Categories

1. **Administrative Tools Tests** - API health, cache stats, regions
2. **Search Tools Tests** - Player, club, and tournament search functionality
3. **Detail Tools Tests** - Profile and detail retrieval operations
4. **Analysis Tools Tests** - Statistical and historical data analysis
5. **MCP Protocol Tests** - Protocol compliance and resource access
6. **Error Scenario Tests** - Error handling and edge cases
7. **Performance Tests** - Response times and server stability

### Test Data

All tests use predefined data that guarantees non-empty results:

- **Players**: Query `"Minh Cuong"`, Player ID `"C0327-297"`
- **Clubs**: Query `"Altbach"`, Club ID `"C0327"`
- **Tournaments**: Query `"Ulm"`, Tournament ID `"C350-C01-SMU"`
- **Date Range**: `2023-2024` season

## Prerequisites

1. **Portal64 MCP Server** running on `localhost:8888`
2. **Go 1.21+** installed
3. **curl** for server health checks (Windows/Linux/macOS)
4. Test data available in the Portal64 database

## Quick Start

### Option 1: Run All Tests (Recommended)

```bash
# Linux/macOS
./test/run_e2e_tests.sh

# Windows
test\run_e2e_tests.bat
```

### Option 2: Run Individual Test Categories

```bash
# Linux/macOS
./test/run_e2e_tests.sh -c search          # Run search tests only
./test/run_e2e_tests.sh -c performance     # Run performance tests only

# Windows
test\run_e2e_tests.bat /c search           # Run search tests only
test\run_e2e_tests.bat /c performance      # Run performance tests only
```

### Option 3: Run Tests with Docker (Isolated Environment)

```bash
# Build and run all tests in Docker
./test/docker-e2e.sh test                  # Linux/macOS
test\docker-e2e.bat test                   # Windows

# Run specific test category in Docker
./test/docker-e2e.sh test-category search  # Linux/macOS
test\docker-e2e.bat test-category search   # Windows

# Run performance tests with Docker
./test/docker-e2e.sh performance           # Linux/macOS
test\docker-e2e.bat performance            # Windows
```

### Option 4: Run Tests Directly with Go

```bash
# Run all e2e tests
go test -v ./test/integration -run "TestPortal64MCP_E2E"

# Run pre-flight checks
go test -v ./test/integration -run "TestPortal64MCP_E2E_PreFlightCheck"

# Run specific test categories
go test -v ./test/integration -run "TestPortal64MCP_E2E_AllTools/1.*Administrative"
go test -v ./test/integration -run "TestPortal64MCP_E2E_ErrorScenarios"
go test -v ./test/integration -run "TestPortal64MCP_E2E_Performance"
```

## Test Files

### Core Test Files

- **`e2e_mcp_tools_test.go`** - Main e2e test suite covering all MCP tools
- **`e2e_error_scenarios_test.go`** - Error handling and edge case tests
- **`e2e_performance_test.go`** - Performance and benchmark tests
- **`e2e_test_utilities.go`** - Test utilities, data validation, and analysis tools

### Supporting Files

- **`api_client_test.go`** - Integration tests for API client
- **`testutil/testutil.go`** - Test utilities and mock servers
- **`fixtures/api_responses.json`** - Test data fixtures

### Test Runners

- **`run_e2e_tests.sh`** - Linux/macOS test runner script
- **`run_e2e_tests.bat`** - Windows test runner script
- **`docker-e2e.sh`** - Linux/macOS Docker test runner
- **`docker-e2e.bat`** - Windows Docker test runner

### Configuration Files

- **`e2e-config.yaml`** - Comprehensive test configuration
- **`docker-compose.e2e.yml`** - Docker Compose for e2e testing
- **`Dockerfile.e2e`** - Multi-stage Docker build for testing

## Docker Testing

### Docker Quick Start

The test suite includes comprehensive Docker support for isolated, reproducible testing environments.

```bash
# Build Docker images
./test/docker-e2e.sh build                 # Linux/macOS
test\docker-e2e.bat build                  # Windows

# Run all tests in Docker
./test/docker-e2e.sh test                  # Linux/macOS
test\docker-e2e.bat test                   # Windows

# Run specific category
./test/docker-e2e.sh test-category search  # Linux/macOS
test\docker-e2e.bat test-category search   # Windows

# Run performance tests and benchmarks
./test/docker-e2e.sh performance           # Linux/macOS
test\docker-e2e.bat performance            # Windows
```

### Docker Features

- **Isolated Environment** - Tests run in clean, reproducible containers
- **Multi-Stage Builds** - Optimized images for testing and production
- **Health Checks** - Automatic server readiness verification
- **Volume Mounting** - Test results preserved on host machine
- **Mock Server** - Offline testing capabilities with mock data
- **Results Viewer** - HTTP server for viewing test results
- **Development Mode** - Live reloading for active development

### Docker Commands

```bash
# Available Docker commands:
./test/docker-e2e.sh build          # Build images
./test/docker-e2e.sh test            # Run all tests
./test/docker-e2e.sh test-category CATEGORY  # Run specific category
./test/docker-e2e.sh performance     # Run performance tests
./test/docker-e2e.sh mock            # Start mock server
./test/docker-e2e.sh viewer          # Start results viewer
./test/docker-e2e.sh dev             # Development mode
./test/docker-e2e.sh logs            # Show container logs
./test/docker-e2e.sh status          # Show container status
./test/docker-e2e.sh cleanup         # Clean up containers
```

## Test Execution Strategy

Tests are executed in the following strategic order:

1. **Administrative Tests First** - Verify server health and connectivity
2. **Search Tools** - Test core search functionality
3. **Detail Tools** - Test data retrieval operations
4. **Analysis Tools** - Test advanced data analysis
5. **MCP Protocol** - Test protocol compliance
6. **Error Scenarios** - Test error handling
7. **Performance Tests** - Verify performance requirements

## Success Criteria

### Functional Success
- ✅ All 50+ test cases pass
- ✅ Non-empty results for all specified test data
- ✅ Proper error handling for invalid inputs
- ✅ Consistent response formats

### Performance Success
- ✅ Response times under 5 seconds for all calls
- ✅ Server remains stable under test load
- ✅ Memory usage stays within acceptable limits

### Protocol Compliance Success
- ✅ Full MCP protocol compliance
- ✅ Proper tool and resource discovery
- ✅ Correct error response formatting
- ✅ Resource URI handling works correctly

## Test Results

Test results are automatically saved to the `test-results/` directory:

- **`e2e_test_results_YYYYMMDD_HHMMSS.txt`** - Full test output
- **`test_report_YYYYMMDD_HHMMSS.md`** - Formatted test report

## Test Utilities and Validation

### Pre-flight Checks

Run comprehensive pre-flight validation before executing tests:

```bash
# Run pre-flight checks only
go test -v ./test/integration -run "TestPortal64MCP_E2E_PreFlightCheck"
```

Pre-flight checks validate:
- ✅ Server availability and responsiveness
- ✅ All required test data exists and is accessible
- ✅ MCP protocol endpoints are functional
- ✅ Test environment is properly configured

### Test Data Validation

The test suite includes automatic validation of required test data:

- **Player Data** - Validates search queries and specific player IDs
- **Club Data** - Validates club searches and profile access
- **Tournament Data** - Validates tournament searches and details
- **Region Data** - Validates region listings and address lookups

### Health Monitoring

Built-in health monitoring provides:

```bash
# Check system health
go test -v ./test/integration -run "TestPortal64MCP_HealthCheck"
```

Health checks include:
- 🏥 Server running status
- 🔌 API responsiveness
- 🛠️ MCP protocol functionality
- ⏱️ Response time monitoring
- 🚨 Error detection and reporting

### Result Analysis

Comprehensive test result analysis provides insights:

- **Success Rate Analysis** - Overall and per-category success rates
- **Performance Metrics** - Response time analysis and slow test detection
- **Failure Analysis** - Detailed failure categorization and recommendations
- **Trend Analysis** - Performance trends over multiple test runs
- **Report Generation** - Automated markdown reports with actionable insights

## Available Test Commands

### Test Runner Options

```bash
# Linux/macOS
./test/run_e2e_tests.sh [OPTIONS]

# Windows  
test\run_e2e_tests.bat [OPTIONS]
```

**Options:**
- `-h, --help` (Windows: `/h, /help`) - Show help message
- `-c, --category CATEGORY` (Windows: `/c CATEGORY`) - Run specific category only
- `-b, --benchmarks` (Windows: `/b`) - Run benchmarks only
- `-r, --report` (Windows: `/r`) - Generate report from existing results  
- `--no-server-check` (Windows: `/no-server-check`) - Skip server availability check

**Categories:**
- `administrative` - Administrative tools tests
- `search` - Search tools tests
- `detail` - Detail tools tests
- `analysis` - Analysis tools tests
- `protocol` - MCP protocol tests
- `error_scenarios` - Error handling tests
- `performance` - Performance tests

### Go Test Commands

```bash
# Run all e2e tests
go test -v ./test/integration

# Run with timeout
go test -v -timeout 300s ./test/integration

# Run specific tests
go test -v ./test/integration -run "TestPortal64MCP_E2E_AllTools"
go test -v ./test/integration -run "TestPortal64MCP_E2E_ErrorScenarios"
go test -v ./test/integration -run "TestPortal64MCP_E2E_Performance"

# Run benchmarks
go test -bench=. -benchmem ./test/integration

# Run benchmarks for specific functions
go test -bench=BenchmarkPortal64MCP_SearchPlayers ./test/integration
```

## Test Data Validation

Each test case verifies:

- ✅ Response status code is 200
- ✅ Response contains expected data structure
- ✅ Results are non-empty (where applicable)
- ✅ Data quality meets expectations
- ✅ No error messages in response

## Troubleshooting

### Common Issues

1. **Server Not Running**
   ```
   Error: Server is not running or not responding at http://localhost:8888
   ```
   **Solution:** Start the Portal64 MCP Server on port 8888

2. **Test Data Not Available**
   ```
   Error: Results are empty for test query
   ```
   **Solution:** Verify test data exists in the database

3. **Timeout Errors**
   ```
   Error: Request failed with timeout
   ```
   **Solution:** Increase timeout or check server performance

4. **Go Module Issues**
   ```
   Error: Cannot find module
   ```
   **Solution:** Run `go mod tidy` from project root

5. **Docker Issues**
   ```
   Error: Docker daemon not running
   ```
   **Solution:** Start Docker Desktop or Docker service

6. **Container Build Failures**
   ```
   Error: Failed to build Docker image
   ```
   **Solution:** Check Dockerfile and run `docker system prune` to clean up

7. **Port Conflicts**
   ```
   Error: Port 8888 already in use
   ```
   **Solution:** Stop other services or change test configuration

8. **Test Data Validation Failures**
   ```
   Error: Test data validation failed
   ```
   **Solution:** Run pre-flight checks to identify missing data

### Debug Mode

Enable verbose logging for debugging:

```bash
# Run with verbose output
go test -v ./test/integration -run "TestPortal64MCP_E2E" 

# Run single test with maximum detail
go test -v ./test/integration -run "TestPortal64MCP_E2E_AllTools/1.*Administrative.*Tools.*Tests/TC-AH-001"
```

### Manual Testing

Test individual MCP tools manually:

```bash
# Test health check
curl -X POST http://localhost:8888/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"check_api_health","arguments":{}}'

# Test player search
curl -X POST http://localhost:8888/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"search_players","arguments":{"query":"Minh Cuong"}}'
```

## Continuous Integration

The test suite includes comprehensive CI/CD integration with GitHub Actions.

### GitHub Actions Workflow

The `.github/workflows/e2e-tests.yml` workflow provides:

- **Automated Testing** - Runs on push, PR, and schedule
- **Matrix Strategy** - Tests all categories in parallel
- **Manual Triggers** - Run specific test categories on demand
- **Artifact Storage** - Test results and logs preserved
- **PR Comments** - Automatic test result comments on pull requests
- **Notifications** - Optional Slack integration for failures

### Workflow Features

- **Pre-flight Checks** - Validate dependencies and environment
- **Parallel Execution** - All test categories run simultaneously
- **Health Monitoring** - Server health verification before tests
- **Result Aggregation** - Consolidated reports from all test runs
- **Performance Tracking** - Benchmark results tracked over time
- **Error Reporting** - Detailed failure analysis and logs

### Manual Workflow Triggers

You can manually trigger the workflow with specific parameters:

1. Go to **Actions** tab in GitHub repository
2. Select **Portal64 MCP E2E Tests** workflow
3. Click **Run workflow**
4. Choose test category or run all tests
5. Optionally skip server health checks

### Local CI Simulation

```bash
# Simulate CI environment locally
export CI=true
export GITHUB_ACTIONS=true
./test/run_e2e_tests.sh

# Run specific category like CI
./test/run_e2e_tests.sh -c performance
```

## Continuous Integration

### GitHub Actions Integration

Add to `.github/workflows/e2e-tests.yml`:

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    
    services:
      portal64-server:
        image: portal64/mcp-server:latest
        ports:
          - 8888:8888
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Wait for server
      run: |
        timeout 60 bash -c 'until curl -f http://localhost:8888/health; do sleep 2; done'
    
    - name: Run E2E Tests
      run: ./test/run_e2e_tests.sh
    
    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: e2e-test-results
        path: test-results/
```

### Jenkins Integration

```groovy
pipeline {
    agent any
    
    stages {
        stage('Start Server') {
            steps {
                sh 'docker run -d -p 8888:8888 portal64/mcp-server:latest'
                sh 'timeout 60 bash -c "until curl -f http://localhost:8888/health; do sleep 2; done"'
            }
        }
        
        stage('Run E2E Tests') {
            steps {
                sh './test/run_e2e_tests.sh'
            }
            post {
                always {
                    archiveArtifacts artifacts: 'test-results/**/*', allowEmptyArchive: true
                    publishTestResults testResultsPattern: 'test-results/*.xml'
                }
            }
        }
    }
}
```

## Performance Benchmarks

The test suite includes comprehensive benchmarks:

```bash
# Run all benchmarks
go test -bench=. -benchmem ./test/integration

# Sample output:
# BenchmarkPortal64MCP_SearchPlayers-8         100    1500000 ns/op    2048 B/op     15 allocs/op
# BenchmarkPortal64MCP_GetPlayerProfile-8       50    2000000 ns/op    4096 B/op     25 allocs/op
# BenchmarkPortal64MCP_HealthCheck-8           200     800000 ns/op     512 B/op      8 allocs/op
```

## Contributing

### Adding New Tests

1. **Create test functions** following the naming convention:
   ```go
   func TestPortal64MCP_E2E_YourTestCategory(t *testing.T) {
       // Your test implementation
   }
   ```

2. **Update test runner scripts** to include new test categories

3. **Update fixtures** if new test data is needed

4. **Document test cases** in the strategy document

### Test Best Practices

- ✅ Use descriptive test names following the `TC-XX-001` pattern
- ✅ Verify both success and error scenarios
- ✅ Test with realistic data that guarantees non-empty results
- ✅ Include performance assertions where applicable
- ✅ Log meaningful information for debugging
- ✅ Clean up resources after tests

## Maintenance

### Regular Tasks

- **Monthly**: Verify test data availability
- **Quarterly**: Update performance baselines
- **On API changes**: Update test cases and fixtures
- **On new features**: Add corresponding test coverage

### Data Refresh

When the Portal64 database changes:

1. Update test data constants in test files
2. Update fixtures in `test/fixtures/api_responses.json`
3. Verify all tests still pass with new data
4. Update documentation with new test data references

## Support

For questions or issues with the e2e tests:

1. Check this README and the [e2e test strategy](../docs/e2e-test-strategy.md)
2. Review test output in `test-results/` directory
3. Run individual test categories to isolate issues
4. Enable verbose logging for detailed debugging information

---

## Test Suite Summary

This comprehensive e2e test implementation provides:

### ✅ **Complete Coverage**
- **50+ Test Cases** - All MCP tools and functions tested
- **Strategic Test Data** - Guaranteed non-empty results using specific test entities
- **Multiple Test Categories** - Administrative, search, detail, analysis, protocol, error handling, and performance
- **Comprehensive Validation** - Pre-flight checks, data validation, and health monitoring

### 🚀 **Multiple Execution Options**
- **Native Execution** - Direct Go test runs for development
- **Script Runners** - Cross-platform scripts for comprehensive testing
- **Docker Support** - Isolated, reproducible container-based testing
- **CI/CD Integration** - GitHub Actions workflow with parallel execution

### 📊 **Advanced Features**
- **Performance Benchmarking** - Response time analysis and server stability testing
- **Error Scenario Testing** - Comprehensive edge case and error handling validation
- **Result Analysis** - Automated insights and recommendations
- **Test Data Validation** - Automatic verification of required test data
- **Health Monitoring** - Real-time server and API health checks

### 🔧 **Developer Experience**
- **Easy Setup** - Single command execution for all test scenarios
- **Clear Documentation** - Comprehensive guides and troubleshooting
- **Flexible Configuration** - YAML-based configuration for customization
- **Debug Support** - Verbose logging and detailed error reporting
- **Cross-Platform** - Full support for Windows, macOS, and Linux

### 📈 **Quality Assurance**
- **Success Criteria** - Clear functional, performance, and protocol compliance requirements
- **Automated Reporting** - Detailed test reports with actionable insights
- **Continuous Monitoring** - Scheduled test runs and performance tracking
- **Failure Analysis** - Comprehensive error categorization and recommendations

*This test suite implements the comprehensive e2e test strategy outlined in `docs/e2e-test-strategy.md` and provides a robust foundation for ensuring Portal64 MCP Server quality and reliability.*
