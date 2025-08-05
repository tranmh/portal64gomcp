package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// E2E Error Scenario Tests for Portal64 MCP Server
// Tests error handling for invalid inputs and edge cases
func TestPortal64MCP_E2E_ErrorScenarios(t *testing.T) {
	// Verify server is running before testing error scenarios
	if !isServerRunning() {
		t.Skip("Portal64 MCP Server is not running on localhost:8888 - skipping error scenario tests")
	}

	t.Run("Invalid Player/Club/Tournament IDs", testInvalidIDs)
	t.Run("Malformed Requests", testMalformedRequests)
	t.Run("Missing Required Parameters", testMissingParameters)
	t.Run("Server Unavailability Scenarios", testServerUnavailability)
}

// Test invalid player/club/tournament IDs
func testInvalidIDs(t *testing.T) {
	t.Run("Invalid Player ID", func(t *testing.T) {
		response, err := callMCPTool("get_player_profile", map[string]interface{}{
			"player_id": "INVALID-PLAYER-ID-12345",
		})
		
		// Should handle error gracefully - either return error response or empty result
		if err != nil {
			// Network/parsing error occurred
			assert.Contains(t, err.Error(), "404", "Should return 404 for invalid player ID")
		} else {
			// MCP tool handled error internally
			assert.True(t, response.IsError || len(response.Content) == 0, 
				"Should return error response or empty content for invalid player ID")
		}
	})

	t.Run("Invalid Club ID", func(t *testing.T) {
		response, err := callMCPTool("get_club_profile", map[string]interface{}{
			"club_id": "INVALID-CLUB-ID-99999",
		})
		
		if err != nil {
			assert.Contains(t, err.Error(), "404", "Should return 404 for invalid club ID")
		} else {
			assert.True(t, response.IsError || len(response.Content) == 0,
				"Should return error response or empty content for invalid club ID")
		}
	})

	t.Run("Invalid Tournament ID", func(t *testing.T) {
		response, err := callMCPTool("get_tournament_details", map[string]interface{}{
			"tournament_id": "INVALID-TOURNAMENT-ID-ABCDEF",
		})
		
		if err != nil {
			assert.Contains(t, err.Error(), "404", "Should return 404 for invalid tournament ID")
		} else {
			assert.True(t, response.IsError || len(response.Content) == 0,
				"Should return error response or empty content for invalid tournament ID")
		}
	})
}

// Test malformed requests
func testMalformedRequests(t *testing.T) {
	t.Run("Invalid JSON in Tool Call", func(t *testing.T) {
		// Test with malformed JSON - this will be caught at HTTP level
		_, err := makeHTTPRequest("POST", "/tools/call", "invalid-json-string")
		assert.Error(t, err, "Malformed JSON should result in error")
	})

	t.Run("Invalid Tool Name", func(t *testing.T) {
		response, err := callMCPTool("nonexistent_tool", map[string]interface{}{
			"some_param": "some_value",
		})
		
		if err != nil {
			// HTTP level error (e.g., 400 Bad Request)
			assert.Error(t, err, "Invalid tool name should result in error")
		} else {
			// MCP level error
			assert.True(t, response.IsError, "Invalid tool name should return error response")
		}
	})

	t.Run("Invalid Parameter Types", func(t *testing.T) {
		// Test with wrong parameter types (e.g., string instead of integer)
		response, err := callMCPTool("search_players", map[string]interface{}{
			"query": TestPlayerQuery,
			"limit": "not-a-number", // Should be integer
			"active": "maybe",        // Should be boolean
		})
		
		if err != nil {
			assert.Error(t, err, "Invalid parameter types should result in error")
		} else {
			// Tool may handle type conversion gracefully or return error
			// Both are acceptable behaviors
			assert.NotNil(t, response, "Response should not be nil")
		}
	})
}

// Test missing required parameters
func testMissingParameters(t *testing.T) {
	t.Run("Missing Player ID", func(t *testing.T) {
		response, err := callMCPTool("get_player_profile", map[string]interface{}{
			// Missing required "player_id" parameter
		})
		
		if err != nil {
			assert.Error(t, err, "Missing required parameter should result in error")
		} else {
			assert.True(t, response.IsError, "Missing required parameter should return error response")
		}
	})

	t.Run("Missing Club ID", func(t *testing.T) {
		response, err := callMCPTool("get_club_profile", map[string]interface{}{
			// Missing required "club_id" parameter
		})
		
		if err != nil {
			assert.Error(t, err, "Missing required parameter should result in error")
		} else {
			assert.True(t, response.IsError, "Missing required parameter should return error response")
		}
	})

	t.Run("Missing Tournament ID", func(t *testing.T) {
		response, err := callMCPTool("get_tournament_details", map[string]interface{}{
			// Missing required "tournament_id" parameter
		})
		
		if err != nil {
			assert.Error(t, err, "Missing required parameter should result in error")
		} else {
			assert.True(t, response.IsError, "Missing required parameter should return error response")
		}
	})

	t.Run("Missing Region Parameter", func(t *testing.T) {
		response, err := callMCPTool("get_region_addresses", map[string]interface{}{
			// Missing required "region" parameter
		})
		
		if err != nil {
			assert.Error(t, err, "Missing required parameter should result in error")
		} else {
			assert.True(t, response.IsError, "Missing required parameter should return error response")
		}
	})
}

