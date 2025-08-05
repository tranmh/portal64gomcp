@echo off
setlocal enabledelayedexpansion

REM Portal64 MCP Server E2E Test Docker Helper (Windows)
REM Simplifies running e2e tests with Docker Compose

REM Configuration
set COMPOSE_FILE=docker-compose.e2e.yml
set PROJECT_NAME=portal64-e2e

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

REM Check Docker and Docker Compose
:check_docker
where docker >nul 2>&1
if !errorlevel! neq 0 (
    call :print_error "Docker is not installed or not in PATH"
    exit /b 1
)

where docker-compose >nul 2>&1
if !errorlevel! neq 0 (
    call :print_error "Docker Compose is not installed or not in PATH"
    exit /b 1
)

call :print_success "Docker and Docker Compose are available"
goto :eof

REM Build images
:build_images
call :print_info "Building Docker images..."
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% build --no-cache
if !errorlevel! equ 0 (
    call :print_success "Images built successfully"
) else (
    call :print_error "Failed to build images"
    exit /b 1
)
goto :eof

REM Run full e2e test suite
:run_full_tests
call :print_header "Running Full E2E Test Suite"

REM Clean up any existing containers
call :cleanup

REM Start server and run tests
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% up --abort-on-container-exit portal64-server e2e-tests

if !errorlevel! equ 0 (
    call :print_success "Full E2E test suite completed successfully"
) else (
    call :print_error "Full E2E test suite failed"
    call :show_logs
    exit /b 1
)
goto :eof

REM Run specific test category
:run_category_tests
set category=%~1
call :print_header "Running E2E Tests - Category: %category%"

REM Clean up any existing containers
call :cleanup

REM Set environment variable and run tests
set TEST_CATEGORY=%category%
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% --profile category-tests up --abort-on-container-exit portal64-server e2e-tests-category

if !errorlevel! equ 0 (
    call :print_success "E2E tests for category '%category%' completed successfully"
) else (
    call :print_error "E2E tests for category '%category%' failed"
    call :show_logs
    exit /b 1
)
goto :eof

REM Run performance tests
:run_performance_tests
call :print_header "Running Performance Tests and Benchmarks"

REM Clean up any existing containers
call :cleanup

REM Run performance tests
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% --profile performance up --abort-on-container-exit portal64-server performance-tests

if !errorlevel! equ 0 (
    call :print_success "Performance tests completed successfully"
) else (
    call :print_error "Performance tests failed"
    call :show_logs
    exit /b 1
)
goto :eof

REM Run with mock server (offline testing)
:run_mock_tests
call :print_header "Running E2E Tests with Mock Server"

REM Clean up any existing containers
call :cleanup

REM Run with mock server
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% --profile mock up --abort-on-container-exit mock-server

call :print_info "Mock server is running on http://localhost:8081"
call :print_info "You can now run tests against the mock server manually"
goto :eof

REM Start results viewer
:start_results_viewer
call :print_header "Starting Test Results Viewer"

REM Ensure test results directory exists
if not exist "test-results" mkdir "test-results"

REM Start results viewer
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% --profile viewer up -d results-viewer

call :print_success "Results viewer started at http://localhost:8082"
call :print_info "Test results will be available at the above URL"
goto :eof

REM Show logs
:show_logs
call :print_info "Showing container logs..."
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% logs --tail=50
goto :eof

REM Clean up containers and volumes
:cleanup
call :print_info "Cleaning up existing containers..."
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% down -v --remove-orphans >nul 2>&1
goto :eof

REM Development mode
:dev_mode
call :print_header "Starting Development Mode"

REM Clean up
call :cleanup

REM Start in development mode with live reloading
docker-compose -f %COMPOSE_FILE% -f docker-compose.dev.yml -p %PROJECT_NAME% up --build
goto :eof

REM Show status
:show_status
call :print_header "Docker Container Status"
docker-compose -f %COMPOSE_FILE% -p %PROJECT_NAME% ps
goto :eof

REM Show help
:show_help
echo Portal64 MCP Server E2E Test Docker Helper (Windows)
echo.
echo USAGE:
echo     %~nx0 [COMMAND] [OPTIONS]
echo.
echo COMMANDS:
echo     build                   Build Docker images
echo     test                    Run full e2e test suite
echo     test-category CATEGORY  Run specific test category
echo     performance             Run performance tests and benchmarks
echo     mock                    Start mock server for offline testing
echo     viewer                  Start test results viewer
echo     dev                     Start development mode with live reloading
echo     logs                    Show container logs
echo     status                  Show container status
echo     cleanup                 Clean up containers and volumes
echo     help                    Show this help message
echo.
echo TEST CATEGORIES:
echo     administrative          Administrative tools tests
echo     search                 Search tools tests
echo     detail                 Detail tools tests
echo     analysis               Analysis tools tests
echo     protocol               MCP protocol tests
echo     error_scenarios        Error handling tests
echo     performance            Performance tests
echo.
echo EXAMPLES:
echo     %~nx0 build                           # Build images
echo     %~nx0 test                            # Run all tests
echo     %~nx0 test-category search            # Run search tests only
echo     %~nx0 performance                     # Run performance tests
echo     %~nx0 mock                            # Start mock server
echo     %~nx0 viewer                          # Start results viewer
echo     %~nx0 dev                             # Development mode
echo     %~nx0 cleanup                         # Clean up everything
echo.
echo ENVIRONMENT VARIABLES:
echo     TEST_CATEGORY          Test category to run (used with test-category)
echo     LOG_LEVEL             Log level (debug, info, warn, error)
echo     BASE_URL              Server base URL (default: http://portal64-server:8080)
echo.
echo For more information, see test/README.md
goto :eof

REM Main execution
:main
REM Check prerequisites
call :check_docker
if !errorlevel! neq 0 exit /b 1

REM Parse command
set command=%~1
if "%command%"=="" set command=help

if "%command%"=="build" (
    call :build_images
) else if "%command%"=="test" (
    call :run_full_tests
) else if "%command%"=="tests" (
    call :run_full_tests
) else if "%command%"=="test-category" (
    if "%~2"=="" (
        call :print_error "Test category is required"
        echo Available categories: administrative, search, detail, analysis, protocol, error_scenarios, performance
        exit /b 1
    )
    call :run_category_tests "%~2"
) else if "%command%"=="performance" (
    call :run_performance_tests
) else if "%command%"=="perf" (
    call :run_performance_tests
) else if "%command%"=="benchmark" (
    call :run_performance_tests
) else if "%command%"=="mock" (
    call :run_mock_tests
) else if "%command%"=="viewer" (
    call :start_results_viewer
) else if "%command%"=="results" (
    call :start_results_viewer
) else if "%command%"=="dev" (
    call :dev_mode
) else if "%command%"=="development" (
    call :dev_mode
) else if "%command%"=="logs" (
    call :show_logs
) else if "%command%"=="status" (
    call :show_status
) else if "%command%"=="cleanup" (
    call :cleanup
    call :print_success "Cleanup completed"
) else if "%command%"=="clean" (
    call :cleanup
    call :print_success "Cleanup completed"
) else if "%command%"=="help" (
    call :show_help
) else if "%command%"=="--help" (
    call :show_help
) else if "%command%"=="-h" (
    call :show_help
) else (
    call :print_error "Unknown command: %command%"
    echo.
    call :show_help
    exit /b 1
)

goto :eof

REM Entry point
call :main %*
