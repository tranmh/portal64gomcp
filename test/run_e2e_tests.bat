@echo off
setlocal enabledelayedexpansion

REM E2E Test Runner for Portal64 MCP Server (Windows)
REM Implements the test execution strategy from docs/e2e-test-strategy.md

REM Configuration
set BASE_URL=http://localhost:8888
set TEST_RESULTS_DIR=test-results
set TIMESTAMP=%date:~-4,4%%date:~-10,2%%date:~-7,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%
set RESULTS_FILE=%TEST_RESULTS_DIR%\e2e_test_results_%TIMESTAMP%.txt

REM Functions
:print_header
echo ============================================
echo %~1
echo ============================================
goto :eof

:print_success
echo [92m✓ %~1[0m
goto :eof

:print_error
echo [91m✗ %~1[0m
goto :eof

:print_warning
echo [93m⚠ %~1[0m
goto :eof

:print_info
echo [94mℹ %~1[0m
goto :eof

REM Check if server is running
:check_server
call :print_info "Checking if Portal64 MCP Server is running on %BASE_URL%..."

curl -s --connect-timeout 5 "%BASE_URL%/health" >nul 2>&1
if !errorlevel! equ 0 (
    call :print_success "Server is running and responsive"
    exit /b 0
) else (
    call :print_error "Server is not running or not responding at %BASE_URL%"
    call :print_info "Please start the Portal64 MCP Server on localhost:8888 before running tests"
    exit /b 1
)

REM Setup test environment
:setup_test_environment
call :print_info "Setting up test environment..."

REM Create test results directory
if not exist "%TEST_RESULTS_DIR%" mkdir "%TEST_RESULTS_DIR%"

REM Initialize results file
(
echo Portal64 MCP Server E2E Test Results
echo ====================================
echo Test Run: %TIMESTAMP%
echo Base URL: %BASE_URL%
echo Test Strategy: docs/e2e-test-strategy.md
echo.
) > "%RESULTS_FILE%"

call :print_success "Test environment setup complete"
goto :eof

REM Run specific test category
:run_test_category
set category=%~1
set description=%~2

call :print_header "Running %description%"

if "%category%"=="administrative" (
    call :run_go_test "TestPortal64MCP_E2E_AllTools/1.*Administrative.*Tools.*Tests"
) else if "%category%"=="search" (
    call :run_go_test "TestPortal64MCP_E2E_AllTools/2.*Search.*Tools.*Tests"
) else if "%category%"=="detail" (
    call :run_go_test "TestPortal64MCP_E2E_AllTools/3.*Detail.*Tools.*Tests"
) else if "%category%"=="analysis" (
    call :run_go_test "TestPortal64MCP_E2E_AllTools/4.*Analysis.*Tools.*Tests"
) else if "%category%"=="protocol" (
    call :run_go_test "TestPortal64MCP_E2E_AllTools/5.*MCP.*Protocol.*Tests"
) else if "%category%"=="error_scenarios" (
    call :run_go_test "TestPortal64MCP_E2E_ErrorScenarios"
) else if "%category%"=="performance" (
    call :run_go_test "TestPortal64MCP_E2E_Performance"
)

goto :eof

REM Run Go tests with proper formatting
:run_go_test
set test_pattern=%~1
set start_time=%time%

echo Running test pattern: %test_pattern% >> "%RESULTS_FILE%"
echo ---------------------------------------- >> "%RESULTS_FILE%"

go test -v -timeout 300s ./test/integration -run "%test_pattern%" >> "%RESULTS_FILE%" 2>&1
if !errorlevel! equ 0 (
    call :print_success "Test category completed successfully"
    echo ✓ Test category completed successfully >> "%RESULTS_FILE%"
    exit /b 0
) else (
    call :print_error "Test category failed"
    echo ✗ Test category failed >> "%RESULTS_FILE%"
    exit /b 1
)

REM Run benchmarks
:run_benchmarks
call :print_header "Running Performance Benchmarks"

echo Performance Benchmarks >> "%RESULTS_FILE%"
echo ===================== >> "%RESULTS_FILE%"

go test -bench=. -benchmem ./test/integration >> "%RESULTS_FILE%" 2>&1

if !errorlevel! equ 0 (
    call :print_success "Benchmarks completed successfully"
) else (
    call :print_warning "Some benchmarks may have failed or been skipped"
)
goto :eof

REM Generate test report
:generate_report
call :print_header "Generating Test Report"

set report_file=%TEST_RESULTS_DIR%\test_report_%TIMESTAMP%.md

(
echo # Portal64 MCP Server E2E Test Report
echo.
echo **Test Run:** %TIMESTAMP%
echo **Base URL:** %BASE_URL%
echo **Test Strategy:** docs/e2e-test-strategy.md
echo.
echo ## Test Execution Summary
echo.
echo The tests follow the execution order specified in the e2e test strategy:
echo.
echo 1. **Administrative Tests First** - Verify API health and basic connectivity
echo 2. **Search Tools** - Test all search functionalities
echo 3. **Detail Tools** - Test profile and detail retrieval
echo 4. **Analysis Tools** - Test statistical and historical data
echo 5. **MCP Protocol** - Test protocol compliance and resource access
echo 6. **Error Scenarios** - Test error handling and edge cases
echo 7. **Performance Tests** - Verify response times and server stability
echo.
echo ## Test Data Used
echo.
echo As specified in the e2e test strategy, all tests use predefined data:
echo.
echo - **Players**: Query "Minh Cuong", Player ID "C0327-297"
echo - **Clubs**: Query "Altbach", Club ID "C0327"
echo - **Tournaments**: Query "Ulm", Tournament ID "C350-C01-SMU"
echo - **Date Range**: 2023-2024 season
echo.
echo ## Detailed Results
echo.
echo See full test output in: `%RESULTS_FILE%`
echo.
echo ## Success Criteria
echo.
echo ### Functional Success
echo - [ ] All 50+ test cases pass
echo - [ ] Non-empty results for all specified test data
echo - [ ] Proper error handling for invalid inputs
echo - [ ] Consistent response formats
echo.
echo ### Performance Success
echo - [ ] Response times under 5 seconds for all calls
echo - [ ] Server remains stable under test load
echo - [ ] Memory usage stays within acceptable limits
echo.
echo ### Protocol Compliance Success
echo - [ ] Full MCP protocol compliance
echo - [ ] Proper tool and resource discovery
echo - [ ] Correct error response formatting
echo - [ ] Resource URI handling works correctly
echo.
echo ---
echo *Generated by Portal64 MCP E2E Test Runner*
) > "%report_file%"

call :print_success "Test report generated: %report_file%"
call :print_info "Full test output available: %RESULTS_FILE%"
goto :eof

REM Show help
:show_help
echo Portal64 MCP Server E2E Test Runner (Windows)
echo.
echo USAGE:
echo     run_e2e_tests.bat [OPTIONS]
echo.
echo OPTIONS:
echo     /h, /help           Show this help message
echo     /c category         Run specific test category only
echo     /b                  Run benchmarks only
echo     /r                  Generate report from existing results
echo     /no-server-check    Skip server availability check
echo.
echo CATEGORIES:
echo     administrative      Administrative tools tests
echo     search             Search tools tests
echo     detail             Detail tools tests
echo     analysis           Analysis tools tests
echo     protocol           MCP protocol tests
echo     error_scenarios    Error handling tests
echo     performance        Performance tests
echo.
echo EXAMPLES:
echo     run_e2e_tests.bat                      # Run all tests
echo     run_e2e_tests.bat /c search            # Run search tests only
echo     run_e2e_tests.bat /b                   # Run benchmarks only
echo     run_e2e_tests.bat /no-server-check     # Skip server check
echo.
echo For more information, see docs/e2e-test-strategy.md
goto :eof

REM Main execution
:main
set failed_categories=0

call :print_header "Portal64 MCP Server E2E Test Suite"
call :print_info "Following test strategy from docs/e2e-test-strategy.md"

REM Check prerequisites
if not "%SKIP_SERVER_CHECK%"=="true" (
    call :check_server
    if !errorlevel! neq 0 exit /b 1
)

REM Setup environment
call :setup_test_environment

REM Test execution order as specified in strategy
call :print_info "Executing tests in strategic order..."

REM 1. Administrative Tests First
call :run_test_category "administrative" "Administrative Tools Tests"
if !errorlevel! neq 0 set /a failed_categories+=1

REM 2. Search Tools
call :run_test_category "search" "Search Tools Tests"
if !errorlevel! neq 0 set /a failed_categories+=1

REM 3. Detail Tools
call :run_test_category "detail" "Detail Tools Tests"
if !errorlevel! neq 0 set /a failed_categories+=1

REM 4. Analysis Tools
call :run_test_category "analysis" "Analysis Tools Tests"
if !errorlevel! neq 0 set /a failed_categories+=1

REM 5. MCP Protocol Tests
call :run_test_category "protocol" "MCP Protocol Tests"
if !errorlevel! neq 0 set /a failed_categories+=1

REM 6. Error Scenario Testing
call :run_test_category "error_scenarios" "Error Scenario Tests"
if !errorlevel! neq 0 set /a failed_categories+=1

REM 7. Performance Testing
call :run_test_category "performance" "Performance Tests"
if !errorlevel! neq 0 set /a failed_categories+=1

REM Run benchmarks
call :run_benchmarks

REM Generate report
call :generate_report

REM Final summary
call :print_header "Test Execution Complete"

if !failed_categories! equ 0 (
    call :print_success "All test categories passed!"
    echo. >> "%RESULTS_FILE%"
    echo FINAL RESULT: SUCCESS - All test categories passed >> "%RESULTS_FILE%"
    exit /b 0
) else (
    call :print_error "!failed_categories! test categories failed."
    echo. >> "%RESULTS_FILE%"
    echo FINAL RESULT: FAILURE - !failed_categories! test categories failed >> "%RESULTS_FILE%"
    exit /b 1
)

REM Parse command line arguments
:parse_args
set arg=%~1
if "%arg%"=="" goto :main

if "%arg%"=="/h" goto :show_help
if "%arg%"=="/help" goto :show_help
if "%arg%"=="/c" (
    set CATEGORY=%~2
    shift
    shift
    goto :parse_args
)
if "%arg%"=="/b" (
    set BENCHMARKS_ONLY=true
    shift
    goto :parse_args
)
if "%arg%"=="/r" (
    set REPORT_ONLY=true
    shift
    goto :parse_args
)
if "%arg%"=="/no-server-check" (
    set SKIP_SERVER_CHECK=true
    shift
    goto :parse_args
)

call :print_error "Unknown option: %arg%"
call :show_help
exit /b 1

REM Execute based on arguments
if "%BENCHMARKS_ONLY%"=="true" (
    call :setup_test_environment
    call :run_benchmarks
    exit /b 0
)

if "%REPORT_ONLY%"=="true" (
    call :generate_report
    exit /b 0
)

if not "%CATEGORY%"=="" (
    if not "%SKIP_SERVER_CHECK%"=="true" (
        call :check_server
        if !errorlevel! neq 0 exit /b 1
    )
    call :setup_test_environment
    
    if "%CATEGORY%"=="administrative" goto :run_category
    if "%CATEGORY%"=="search" goto :run_category
    if "%CATEGORY%"=="detail" goto :run_category
    if "%CATEGORY%"=="analysis" goto :run_category
    if "%CATEGORY%"=="protocol" goto :run_category
    if "%CATEGORY%"=="error_scenarios" goto :run_category
    if "%CATEGORY%"=="performance" goto :run_category
    
    call :print_error "Invalid category: %CATEGORY%"
    call :show_help
    exit /b 1
    
    :run_category
    call :run_test_category "%CATEGORY%" "%CATEGORY% Tests"
    exit /b !errorlevel!
)

goto :main

REM Entry point
call :parse_args %*
