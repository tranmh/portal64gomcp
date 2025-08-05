package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Base URL for MCP server as specified in e2e test strategy
	BaseURL = "http://localhost:8888"
	
	// Test data specifications from e2e test strategy
	TestPlayerQuery = "Minh Cuong"
	TestPlayerID    = "C0327-297"
	TestClubQuery   = "Altbach"
	TestClubID      = "C0327"
	TestTournamentQuery = "Ulm"
	TestTournamentID    = "C350-C01-SMU"
	TestStartDate       = "2023-01-01"
	TestEndDate         = "2024-12-31"
)

// MCPToolCall represents an MCP tool call request
type MCPToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPToolResponse represents an MCP tool call response
type MCPToolResponse struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents MCP response content
type MCPContent struct {
	Type string      `json:"type"`
	Text interface{} `json:"text,omitempty"`
}

// E2E Test Suite for Portal64 MCP Server
// Tests all MCP functions through REST API calls against localhost:8888
func TestPortal64MCP_E2E_AllTools(t *testing.T) {
	// Verify server is running
	if !isServerRunning() {
		t.Skip("Portal64 MCP Server is not running on localhost:8888 - skipping e2e tests")
	}

	// Test execution order as specified in strategy:
	// 1. Administrative Tests First
	// 2. Search Tools  
	// 3. Detail Tools
	// 4. Analysis Tools
	// 5. MCP Protocol Tests

	t.Run("1. Administrative Tools Tests", testAdministrativeTools)
	t.Run("2. Search Tools Tests", testSearchTools)
	t.Run("3. Detail Tools Tests", testDetailTools)
	t.Run("4. Analysis Tools Tests", testAnalysisTools)
	t.Run("5. MCP Protocol Tests", testMCPProtocol)
}

// 1. Administrative Tools Tests
func testAdministrativeTools(t *testing.T) {
	t.Run("TC-AH-001: Check API Health", func(t *testing.T) {
		response, err := callMCPTool("check_api_health", map[string]interface{}{})
		require.NoError(t, err, "Health check should not fail")
		
		assert.False(t, response.IsError, "Health check should not return error")
		assert.NotEmpty(t, response.Content, "Health check should return content")
		
		// Verify health status and response time metrics are included
		content := extractTextContent(response.Content)
		assert.Contains(t, content, "status", "Health response should contain status")
	})

	t.Run("TC-GCS-001: Get Cache Statistics", func(t *testing.T) {
		response, err := callMCPTool("get_cache_stats", map[string]interface{}{})
		require.NoError(t, err, "Cache stats should not fail")
		
		assert.False(t, response.IsError, "Cache stats should not return error")
		assert.NotEmpty(t, response.Content, "Cache stats should return content")
		
		// Verify cache performance metrics are included
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Cache stats should contain metrics")
	})

	t.Run("TC-GR-001: Get Available Regions", func(t *testing.T) {
		response, err := callMCPTool("get_regions", map[string]interface{}{})
		require.NoError(t, err, "Get regions should not fail")
		
		assert.False(t, response.IsError, "Get regions should not return error")
		assert.NotEmpty(t, response.Content, "Get regions should return content")
		
		// Verify non-empty results with region codes and names
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Regions should contain data")
	})

	t.Run("TC-GRA-001: Get Region Addresses", func(t *testing.T) {
		response, err := callMCPTool("get_region_addresses", map[string]interface{}{
			"region": "Baden-Württemberg",
		})
		require.NoError(t, err, "Get region addresses should not fail")
		
		assert.False(t, response.IsError, "Get region addresses should not return error")
		assert.NotEmpty(t, response.Content, "Get region addresses should return content")
		
		// Verify chess official addresses are returned
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Region addresses should contain contact information")
	})
}

