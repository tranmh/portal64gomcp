package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E2E Performance Tests for Portal64 MCP Server
// Tests performance requirements as specified in the e2e test strategy
func TestPortal64MCP_E2E_Performance(t *testing.T) {
	if !isServerRunning() {
		t.Skip("Portal64 MCP Server is not running on localhost:8888 - skipping performance tests")
	}

	t.Run("Response Time Requirements", testResponseTimes)
	t.Run("Server Stability Under Load", testServerStability)
	t.Run("Memory Usage Monitoring", testMemoryUsage)
	t.Run("Concurrent Request Handling", testConcurrentRequests)
}

// Test response times - all calls should be under 5 seconds as per strategy
func testResponseTimes(t *testing.T) {
	testCases := []struct {
		name string
		tool string
		args map[string]interface{}
	}{
		{
			name: "Search Players Response Time",
			tool: "search_players",
			args: map[string]interface{}{
				"query": TestPlayerQuery,
				"limit": 20,
			},
		},
		{
			name: "Search Clubs Response Time", 
			tool: "search_clubs",
			args: map[string]interface{}{
				"query": TestClubQuery,
				"limit": 20,
			},
		},
		{
			name: "Search Tournaments Response Time",
			tool: "search_tournaments",
			args: map[string]interface{}{
				"query": TestTournamentQuery,
				"limit": 20,
			},
		},
		{
			name: "Get Player Profile Response Time",
			tool: "get_player_profile",
			args: map[string]interface{}{
				"player_id": TestPlayerID,
			},
		},
		{
			name: "Get Club Profile Response Time",
			tool: "get_club_profile", 
			args: map[string]interface{}{
				"club_id": TestClubID,
			},
		},
		{
			name: "Get Tournament Details Response Time",
			tool: "get_tournament_details",
			args: map[string]interface{}{
				"tournament_id": TestTournamentID,
			},
		},
		{
			name: "Get Player Rating History Response Time",
			tool: "get_player_rating_history",
			args: map[string]interface{}{
				"player_id": TestPlayerID,
			},
		},
		{
			name: "Get Club Statistics Response Time",
			tool: "get_club_statistics",
			args: map[string]interface{}{
				"club_id": TestClubID,
			},
		},
		{
			name: "Health Check Response Time",
			tool: "check_api_health",
			args: map[string]interface{}{},
		},
		{
			name: "Cache Stats Response Time",
			tool: "get_cache_stats",
			args: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			
			response, err := callMCPTool(tc.tool, tc.args)
			
			elapsed := time.Since(start)
			
			// Verify call succeeded
			require.NoError(t, err, "Tool call should succeed for performance test")
			require.False(t, response.IsError, "Tool call should not return error")
			
			// Verify response time is under 5 seconds as per strategy
			assert.Less(t, elapsed, 5*time.Second, 
				"Response time should be under 5 seconds (actual: %v)", elapsed)
			
			// Log performance metrics
			t.Logf("%s completed in %v", tc.name, elapsed)
		})
	}
}