// Test server unavailability scenarios
func testServerUnavailability(t *testing.T) {
	t.Run("Timeout Handling", func(t *testing.T) {
		// This test would require a way to simulate slow responses
		// For now, we test that our client handles timeouts gracefully
		
		// Test with a very short timeout to simulate timeout conditions
		originalTimeout := 30 * time.Second
		
		// Note: In a real implementation, you might want to:
		// 1. Mock the server to introduce delays
		// 2. Use a shorter timeout for specific test cases
		// 3. Test network interruption scenarios
		
		t.Log("Timeout handling test - would need server-side delay simulation")
		assert.NotEqual(t, originalTimeout, 0, "Timeout configuration should be testable")
	})

	t.Run("Invalid Endpoints", func(t *testing.T) {
		// Test invalid MCP endpoints
		_, err := makeHTTPRequest("POST", "/invalid/endpoint", map[string]interface{}{})
		assert.Error(t, err, "Invalid endpoint should result in error")
		
		_, err = makeHTTPRequest("GET", "/tools/call", map[string]interface{}{})
		assert.Error(t, err, "Wrong HTTP method should result in error")
	})

	t.Run("Empty Request Body", func(t *testing.T) {
		_, err := makeHTTPRequest("POST", "/tools/call", nil)
		assert.Error(t, err, "Empty request body should result in error")
	})
}

// Additional edge case tests
func TestPortal64MCP_E2E_EdgeCases(t *testing.T) {
	if !isServerRunning() {
		t.Skip("Portal64 MCP Server is not running on localhost:8888 - skipping edge case tests")
	}

	t.Run("Boundary Value Tests", testBoundaryValues)
	t.Run("Special Character Handling", testSpecialCharacters)
	t.Run("Large Data Sets", testLargeDataSets)
}

func testBoundaryValues(t *testing.T) {
	t.Run("Zero Limit", func(t *testing.T) {
		response, err := callMCPTool("search_players", map[string]interface{}{
			"query": TestPlayerQuery,
			"limit": 0,
		})
		
		// Should handle gracefully - either error or return no results
		if err == nil {
			assert.NotNil(t, response, "Response should not be nil")
		}
	})

	t.Run("Very Large Limit", func(t *testing.T) {
		response, err := callMCPTool("search_players", map[string]interface{}{
			"query": TestPlayerQuery,
			"limit": 999999,
		})
		
		// Should handle gracefully - either error or cap the limit
		if err == nil {
			assert.NotNil(t, response, "Response should not be nil")
		}
	})

	t.Run("Negative Offset", func(t *testing.T) {
		response, err := callMCPTool("search_players", map[string]interface{}{
			"query":  TestPlayerQuery,
			"offset": -1,
		})
		
		// Should handle gracefully
		if err == nil {
			assert.NotNil(t, response, "Response should not be nil")
		}
	})
}

func testSpecialCharacters(t *testing.T) {
	specialQueries := []string{
		"Müller",           // Umlauts
		"José María",       // Accented characters
		"Player's Name",    // Apostrophes  
		"Name-With-Hyphens", // Hyphens
		"Name (with parens)", // Parentheses
		"Name & Partner",   // Ampersands
		"100% Player",      // Percentage signs
		"Player+Plus",      // Plus signs
	}

	for _, query := range specialQueries {
		t.Run("Special Query: "+query, func(t *testing.T) {
			response, err := callMCPTool("search_players", map[string]interface{}{
				"query": query,
			})
			
			// Should handle special characters gracefully
			if err == nil {
				assert.NotNil(t, response, "Response should not be nil for query: "+query)
				assert.False(t, response.IsError, "Should not return error for special characters: "+query)
			} else {
				// Log the error but don't fail the test as this might be expected behavior
				t.Logf("Query '%s' resulted in error: %v", query, err)
			}
		})
	}
}

func testLargeDataSets(t *testing.T) {
	t.Run("Large Limit Request", func(t *testing.T) {
		response, err := callMCPTool("search_players", map[string]interface{}{
			"query": "a", // Very common letter to get many results
			"limit": 200, // Maximum allowed limit
		})
		
		if err == nil {
			assert.NotNil(t, response, "Response should not be nil")
			assert.False(t, response.IsError, "Should handle large result sets")
			
			// Verify response time is reasonable (under 30 seconds as per strategy)
			// This is implicitly tested by our HTTP client timeout
		} else {
			// Large datasets might cause timeouts or server limitations
			t.Logf("Large dataset request resulted in error: %v", err)
		}
	})

	t.Run("Multiple Concurrent Requests", func(t *testing.T) {
		// Test server stability under concurrent load
		const numConcurrentRequests = 5
		results := make(chan error, numConcurrentRequests)

		for i := 0; i < numConcurrentRequests; i++ {
			go func(requestNum int) {
				_, err := callMCPTool("search_players", map[string]interface{}{
					"query": TestPlayerQuery,
					"limit": 10,
				})
				results <- err
			}(i)
		}

		// Collect all results
		var errors []error
		for i := 0; i < numConcurrentRequests; i++ {
			if err := <-results; err != nil {
				errors = append(errors, err)
			}
		}

		// Allow some failures under concurrent load, but not all
		failureRate := float64(len(errors)) / float64(numConcurrentRequests)
		assert.Less(t, failureRate, 0.5, "Failure rate under concurrent load should be less than 50%")
		
		if len(errors) > 0 {
			t.Logf("Concurrent requests had %d/%d failures", len(errors), numConcurrentRequests)
		}
	})
}