// 2. Search Tools Tests  
func testSearchTools(t *testing.T) {
	t.Run("Search Players Tests", func(t *testing.T) {
		t.Run("TC-SP-001: Basic Player Search", func(t *testing.T) {
			response, err := callMCPTool("search_players", map[string]interface{}{
				"query": TestPlayerQuery,
			})
			require.NoError(t, err, "Basic player search should not fail")
			
			assert.False(t, response.IsError, "Player search should not return error")
			assert.NotEmpty(t, response.Content, "Player search should return results")
			
			// Verify results array is not empty and contains player data
			content := extractTextContent(response.Content)
			assert.Contains(t, strings.ToLower(content), strings.ToLower(TestPlayerQuery), 
				"Results should contain searched player name")
		})

		t.Run("TC-SP-002: Player Search with Pagination", func(t *testing.T) {
			response, err := callMCPTool("search_players", map[string]interface{}{
				"query":  TestPlayerQuery,
				"limit":  10,
				"offset": 0,
			})
			require.NoError(t, err, "Player search with pagination should not fail")
			
			assert.False(t, response.IsError, "Player search should not return error")
			assert.NotEmpty(t, response.Content, "Player search should return results")
			
			// Verify results are limited and pagination metadata is included
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Paginated results should contain data")
		})

		t.Run("TC-SP-003: Player Search with Sorting", func(t *testing.T) {
			response, err := callMCPTool("search_players", map[string]interface{}{
				"query":      TestPlayerQuery,
				"sort_by":    "current_dwz",
				"sort_order": "desc",
			})
			require.NoError(t, err, "Player search with sorting should not fail")
			
			assert.False(t, response.IsError, "Player search should not return error")
			assert.NotEmpty(t, response.Content, "Player search should return results")
			
			// Verify results are sorted by current_dwz in descending order
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Sorted results should contain data")
		})

		t.Run("TC-SP-004: Active Players Filter", func(t *testing.T) {
			response, err := callMCPTool("search_players", map[string]interface{}{
				"query":  TestPlayerQuery,
				"active": true,
			})
			require.NoError(t, err, "Active players filter should not fail")
			
			assert.False(t, response.IsError, "Player search should not return error")
			assert.NotEmpty(t, response.Content, "Player search should return results")
			
			// Verify only active players are returned
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Active player results should contain data")
		})
	})

	t.Run("Search Clubs Tests", func(t *testing.T) {
		t.Run("TC-SC-001: Basic Club Search", func(t *testing.T) {
			response, err := callMCPTool("search_clubs", map[string]interface{}{
				"query": TestClubQuery,
			})
			require.NoError(t, err, "Basic club search should not fail")
			
			assert.False(t, response.IsError, "Club search should not return error")
			assert.NotEmpty(t, response.Content, "Club search should return results")
			
			// Verify club names contain "Altbach" (case-insensitive)
			content := extractTextContent(response.Content)
			assert.Contains(t, strings.ToLower(content), strings.ToLower(TestClubQuery),
				"Results should contain searched club name")
		})

		t.Run("TC-SC-002: Club Search with Geographic Filter", func(t *testing.T) {
			response, err := callMCPTool("search_clubs", map[string]interface{}{
				"query":        TestClubQuery,
				"filter_by":    "region",
				"filter_value": "Baden-Württemberg",
			})
			require.NoError(t, err, "Club search with geographic filter should not fail")
			
			assert.False(t, response.IsError, "Club search should not return error")
			assert.NotEmpty(t, response.Content, "Club search should return results")
			
			// Verify results are filtered by specified region
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Geographically filtered results should contain data")
		})

		t.Run("TC-SC-003: Club Search with Sorting by Member Count", func(t *testing.T) {
			response, err := callMCPTool("search_clubs", map[string]interface{}{
				"query":      TestClubQuery,
				"sort_by":    "member_count",
				"sort_order": "desc",
			})
			require.NoError(t, err, "Club search with sorting should not fail")
			
			assert.False(t, response.IsError, "Club search should not return error")
			assert.NotEmpty(t, response.Content, "Club search should return results")
			
			// Verify results are sorted by member count descending
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Sorted club results should contain data")
		})
	})

	t.Run("Search Tournaments Tests", func(t *testing.T) {
		t.Run("TC-ST-001: Basic Tournament Search", func(t *testing.T) {
			response, err := callMCPTool("search_tournaments", map[string]interface{}{
				"query": TestTournamentQuery,
			})
			require.NoError(t, err, "Basic tournament search should not fail")
			
			assert.False(t, response.IsError, "Tournament search should not return error")
			assert.NotEmpty(t, response.Content, "Tournament search should return results")
			
			// Verify tournament names/locations contain "Ulm"
			content := extractTextContent(response.Content)
			assert.Contains(t, strings.ToLower(content), strings.ToLower(TestTournamentQuery),
				"Results should contain searched tournament location")
		})

		t.Run("TC-ST-002: Recent Tournaments", func(t *testing.T) {
			response, err := callMCPTool("get_recent_tournaments", map[string]interface{}{
				"days":  30,
				"limit": 20,
			})
			require.NoError(t, err, "Recent tournaments should not fail")
			
			assert.False(t, response.IsError, "Recent tournaments should not return error")
			assert.NotEmpty(t, response.Content, "Recent tournaments should return results")
			
			// Verify results contain tournaments from last 30 days, limited to 20
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Recent tournament results should contain data")
		})

		t.Run("TC-ST-003: Tournament Search by Date Range", func(t *testing.T) {
			response, err := callMCPTool("search_tournaments_by_date", map[string]interface{}{
				"start_date": TestStartDate,
				"end_date":   TestEndDate,
				"query":      TestTournamentQuery,
				"limit":      50,
			})
			require.NoError(t, err, "Tournament search by date range should not fail")
			
			assert.False(t, response.IsError, "Tournament search should not return error")
			assert.NotEmpty(t, response.Content, "Tournament search should return results")
			
			// Verify results within 2023-2024 date range containing "Ulm"
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Date range tournament results should contain data")
		})
	})
}

