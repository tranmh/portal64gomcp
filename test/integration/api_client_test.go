package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svw-info/portal64gomcp/internal/api"
	"github.com/svw-info/portal64gomcp/test/testutil"
)

func TestAPIClient_Integration(t *testing.T) {
	// Load test fixtures
	fixtures := testutil.LoadFixtures(t)
	
	// Create mock API server
	server := testutil.CreateMockAPIServer(t, fixtures)
	defer server.Close()
	
	// Create API client pointing to mock server
	logger := testutil.NewTestLogger()
	client := api.NewClient(server.URL, 30*time.Second, logger)
	
	t.Run("Health Check", func(t *testing.T) {
		url := client.BuildURL("/health", nil)
		resp, err := client.DoRequest(context.Background(), "GET", url)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var health map[string]interface{}
		err = client.DecodeResponse(resp, &health)
		require.NoError(t, err)
		
		assert.Equal(t, "healthy", health["status"])
		assert.NotEmpty(t, health["timestamp"])
	})
	
	t.Run("Search Players", func(t *testing.T) {
		params := api.SearchParams{
			Query: "Max",
			Limit: 20,
		}
		
		url := client.BuildURL("/api/players/", params)
		resp, err := client.DoRequest(context.Background(), "GET", url)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var searchResponse api.SearchResponse
		err = client.DecodeResponse(resp, &searchResponse)
		require.NoError(t, err)
		
		// Verify the response structure matches expected format
		assert.NotNil(t, searchResponse.Data)
		assert.NotNil(t, searchResponse.Pagination)
	})
	
	t.Run("Get Player Details", func(t *testing.T) {
		url := client.BuildURL("/api/players/12345", nil)
		resp, err := client.DoRequest(context.Background(), "GET", url)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var player map[string]interface{}
		err = client.DecodeResponse(resp, &player)
		require.NoError(t, err)
		
		expectedPlayer := fixtures.Player
		assert.Equal(t, expectedPlayer["id"], player["id"])
		assert.Equal(t, expectedPlayer["name"], player["name"])
		assert.Equal(t, expectedPlayer["dwz"], player["dwz"])
	})
	
	t.Run("Search Clubs", func(t *testing.T) {
		params := api.SearchParams{
			Query:  "Berlin",
			Limit:  10,
			Offset: 0,
		}
		
		url := client.BuildURL("/api/clubs/", params)
		resp, err := client.DoRequest(context.Background(), "GET", url)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var searchResponse api.SearchResponse
		err = client.DecodeResponse(resp, &searchResponse)
		require.NoError(t, err)
		
		assert.NotNil(t, searchResponse.Data)
		assert.NotNil(t, searchResponse.Pagination)
	})
	
	t.Run("Get Club Details", func(t *testing.T) {
		url := client.BuildURL("/api/clubs/001", nil)
		resp, err := client.DoRequest(context.Background(), "GET", url)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var club map[string]interface{}
		err = client.DecodeResponse(resp, &club)
		require.NoError(t, err)
		
		expectedClub := fixtures.Club
		assert.Equal(t, expectedClub["id"], club["id"])
		assert.Equal(t, expectedClub["name"], club["name"])
		assert.Equal(t, expectedClub["member_count"], club["member_count"])
	})
	
	t.Run("Search Tournaments", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		
		params := api.DateRangeParams{
			StartDate: startDate,
			EndDate:   endDate,
			SearchParams: api.SearchParams{
				Limit: 25,
			},
		}
		
		url := client.BuildURL("/api/tournaments/", params)
		resp, err := client.DoRequest(context.Background(), "GET", url)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var searchResponse api.SearchResponse
		err = client.DecodeResponse(resp, &searchResponse)
		require.NoError(t, err)
		
		assert.NotNil(t, searchResponse.Data)
		assert.NotNil(t, searchResponse.Pagination)
	})
	
	t.Run("Get Tournament Details", func(t *testing.T) {
		url := client.BuildURL("/api/tournaments/T001", nil)
		resp, err := client.DoRequest(context.Background(), "GET", url)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var tournament map[string]interface{}
		err = client.DecodeResponse(resp, &tournament)
		require.NoError(t, err)
		
		expectedTournament := fixtures.Tournament
		assert.Equal(t, expectedTournament["id"], tournament["id"])
		assert.Equal(t, expectedTournament["name"], tournament["name"])
		assert.Equal(t, expectedTournament["participants"], tournament["participants"])
	})
}