// Test server stability under load
func testServerStability(t *testing.T) {
	const (
		numRequests = 50
		concurrency = 10
		testDuration = 30 * time.Second
	)

	t.Run("Sustained Load Test", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numRequests)
		responseTimes := make(chan time.Duration, numRequests)
		
		start := time.Now()
		
		// Launch concurrent workers
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				requestCount := numRequests / concurrency
				for j := 0; j < requestCount; j++ {
					requestStart := time.Now()
					
					// Alternate between different tool calls for realistic load
					var err error
					switch j % 4 {
					case 0:
						_, err = callMCPTool("search_players", map[string]interface{}{
							"query": TestPlayerQuery,
							"limit": 10,
						})
					case 1:
						_, err = callMCPTool("search_clubs", map[string]interface{}{
							"query": TestClubQuery,
							"limit": 10,
						})
					case 2:
						_, err = callMCPTool("get_player_profile", map[string]interface{}{
							"player_id": TestPlayerID,
						})
					case 3:
						_, err = callMCPTool("check_api_health", map[string]interface{}{})
					}
					
					requestTime := time.Since(requestStart)
					responseTimes <- requestTime
					
					if err != nil {
						errors <- err
					}
				}
			}(i)
		}
		
		wg.Wait()
		close(errors)
		close(responseTimes)
		
		totalTime := time.Since(start)
		
		// Collect results
		var errorList []error
		var times []time.Duration
		
		for err := range errors {
			errorList = append(errorList, err)
		}
		
		for duration := range responseTimes {
			times = append(times, duration)
		}
		
		// Calculate statistics
		successRate := float64(len(times)-len(errorList)) / float64(len(times)) * 100
		
		var totalResponseTime time.Duration
		var maxResponseTime time.Duration
		for _, t := range times {
			totalResponseTime += t
			if t > maxResponseTime {
				maxResponseTime = t
			}
		}
		
		avgResponseTime := totalResponseTime / time.Duration(len(times))
		requestsPerSecond := float64(len(times)) / totalTime.Seconds()
		
		// Log performance metrics
		t.Logf("Load test completed:")
		t.Logf("  Total requests: %d", len(times))
		t.Logf("  Total time: %v", totalTime)
		t.Logf("  Success rate: %.2f%%", successRate)
		t.Logf("  Average response time: %v", avgResponseTime)
		t.Logf("  Max response time: %v", maxResponseTime)
		t.Logf("  Requests per second: %.2f", requestsPerSecond)
		t.Logf("  Errors: %d", len(errorList))
		
		// Verify performance requirements
		assert.GreaterOrEqual(t, successRate, 95.0, "Success rate should be at least 95%")
		assert.Less(t, avgResponseTime, 5*time.Second, "Average response time should be under 5 seconds")
		assert.Less(t, maxResponseTime, 10*time.Second, "Max response time should be under 10 seconds")
		
		// Server should remain stable (success rate high enough)
		if successRate < 95.0 {
			t.Errorf("Server stability test failed - success rate too low: %.2f%%", successRate)
		}
	})
}

// Test memory usage monitoring (basic test - would need more sophisticated monitoring in production)
func testMemoryUsage(t *testing.T) {
	t.Run("Memory Usage Monitoring", func(t *testing.T) {
		// This is a basic test - in a real environment you might want to:
		// 1. Monitor actual server memory usage
		// 2. Check for memory leaks
		// 3. Verify garbage collection efficiency
		
		// For now, we test that the server responds consistently
		// which indicates it's not running out of memory
		
		const iterations = 20
		var responseTimes []time.Duration
		
		for i := 0; i < iterations; i++ {
			start := time.Now()
			
			response, err := callMCPTool("search_players", map[string]interface{}{
				"query": TestPlayerQuery,
				"limit": 50, // Larger result set to use more memory
			})
			
			elapsed := time.Since(start)
			responseTimes = append(responseTimes, elapsed)
			
			require.NoError(t, err, "Request %d should succeed", i+1)
			require.False(t, response.IsError, "Request %d should not return error", i+1)
			
			// Small delay between requests
			time.Sleep(100 * time.Millisecond)
		}
		
		// Check that response times don't degrade significantly over time
		// (which could indicate memory issues)
		firstHalf := responseTimes[:iterations/2]
		secondHalf := responseTimes[iterations/2:]
		
		var firstHalfAvg, secondHalfAvg time.Duration
		for _, t := range firstHalf {
			firstHalfAvg += t
		}
		firstHalfAvg /= time.Duration(len(firstHalf))
		
		for _, t := range secondHalf {
			secondHalfAvg += t
		}
		secondHalfAvg /= time.Duration(len(secondHalf))
		
		// Response times shouldn't increase by more than 50% (indicating potential memory issues)
		degradationRatio := float64(secondHalfAvg) / float64(firstHalfAvg)
		assert.Less(t, degradationRatio, 1.5, 
			"Response times should not degrade significantly (first half avg: %v, second half avg: %v)", 
			firstHalfAvg, secondHalfAvg)
		
		t.Logf("Memory usage test - first half avg: %v, second half avg: %v, ratio: %.2f", 
			firstHalfAvg, secondHalfAvg, degradationRatio)
	})
}