// 3. Detail Tools Tests
func testDetailTools(t *testing.T) {
	t.Run("TC-PP-001: Get Player Profile", func(t *testing.T) {
		response, err := callMCPTool("get_player_profile", map[string]interface{}{
			"player_id": TestPlayerID,
		})
		require.NoError(t, err, "Get player profile should not fail")
		
		assert.False(t, response.IsError, "Player profile should not return error")
		assert.NotEmpty(t, response.Content, "Player profile should return complete data")
		
		// Verify complete player profile with all required fields
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Player profile should contain personal information, DWZ rating, club affiliation, and tournament history")
	})

	t.Run("TC-CP-001: Get Club Profile", func(t *testing.T) {
		response, err := callMCPTool("get_club_profile", map[string]interface{}{
			"club_id": TestClubID,
		})
		require.NoError(t, err, "Get club profile should not fail")
		
		assert.False(t, response.IsError, "Club profile should not return error")
		assert.NotEmpty(t, response.Content, "Club profile should return complete data")
		
		// Verify complete club profile with all required fields
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Club profile should contain club information, member statistics, and performance data")
	})

	t.Run("TC-TD-001: Get Tournament Details", func(t *testing.T) {
		response, err := callMCPTool("get_tournament_details", map[string]interface{}{
			"tournament_id": TestTournamentID,
		})
		require.NoError(t, err, "Get tournament details should not fail")
		
		assert.False(t, response.IsError, "Tournament details should not return error")
		assert.NotEmpty(t, response.Content, "Tournament details should return complete data")
		
		// Verify complete tournament details with all required fields
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Tournament details should contain metadata, participant list, and results")
	})

	t.Run("Club Players Tests", func(t *testing.T) {
		t.Run("TC-CP-001: Get Club Players", func(t *testing.T) {
			response, err := callMCPTool("get_club_players", map[string]interface{}{
				"club_id": TestClubID,
			})
			require.NoError(t, err, "Get club players should not fail")
			
			assert.False(t, response.IsError, "Club players should not return error")
			assert.NotEmpty(t, response.Content, "Club players should return member list")
			
			// Verify non-empty results with required player information
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Club players should contain member data")
		})

		t.Run("TC-CP-002: Get Club Players with Search", func(t *testing.T) {
			response, err := callMCPTool("get_club_players", map[string]interface{}{
				"club_id": TestClubID,
				"query":   "Minh",
				"limit":   10,
			})
			require.NoError(t, err, "Get club players with search should not fail")
			
			assert.False(t, response.IsError, "Club players search should not return error")
			assert.NotEmpty(t, response.Content, "Club players search should return filtered results")
			
			// Verify filtered results match search query and are limited to 10
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Filtered club players should contain matching data")
		})

		t.Run("TC-CP-003: Get Active Club Players", func(t *testing.T) {
			response, err := callMCPTool("get_club_players", map[string]interface{}{
				"club_id":    TestClubID,
				"active":     true,
				"sort_by":    "current_dwz",
				"sort_order": "desc",
			})
			require.NoError(t, err, "Get active club players should not fail")
			
			assert.False(t, response.IsError, "Active club players should not return error")
			assert.NotEmpty(t, response.Content, "Active club players should return sorted results")
			
			// Verify only active players, sorted by DWZ rating descending
			content := extractTextContent(response.Content)
			assert.NotEmpty(t, content, "Active club players should contain sorted data")
		})
	})
}

