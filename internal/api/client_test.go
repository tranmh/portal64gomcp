package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svw-info/portal64gomcp/test/testutil"
)

func TestNewClient(t *testing.T) {
	logger := testutil.NewTestLogger()
	timeout := 30 * time.Second
	baseURL := "http://localhost:8080"
	
	client := NewClient(baseURL, timeout, logger)
	
	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.baseURL)
	assert.Equal(t, timeout, client.httpClient.Timeout)
	assert.Equal(t, logger, client.logger)
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	logger := testutil.NewTestLogger()
	
	client := NewClient("http://localhost:8080/", 30*time.Second, logger)
	
	assert.Equal(t, "http://localhost:8080", client.baseURL)
}

func TestClient_BuildURL(t *testing.T) {
	client := createTestClient()
	
	t.Run("No params", func(t *testing.T) {
		url := client.BuildURL("/api/players", nil)
		assert.Equal(t, "http://localhost:8080/api/players", url)
	})
	
	t.Run("With search params", func(t *testing.T) {
		params := SearchParams{
			Query:       "Max",
			Limit:       20,
			Offset:      10,
			SortBy:      "name",
			SortOrder:   "asc",
		}
		
		url := client.BuildURL("/api/players", params)
		
		// Check that URL contains expected parameters
		assert.Contains(t, url, "query=Max")
		assert.Contains(t, url, "limit=20")
		assert.Contains(t, url, "offset=10")
		assert.Contains(t, url, "sort_by=name")
		assert.Contains(t, url, "sort_order=asc")
	})
	
	t.Run("With date range params", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		
		params := DateRangeParams{
			StartDate: startDate,
			EndDate:   endDate,
			SearchParams: SearchParams{
				Limit: 50,
			},
		}
		
		url := client.BuildURL("/api/tournaments", params)
		
		assert.Contains(t, url, "start_date=2024-01-01")
		assert.Contains(t, url, "end_date=2024-12-31")
		assert.Contains(t, url, "limit=50")
	})
	
	t.Run("With map params", func(t *testing.T) {
		params := map[string]string{
			"region": "Berlin",
			"active": "true",
		}
		
		url := client.BuildURL("/api/clubs", params)
		
		assert.Contains(t, url, "active=true")
		assert.Contains(t, url, "region=Berlin")
	})
}

func TestClient_PublicMethods(t *testing.T) {
	fixtures := testutil.LoadFixtures(t)
	server := testutil.CreateMockAPIServer(t, fixtures)
	defer server.Close()
	
	client := createTestClientWithURL(server.URL)
	ctx := context.Background()
	
	t.Run("Health check", func(t *testing.T) {
		health, err := client.Health(ctx)
		require.NoError(t, err)
		assert.NotNil(t, health)
	})
	
	t.Run("Search players", func(t *testing.T) {
		params := SearchParams{
			Query: "Max",
			Limit: 20,
		}
		
		result, err := client.SearchPlayers(ctx, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Data)
		assert.NotNil(t, result.Pagination)
	})
	
	t.Run("Get player profile", func(t *testing.T) {
		player, err := client.GetPlayerProfile(ctx, "12345")
		require.NoError(t, err)
		assert.NotNil(t, player)
		assert.Equal(t, "12345", player.ID)
	})
	
	t.Run("Search clubs", func(t *testing.T) {
		params := SearchParams{
			Query:  "Berlin",
			Limit:  10,
			Offset: 0,
		}
		
		result, err := client.SearchClubs(ctx, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Data)
	})
	
	t.Run("Get club profile", func(t *testing.T) {
		club, err := client.GetClubProfile(ctx, "001")
		require.NoError(t, err)
		assert.NotNil(t, club)
		assert.NotNil(t, club.Club)
	})
	
	t.Run("Search tournaments", func(t *testing.T) {
		params := SearchParams{
			Limit: 25,
		}
		
		result, err := client.SearchTournaments(ctx, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Data)
	})
}

func TestClient_ErrorHandling(t *testing.T) {
	server := testutil.CreateErrorMockServer(t)
	defer server.Close()
	
	client := createTestClientWithURL(server.URL)
	ctx := context.Background()
	
	t.Run("Handle 404 Not Found", func(t *testing.T) {
		_, err := client.GetPlayerProfile(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
	
	t.Run("Handle timeout", func(t *testing.T) {
		// Create client with very short timeout
		logger := testutil.NewTestLogger()
		shortTimeoutClient := NewClient(server.URL, 1*time.Millisecond, logger)
		
		_, err := shortTimeoutClient.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API request failed")
	})
}

func TestClient_ConcurrentRequests(t *testing.T) {
	fixtures := testutil.LoadFixtures(t)
	server := testutil.CreateMockAPIServer(t, fixtures)
	defer server.Close()
	
	client := createTestClientWithURL(server.URL)
	ctx := context.Background()
	
	// Test concurrent requests
	const numConcurrentRequests = 10
	results := make(chan error, numConcurrentRequests)
	
	for i := 0; i < numConcurrentRequests; i++ {
		go func() {
			_, err := client.Health(ctx)
			results <- err
		}()
	}
	
	// Collect all results
	for i := 0; i < numConcurrentRequests; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent request %d failed", i)
	}
}

func TestClient_RequestTimeout(t *testing.T) {
	// Create a server that delays response longer than client timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()
	
	// Create client with very short timeout
	logger := testutil.NewTestLogger()
	client := NewClient(server.URL, 10*time.Millisecond, logger)
	
	_, err := client.Health(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed")
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()
	
	client := createTestClientWithURL(server.URL)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	_, err := client.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed")
}

func BenchmarkClient_HealthCheck(b *testing.B) {
	fixtures := testutil.LoadFixtures(&testing.T{})
	server := testutil.CreateMockAPIServer(&testing.T{}, fixtures)
	defer server.Close()
	
	client := createTestClientWithURL(server.URL)
	ctx := context.Background()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.Health(ctx)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// Helper functions

func createTestClient() *Client {
	return createTestClientWithURL("http://localhost:8080")
}

func createTestClientWithURL(baseURL string) *Client {
	logger := testutil.NewTestLogger()
	return NewClient(baseURL, 30*time.Second, logger)
}