// Test concurrent request handling capabilities
func testConcurrentRequests(t *testing.T) {
	concurrencyLevels := []int{5, 10, 20}
	
	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency Level %d", concurrency), func(t *testing.T) {
			var wg sync.WaitGroup
			results := make(chan result, concurrency)
			
			start := time.Now()
			
			// Launch concurrent requests
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(requestID int) {
					defer wg.Done()
					
					requestStart := time.Now()
					response, err := callMCPTool("search_players", map[string]interface{}{
						"query": TestPlayerQuery,
						"limit": 10,
					})
					requestTime := time.Since(requestStart)
					
					results <- result{
						requestID:    requestID,
						responseTime: requestTime,
						err:          err,
						hasError:     response != nil && response.IsError,
					}
				}(i)
			}
			
			wg.Wait()
			close(results)
			
			totalTime := time.Since(start)
			
			// Analyze results
			var successCount, errorCount int
			var totalResponseTime time.Duration
			var maxResponseTime time.Duration
			
			for res := range results {
				if res.err != nil || res.hasError {
					errorCount++
				} else {
					successCount++
				}
				
				totalResponseTime += res.responseTime
				if res.responseTime > maxResponseTime {
					maxResponseTime = res.responseTime
				}
			}
			
			avgResponseTime := totalResponseTime / time.Duration(concurrency)
			successRate := float64(successCount) / float64(concurrency) * 100
			
			t.Logf("Concurrency %d results:", concurrency)
			t.Logf("  Total time: %v", totalTime)
			t.Logf("  Success rate: %.2f%%", successRate)
			t.Logf("  Average response time: %v", avgResponseTime)
			t.Logf("  Max response time: %v", maxResponseTime)
			
			// Verify concurrent handling capabilities
			assert.GreaterOrEqual(t, successRate, 90.0, "Success rate should be at least 90% under concurrency")
			assert.Less(t, avgResponseTime, 10*time.Second, "Average response time should be reasonable under load")
			
			// Concurrent requests shouldn't take much longer than sequential requests
			// (server should handle concurrency well)
			expectedMaxTime := 8 * time.Second // Allow some overhead for concurrency
			assert.Less(t, maxResponseTime, expectedMaxTime, 
				"Max response time under concurrency should be reasonable")
		})
	}
}

// Benchmark tests
func BenchmarkPortal64MCP_SearchPlayers(b *testing.B) {
	if !isServerRunning() {
		b.Skip("Portal64 MCP Server is not running - skipping benchmarks")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callMCPTool("search_players", map[string]interface{}{
			"query": TestPlayerQuery,
			"limit": 10,
		})
	}
}

func BenchmarkPortal64MCP_GetPlayerProfile(b *testing.B) {
	if !isServerRunning() {
		b.Skip("Portal64 MCP Server is not running - skipping benchmarks")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callMCPTool("get_player_profile", map[string]interface{}{
			"player_id": TestPlayerID,
		})
	}
}

func BenchmarkPortal64MCP_HealthCheck(b *testing.B) {
	if !isServerRunning() {
		b.Skip("Portal64 MCP Server is not running - skipping benchmarks")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callMCPTool("check_api_health", map[string]interface{}{})
	}
}

func BenchmarkPortal64MCP_Concurrent(b *testing.B) {
	if !isServerRunning() {
		b.Skip("Portal64 MCP Server is not running - skipping benchmarks")
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			callMCPTool("search_players", map[string]interface{}{
				"query": TestPlayerQuery,
				"limit": 5,
			})
		}
	})
}

// Helper types
type result struct {
	requestID    int
	responseTime time.Duration
	err          error
	hasError     bool
}
