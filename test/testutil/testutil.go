package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestFixtures holds all test fixture data
type TestFixtures struct {
	Player           map[string]interface{} `json:"player"`
	PlayersSearch    map[string]interface{} `json:"players_search"`
	Club             map[string]interface{} `json:"club"`
	ClubsSearch      map[string]interface{} `json:"clubs_search"`
	Tournament       map[string]interface{} `json:"tournament"`
	TournamentsSearch map[string]interface{} `json:"tournaments_search"`
}

// LoadFixtures loads test fixtures from JSON file
func LoadFixtures(t *testing.T) *TestFixtures {
	// Get the project root directory
	wd, err := os.Getwd()
	require.NoError(t, err)
	
	// Navigate to project root (assuming we're in a subdirectory)
	for !fileExists(filepath.Join(wd, "go.mod")) {
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("Could not find project root (go.mod not found)")
		}
		wd = parent
	}
	
	fixturesPath := filepath.Join(wd, "test", "fixtures", "api_responses.json")
	data, err := os.ReadFile(fixturesPath)
	require.NoError(t, err, "Failed to read fixtures file")
	
	var fixtures TestFixtures
	err = json.Unmarshal(data, &fixtures)
	require.NoError(t, err, "Failed to unmarshal fixtures")
	
	return &fixtures
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// CreateMockAPIServer creates a mock HTTP server for Portal64 API
func CreateMockAPIServer(t *testing.T, fixtures *TestFixtures) *httptest.Server {
	mux := http.NewServeMux()
	
	// Health endpoints (both versioned and non-versioned)
	healthResponse := map[string]interface{}{
		"status": "healthy",
		"timestamp": "2024-01-01T12:00:00Z",
	}
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, healthResponse)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, healthResponse)
	})
	
	// Players endpoints (both versioned and non-versioned)
	playersSearchHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			writeJSONResponse(w, convertSearchResponse(fixtures.PlayersSearch))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/api/v1/players", playersSearchHandler)
	mux.HandleFunc("/api/players/", playersSearchHandler)
	
	playersDetailHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/api/v1/players/12345" || r.URL.Path == "/api/players/12345" {
				writeJSONResponse(w, fixtures.Player)
			} else {
				writeJSONResponse(w, convertSearchResponse(fixtures.PlayersSearch))
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/api/v1/players/", playersDetailHandler)
	mux.HandleFunc("/api/players/12345", func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, fixtures.Player)
	})
	
	// Clubs endpoints (both versioned and non-versioned)
	clubsSearchHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			writeJSONResponse(w, convertSearchResponse(fixtures.ClubsSearch))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/api/v1/clubs", clubsSearchHandler)
	mux.HandleFunc("/api/clubs/", clubsSearchHandler)
	
	clubsDetailHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/api/v1/clubs/001/profile" || r.URL.Path == "/api/clubs/001" {
				writeJSONResponse(w, map[string]interface{}{
					"club": fixtures.Club,
					"players": []interface{}{},
					"contact": map[string]interface{}{},
					"teams": []interface{}{},
					"rating_stats": map[string]interface{}{},
					"recent_tournaments": []interface{}{},
					"player_count": 85,
					"active_player_count": 65,
					"tournament_count": 12,
				})
			} else {
				writeJSONResponse(w, convertSearchResponse(fixtures.ClubsSearch))
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/api/v1/clubs/", clubsDetailHandler)
	mux.HandleFunc("/api/clubs/001", func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, fixtures.Club)
	})
	
	// Tournaments endpoints (both versioned and non-versioned)
	tournamentsSearchHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			writeJSONResponse(w, convertSearchResponse(fixtures.TournamentsSearch))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/api/v1/tournaments", tournamentsSearchHandler)
	mux.HandleFunc("/api/tournaments/", tournamentsSearchHandler)
	
	tournamentsDetailHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/api/v1/tournaments/T001" || r.URL.Path == "/api/tournaments/T001" {
				writeJSONResponse(w, fixtures.Tournament)
			} else {
				writeJSONResponse(w, convertSearchResponse(fixtures.TournamentsSearch))
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/api/v1/tournaments/", tournamentsDetailHandler)
	mux.HandleFunc("/api/tournaments/T001", func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, fixtures.Tournament)
	})
	
	return httptest.NewServer(mux)
}

// convertSearchResponse converts fixture format to API response format
func convertSearchResponse(fixture map[string]interface{}) map[string]interface{} {
	if results, ok := fixture["results"]; ok {
		return map[string]interface{}{
			"data":       results,
			"pagination": fixture["pagination"],
		}
	}
	return fixture
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode JSON: %v", err), http.StatusInternalServerError)
	}
}

// NewTestLogger creates a logger for testing
func NewTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	return logger
}

// CreateTempConfigFile creates a temporary config file for testing
func CreateTempConfigFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	
	err = tmpFile.Close()
	require.NoError(t, err)
	
	// Clean up the file after the test
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})
	
	return tmpFile.Name()
}

// AssertJSONEqual compares two JSON objects for equality
func AssertJSONEqual(t *testing.T, expected, actual interface{}) {
	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	require.NoError(t, err)
	
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)
	
	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}
// CreateErrorMockServer creates a mock HTTP server that returns various error responses
func CreateErrorMockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	
	// 404 Not Found - both versioned and non-versioned paths
	notFoundHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]interface{}{
			"message": "Player not found",
			"code":    "PLAYER_NOT_FOUND",
		})
	}
	mux.HandleFunc("/api/v1/players/nonexistent", notFoundHandler)
	mux.HandleFunc("/api/players/nonexistent", notFoundHandler)
	
	// 500 Internal Server Error - both versioned and non-versioned paths
	errorHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"message": "Internal server error",
			"code":    "INTERNAL_ERROR",
		})
	}
	mux.HandleFunc("/api/v1/error", errorHandler)
	mux.HandleFunc("/api/error", errorHandler)
	
	// Slow endpoint for timeout testing - all health endpoints should be slow
	slowHandler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		writeJSONResponse(w, map[string]interface{}{
			"message": "slow response",
		})
	}
	mux.HandleFunc("/api/v1/health", slowHandler)
	mux.HandleFunc("/health", slowHandler)
	mux.HandleFunc("/api/v1/slow", slowHandler)
	mux.HandleFunc("/api/slow", slowHandler)
	
	// Add catch-all handler that returns 404 for any unmatched endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]interface{}{
			"message": "Endpoint not found",
			"code":    "NOT_FOUND",
		})
	})
	
	// 400 Bad Request
	mux.HandleFunc("/api/v1/bad-request", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"message": "Invalid parameters",
			"code":    "BAD_REQUEST",
		})
	})
	
	// 403 Forbidden
	mux.HandleFunc("/api/v1/forbidden", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		writeJSONResponse(w, map[string]interface{}{
			"message": "Access denied",
			"code":    "FORBIDDEN",
		})
	})
	
	return httptest.NewServer(mux)
}