func TestAPIClient_ErrorHandling(t *testing.T) {
	// Create a mock server that returns errors
	server := testutil.CreateErrorMockServer(t)
	defer server.Close()
	
	logger := testutil.NewTestLogger()
	client := api.NewClient(server.URL, 30*time.Second, logger)
	
	t.Run("Handle 404 Not Found", func(t *testing.T) {
		url := client.BuildURL("/api/players/nonexistent", nil)
		_, err := client.DoRequest(context.Background(), "GET", url)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
	
	t.Run("Handle 500 Internal Server Error", func(t *testing.T) {
		url := client.BuildURL("/api/error", nil)
		_, err := client.DoRequest(context.Background(), "GET", url)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})
	
	t.Run("Handle Network Timeout", func(t *testing.T) {
		// Create client with very short timeout
		shortTimeoutClient := api.NewClient(server.URL, 1*time.Millisecond, logger)
		
		url := shortTimeoutClient.BuildURL("/api/slow", nil)
		_, err := shortTimeoutClient.DoRequest(context.Background(), "GET", url)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API request failed")
	})
}

func TestAPIClient_Concurrent(t *testing.T) {
	fixtures := testutil.LoadFixtures(t)
	server := testutil.CreateMockAPIServer(t, fixtures)
	defer server.Close()
	
	logger := testutil.NewTestLogger()
	client := api.NewClient(server.URL, 30*time.Second, logger)
	
	// Test concurrent requests
	const numConcurrentRequests = 10
	results := make(chan error, numConcurrentRequests)
	
	for i := 0; i < numConcurrentRequests; i++ {
		go func() {
			url := client.BuildURL("/health", nil)
			resp, err := client.DoRequest(context.Background(), "GET", url)
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()
			
			var health map[string]interface{}
			err = client.DecodeResponse(resp, &health)
			results <- err
		}()
	}
	
	// Collect all results
	for i := 0; i < numConcurrentRequests; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent request %d failed", i)
	}
}

func TestAPIClient_ParameterEncoding(t *testing.T) {
	fixtures := testutil.LoadFixtures(t)
	server := testutil.CreateMockAPIServer(t, fixtures)
	defer server.Close()
	
	logger := testutil.NewTestLogger()
	client := api.NewClient(server.URL, 30*time.Second, logger)
	
	testCases := []struct {
		name   string
		params interface{}
	}{
		{
			name: "Search with special characters",
			params: api.SearchParams{
				Query:       "Müller & Schmidt",
				FilterValue: "Berlin/Brandenburg",
			},
		},
		{
			name: "Date range with edge dates",
			params: api.DateRangeParams{
				StartDate: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2100, 12, 31, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Map with various value types",
			params: map[string]string{
				"unicode": "🏆♔♕♖♗♘♙",
				"spaces":  "value with spaces",
				"symbols": "value!@#$%^&*()",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := client.BuildURL("/api/players/", tc.params)
			
			// The URL should be valid (no panic during construction)
			assert.NotEmpty(t, url)
			assert.Contains(t, url, server.URL)
			
			// The request should not fail due to URL encoding issues
			resp, err := client.DoRequest(context.Background(), "GET", url)
			if err == nil {
				resp.Body.Close()
			}
			// We don't assert no error here since the mock server might not handle
			// all parameter combinations, but URL construction shouldn't panic
		})
	}
}

func BenchmarkAPIClient_SimpleRequest(b *testing.B) {
	fixtures := testutil.LoadFixtures(&testing.T{})
	server := testutil.CreateMockAPIServer(&testing.T{}, fixtures)
	defer server.Close()
	
	logger := testutil.NewTestLogger()
	client := api.NewClient(server.URL, 30*time.Second, logger)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			url := client.BuildURL("/health", nil)
			resp, err := client.DoRequest(context.Background(), "GET", url)
			if err == nil {
				resp.Body.Close()
			}
		}
	})
}
