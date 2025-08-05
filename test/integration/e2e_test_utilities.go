package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDataValidator validates that the required test data exists before running tests
type TestDataValidator struct {
	baseURL string
	client  *http.Client
}

// NewTestDataValidator creates a new test data validator
func NewTestDataValidator(baseURL string) *TestDataValidator {
	return &TestDataValidator{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// ValidateTestData validates all required test data exists
func (v *TestDataValidator) ValidateTestData(t *testing.T) bool {
	t.Log("Validating test data availability...")
	
	allValid := true
	
	// Validate player test data
	if !v.validatePlayerData(t) {
		allValid = false
	}
	
	// Validate club test data
	if !v.validateClubData(t) {
		allValid = false
	}
	
	// Validate tournament test data
	if !v.validateTournamentData(t) {
		allValid = false
	}
	
	// Validate region test data
	if !v.validateRegionData(t) {
		allValid = false
	}
	
	return allValid
}

// validatePlayerData validates player-specific test data
func (v *TestDataValidator) validatePlayerData(t *testing.T) bool {
	t.Log("Validating player test data...")
	
	// Test player search query
	response, err := callMCPTool("search_players", map[string]interface{}{
		"query": TestPlayerQuery,
		"limit": 1,
	})
	
	if err != nil {
		t.Errorf("Failed to validate player search data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Player search query '%s' returned no results", TestPlayerQuery)
		return false
	}
	
	// Test specific player ID
	response, err = callMCPTool("get_player_profile", map[string]interface{}{
		"player_id": TestPlayerID,
	})
	
	if err != nil {
		t.Errorf("Failed to validate player ID data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Player ID '%s' not found", TestPlayerID)
		return false
	}
	
	t.Log("✓ Player test data validated successfully")
	return true
}

// validateClubData validates club-specific test data
func (v *TestDataValidator) validateClubData(t *testing.T) bool {
	t.Log("Validating club test data...")
	
	// Test club search query
	response, err := callMCPTool("search_clubs", map[string]interface{}{
		"query": TestClubQuery,
		"limit": 1,
	})
	
	if err != nil {
		t.Errorf("Failed to validate club search data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Club search query '%s' returned no results", TestClubQuery)
		return false
	}
	
	// Test specific club ID
	response, err = callMCPTool("get_club_profile", map[string]interface{}{
		"club_id": TestClubID,
	})
	
	if err != nil {
		t.Errorf("Failed to validate club ID data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Club ID '%s' not found", TestClubID)
		return false
	}
	
	t.Log("✓ Club test data validated successfully")
	return true
}

// validateTournamentData validates tournament-specific test data
func (v *TestDataValidator) validateTournamentData(t *testing.T) bool {
	t.Log("Validating tournament test data...")
	
	// Test tournament search query
	response, err := callMCPTool("search_tournaments", map[string]interface{}{
		"query": TestTournamentQuery,
		"limit": 1,
	})
	
	if err != nil {
		t.Errorf("Failed to validate tournament search data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Tournament search query '%s' returned no results", TestTournamentQuery)
		return false
	}
	
	// Test specific tournament ID
	response, err = callMCPTool("get_tournament_details", map[string]interface{}{
		"tournament_id": TestTournamentID,
	})
	
	if err != nil {
		t.Errorf("Failed to validate tournament ID data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Tournament ID '%s' not found", TestTournamentID)
		return false
	}
	
	t.Log("✓ Tournament test data validated successfully")
	return true
}

// validateRegionData validates region-specific test data
func (v *TestDataValidator) validateRegionData(t *testing.T) bool {
	t.Log("Validating region test data...")
	
	// Test regions endpoint
	response, err := callMCPTool("get_regions", map[string]interface{}{})
	
	if err != nil {
		t.Errorf("Failed to validate regions data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Regions endpoint returned no results")
		return false
	}
	
	// Test region addresses
	response, err = callMCPTool("get_region_addresses", map[string]interface{}{
		"region": "Baden-Württemberg",
	})
	
	if err != nil {
		t.Errorf("Failed to validate region addresses data: %v", err)
		return false
	}
	
	if response.IsError || len(response.Content) == 0 {
		t.Errorf("Region addresses for 'Baden-Württemberg' returned no results")
		return false
	}
	
	t.Log("✓ Region test data validated successfully")
	return true
}

// Pre-flight test to validate test environment and data
func TestPortal64MCP_E2E_PreFlightCheck(t *testing.T) {
	t.Log("Running pre-flight checks for e2e test suite...")
	
	// Check server availability
	if !isServerRunning() {
		t.Fatal("Portal64 MCP Server is not running on localhost:8888")
	}
	t.Log("✓ Server is running and responsive")
	
	// Validate test data
	validator := NewTestDataValidator(BaseURL)
	if !validator.ValidateTestData(t) {
		t.Fatal("Test data validation failed - some required test data is missing")
	}
	t.Log("✓ All test data validated successfully")
	
	// Check MCP protocol endpoints
	if !validateMCPEndpoints(t) {
		t.Fatal("MCP protocol endpoints validation failed")
	}
	t.Log("✓ MCP protocol endpoints validated successfully")
	
	t.Log("✅ Pre-flight checks completed successfully - ready to run e2e tests")
}

// validateMCPEndpoints validates that MCP protocol endpoints are accessible
func validateMCPEndpoints(t *testing.T) bool {
	endpoints := []string{
		"/tools/list",
		"/resources/list",
	}
	
	for _, endpoint := range endpoints {
		_, err := makeHTTPRequest("POST", endpoint, map[string]interface{}{})
		if err != nil {
			t.Errorf("MCP endpoint %s is not accessible: %v", endpoint, err)
			return false
		}
	}
	
	return true
}

// TestResultAnalyzer analyzes test results and provides insights
type TestResultAnalyzer struct {
	Results []TestResult `json:"results"`
}

// TestResult represents a single test result
type TestResult struct {
	TestName     string        `json:"test_name"`
	Category     string        `json:"category"`
	Status       string        `json:"status"` // "PASS", "FAIL", "SKIP"
	Duration     time.Duration `json:"duration"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// AnalyzeResults analyzes test results and generates insights
func (a *TestResultAnalyzer) AnalyzeResults() TestAnalysis {
	analysis := TestAnalysis{
		TotalTests:      len(a.Results),
		PassedTests:     0,
		FailedTests:     0,
		SkippedTests:    0,
		Categories:      make(map[string]CategoryStats),
		SlowTests:       []TestResult{},
		FailedTestsList: []TestResult{},
		Recommendations: []string{},
	}
	
	var totalResponseTime time.Duration
	responseTimeCount := 0
	
	for _, result := range a.Results {
		// Count by status
		switch result.Status {
		case "PASS":
			analysis.PassedTests++
		case "FAIL":
			analysis.FailedTests++
			analysis.FailedTestsList = append(analysis.FailedTestsList, result)
		case "SKIP":
			analysis.SkippedTests++
		}
		
		// Category stats
		if _, exists := analysis.Categories[result.Category]; !exists {
			analysis.Categories[result.Category] = CategoryStats{}
		}
		stats := analysis.Categories[result.Category]
		stats.Total++
		if result.Status == "PASS" {
			stats.Passed++
		} else if result.Status == "FAIL" {
			stats.Failed++
		}
		analysis.Categories[result.Category] = stats
		
		// Response time analysis
		if result.ResponseTime > 0 {
			totalResponseTime += result.ResponseTime
			responseTimeCount++
			
			// Mark slow tests (> 5 seconds as per strategy)
			if result.ResponseTime > 5*time.Second {
				analysis.SlowTests = append(analysis.SlowTests, result)
			}
		}
	}
	
	// Calculate success rate
	if analysis.TotalTests > 0 {
		analysis.SuccessRate = (float64(analysis.PassedTests) / float64(analysis.TotalTests)) * 100
	}
	
	// Calculate average response time
	if responseTimeCount > 0 {
		analysis.AvgResponseTime = totalResponseTime / time.Duration(responseTimeCount)
	}
	
	// Generate recommendations
	analysis.Recommendations = a.generateRecommendations(analysis)
	
	return analysis
}

// TestAnalysis contains the analysis results
type TestAnalysis struct {
	TotalTests        int                    `json:"total_tests"`
	PassedTests       int                    `json:"passed_tests"`
	FailedTests       int                    `json:"failed_tests"`
	SkippedTests      int                    `json:"skipped_tests"`
	SuccessRate       float64               `json:"success_rate"`
	AvgResponseTime   time.Duration         `json:"avg_response_time"`
	Categories        map[string]CategoryStats `json:"categories"`
	SlowTests         []TestResult          `json:"slow_tests"`
	FailedTestsList   []TestResult          `json:"failed_tests_list"`
	Recommendations   []string              `json:"recommendations"`
}

// CategoryStats contains statistics for a test category
type CategoryStats struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// generateRecommendations generates recommendations based on test results
func (a *TestResultAnalyzer) generateRecommendations(analysis TestAnalysis) []string {
	recommendations := []string{}
	
	// Success rate recommendations
	if analysis.SuccessRate < 95.0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Success rate is %.1f%% - investigate failed tests to improve reliability", analysis.SuccessRate))
	}
	
	// Performance recommendations
	if analysis.AvgResponseTime > 3*time.Second {
		recommendations = append(recommendations,
			fmt.Sprintf("Average response time is %v - consider performance optimization", analysis.AvgResponseTime))
	}
	
	if len(analysis.SlowTests) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("%d tests exceeded 5-second response time limit - review performance", len(analysis.SlowTests)))
	}
	
	// Category-specific recommendations
	for category, stats := range analysis.Categories {
		if stats.Failed > 0 {
			failureRate := (float64(stats.Failed) / float64(stats.Total)) * 100
			if failureRate > 10.0 {
				recommendations = append(recommendations,
					fmt.Sprintf("Category '%s' has %.1f%% failure rate - needs attention", category, failureRate))
			}
		}
	}
	
	// General recommendations
	if len(analysis.FailedTestsList) > 0 {
		recommendations = append(recommendations, "Review failed test logs and fix underlying issues")
	}
	
	if analysis.SkippedTests > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("%d tests were skipped - ensure test environment is properly configured", analysis.SkippedTests))
	}
	
	return recommendations
}

// GenerateDetailedReport generates a detailed markdown report
func (a *TestAnalysis) GenerateDetailedReport() string {
	report := fmt.Sprintf(`# Portal64 MCP Server E2E Test Analysis Report

## Summary

- **Total Tests**: %d
- **Passed**: %d (%.1f%%)
- **Failed**: %d (%.1f%%)
- **Skipped**: %d (%.1f%%)
- **Success Rate**: %.1f%%
- **Average Response Time**: %v

`, a.TotalTests, a.PassedTests, (float64(a.PassedTests)/float64(a.TotalTests))*100,
		a.FailedTests, (float64(a.FailedTests)/float64(a.TotalTests))*100,
		a.SkippedTests, (float64(a.SkippedTests)/float64(a.TotalTests))*100,
		a.SuccessRate, a.AvgResponseTime)
	
	// Category breakdown
	report += "## Category Breakdown\n\n"
	for category, stats := range a.Categories {
		successRate := (float64(stats.Passed) / float64(stats.Total)) * 100
		report += fmt.Sprintf("- **%s**: %d tests, %d passed (%.1f%%)\n", 
			category, stats.Total, stats.Passed, successRate)
	}
	
	// Failed tests
	if len(a.FailedTestsList) > 0 {
		report += "\n## Failed Tests\n\n"
		for _, test := range a.FailedTestsList {
			report += fmt.Sprintf("- **%s**: %s\n", test.TestName, test.ErrorMessage)
		}
	}
	
	// Slow tests
	if len(a.SlowTests) > 0 {
		report += "\n## Slow Tests (>5s)\n\n"
		for _, test := range a.SlowTests {
			report += fmt.Sprintf("- **%s**: %v\n", test.TestName, test.ResponseTime)
		}
	}
	
	// Recommendations
	if len(a.Recommendations) > 0 {
		report += "\n## Recommendations\n\n"
		for _, rec := range a.Recommendations {
			report += fmt.Sprintf("- %s\n", rec)
		}
	}
	
	return report
}

// Health checker utility
func TestPortal64MCP_HealthCheck(t *testing.T) {
	health := PerformHealthCheck()
	
	assert.True(t, health.ServerRunning, "Server should be running")
	assert.True(t, health.APIResponsive, "API should be responsive")
	assert.True(t, health.MCPProtocolWorking, "MCP protocol should be working")
	
	if !health.AllHealthy() {
		t.Errorf("Health check failed: %+v", health)
	} else {
		t.Log("✅ All health checks passed")
	}
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	ServerRunning      bool          `json:"server_running"`
	APIResponsive      bool          `json:"api_responsive"`
	MCPProtocolWorking bool          `json:"mcp_protocol_working"`
	ResponseTime       time.Duration `json:"response_time"`
	ErrorMessage       string        `json:"error_message,omitempty"`
}

// AllHealthy returns true if all health checks pass
func (h *HealthStatus) AllHealthy() bool {
	return h.ServerRunning && h.APIResponsive && h.MCPProtocolWorking
}

// PerformHealthCheck performs a comprehensive health check
func PerformHealthCheck() HealthStatus {
	status := HealthStatus{}
	
	// Check if server is running
	start := time.Now()
	status.ServerRunning = isServerRunning()
	status.ResponseTime = time.Since(start)
	
	if !status.ServerRunning {
		status.ErrorMessage = "Server is not running or not responsive"
		return status
	}
	
	// Check API responsiveness
	_, err := makeHTTPRequest("GET", "/health", nil)
	status.APIResponsive = (err == nil)
	
	if !status.APIResponsive {
		status.ErrorMessage = fmt.Sprintf("API not responsive: %v", err)
		return status
	}
	
	// Check MCP protocol
	_, err = callMCPTool("check_api_health", map[string]interface{}{})
	status.MCPProtocolWorking = (err == nil)
	
	if !status.MCPProtocolWorking {
		status.ErrorMessage = fmt.Sprintf("MCP protocol not working: %v", err)
	}
	
	return status
}