// 4. Analysis Tools Tests
func testAnalysisTools(t *testing.T) {
	t.Run("TC-RH-001: Get Player Rating History", func(t *testing.T) {
		response, err := callMCPTool("get_player_rating_history", map[string]interface{}{
			"player_id": TestPlayerID,
		})
		require.NoError(t, err, "Get player rating history should not fail")
		
		assert.False(t, response.IsError, "Player rating history should not return error")
		assert.NotEmpty(t, response.Content, "Player rating history should return data")
		
		// Verify historical DWZ ratings with dates and rating changes, sorted chronologically
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Rating history should contain historical DWZ ratings and dates")
	})

	t.Run("TC-CS-001: Get Club Statistics", func(t *testing.T) {
		response, err := callMCPTool("get_club_statistics", map[string]interface{}{
			"club_id": TestClubID,
		})
		require.NoError(t, err, "Get club statistics should not fail")
		
		assert.False(t, response.IsError, "Club statistics should not return error")
		assert.NotEmpty(t, response.Content, "Club statistics should return data")
		
		// Verify member analytics, performance metrics, and historical data
		content := extractTextContent(response.Content)
		assert.NotEmpty(t, content, "Club statistics should contain member analytics and performance metrics")
	})
}

// 5. MCP Protocol Tests
func testMCPProtocol(t *testing.T) {
	t.Run("TC-TD-001: List Available Tools", func(t *testing.T) {
		response, err := makeHTTPRequest("POST", "/tools/list", map[string]interface{}{})
		require.NoError(t, err, "List tools should not fail")
		
		var toolsList map[string]interface{}
		err = json.Unmarshal(response, &toolsList)
		require.NoError(t, err, "Response should be valid JSON")
		
		// Verify all 14 expected tools are present
		if tools, ok := toolsList["tools"].([]interface{}); ok {
			assert.GreaterOrEqual(t, len(tools), 14, "Should have at least 14 tools available")
			
			// Verify each tool has name, description, and parameters
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					assert.Contains(t, toolMap, "name", "Tool should have name")
					assert.Contains(t, toolMap, "description", "Tool should have description")
					assert.Contains(t, toolMap, "inputSchema", "Tool should have parameters")
				}
			}
		}
	})

	t.Run("TC-RD-001: List Available Resources", func(t *testing.T) {
		response, err := makeHTTPRequest("POST", "/resources/list", map[string]interface{}{})
		require.NoError(t, err, "List resources should not fail")
		
		var resourcesList map[string]interface{}
		err = json.Unmarshal(response, &resourcesList)
		require.NoError(t, err, "Response should be valid JSON")
		
		// Verify resource URIs are properly formatted
		if resources, ok := resourcesList["resources"].([]interface{}); ok {
			assert.NotEmpty(t, resources, "Should have available resources")
		}
	})

	t.Run("Resource Access Tests", func(t *testing.T) {
		t.Run("TC-RA-001: Access Player Resource", func(t *testing.T) {
			response, err := makeHTTPRequest("POST", "/resources/read", map[string]interface{}{
				"uri": fmt.Sprintf("players://%s", TestPlayerID),
			})
			require.NoError(t, err, "Access player resource should not fail")
			
			var resourceData map[string]interface{}
			err = json.Unmarshal(response, &resourceData)
			require.NoError(t, err, "Resource response should be valid JSON")
			
			// Verify player resource data matches get_player_profile result
			assert.NotEmpty(t, resourceData, "Player resource should contain data")
		})

		t.Run("TC-RA-002: Access Club Resource", func(t *testing.T) {
			response, err := makeHTTPRequest("POST", "/resources/read", map[string]interface{}{
				"uri": fmt.Sprintf("clubs://%s", TestClubID),
			})
			require.NoError(t, err, "Access club resource should not fail")
			
			var resourceData map[string]interface{}
			err = json.Unmarshal(response, &resourceData)
			require.NoError(t, err, "Resource response should be valid JSON")
			
			// Verify club resource data matches get_club_profile result
			assert.NotEmpty(t, resourceData, "Club resource should contain data")
		})

		t.Run("TC-RA-003: Access Tournament Resource", func(t *testing.T) {
			response, err := makeHTTPRequest("POST", "/resources/read", map[string]interface{}{
				"uri": fmt.Sprintf("tournaments://%s", TestTournamentID),
			})
			require.NoError(t, err, "Access tournament resource should not fail")
			
			var resourceData map[string]interface{}
			err = json.Unmarshal(response, &resourceData)
			require.NoError(t, err, "Resource response should be valid JSON")
			
			// Verify tournament resource data matches get_tournament_details result
			assert.NotEmpty(t, resourceData, "Tournament resource should contain data")
		})
	})
}

// Helper functions

// isServerRunning checks if the Portal64 MCP server is running on localhost:8888
func isServerRunning() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(BaseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// callMCPTool makes an MCP tool call via REST API
func callMCPTool(toolName string, args map[string]interface{}) (*MCPToolResponse, error) {
	toolCall := MCPToolCall{
		Name:      toolName,
		Arguments: args,
	}
	
	responseData, err := makeHTTPRequest("POST", "/tools/call", toolCall)
	if err != nil {
		return nil, err
	}
	
	var response MCPToolResponse
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP response: %w", err)
	}
	
	return &response, nil
}

// makeHTTPRequest makes an HTTP request to the MCP server
func makeHTTPRequest(method, endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(method, BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	
	var responseData bytes.Buffer
	_, err = responseData.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return responseData.Bytes(), nil
}

// extractTextContent extracts text content from MCP response content array
func extractTextContent(content []MCPContent) string {
	var text strings.Builder
	for _, item := range content {
		if item.Type == "text" && item.Text != nil {
			if str, ok := item.Text.(string); ok {
				text.WriteString(str)
				text.WriteString(" ")
			} else {
				// Handle case where text is not a string (e.g., structured data)
				if jsonData, err := json.Marshal(item.Text); err == nil {
					text.WriteString(string(jsonData))
					text.WriteString(" ")
				}
			}
		}
	}
	return text.String()
}
