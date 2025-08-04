package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svw-info/portal64gomcp/internal/config"
	"github.com/svw-info/portal64gomcp/test/testutil"
)

func TestMain_Integration(t *testing.T) {
	// Test main application components integration
	
	t.Run("Configuration Loading", func(t *testing.T) {
		// Create a temporary config file
		configContent := `
api:
  base_url: "http://localhost:8080"
  timeout: "30s"
mcp:
  port: 3000
logging:
  level: "debug"
  format: "json"
`
		configFile := testutil.CreateTempConfigFile(t, configContent)
		
		// Test configuration loading
		cfg, err := config.Load(configFile)
		require.NoError(t, err)
		
		assert.Equal(t, "http://localhost:8080", cfg.API.BaseURL)
		assert.Equal(t, 3000, cfg.MCP.Port)
		assert.Equal(t, "debug", cfg.Logger.Level)
		
		// Test configuration validation
		err = cfg.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("Environment Variable Override", func(t *testing.T) {
		// Set environment variables
		originalURL := os.Getenv("PORTAL64_API_URL")
		originalPort := os.Getenv("MCP_SERVER_PORT")
		
		os.Setenv("PORTAL64_API_URL", "http://test.example.com:9000")
		os.Setenv("MCP_SERVER_PORT", "4000")
		
		// Cleanup
		defer func() {
			if originalURL != "" {
				os.Setenv("PORTAL64_API_URL", originalURL)
			} else {
				os.Unsetenv("PORTAL64_API_URL")
			}
			
			if originalPort != "" {
				os.Setenv("MCP_SERVER_PORT", originalPort)
			} else {
				os.Unsetenv("MCP_SERVER_PORT")
			}
		}()
		
		// Load configuration
		cfg, err := config.Load("")
		require.NoError(t, err)
		
		// Verify environment variables take precedence
		assert.Equal(t, "http://test.example.com:9000", cfg.API.BaseURL)
		assert.Equal(t, 4000, cfg.MCP.Port)
	})
}

func TestApplication_EndToEnd(t *testing.T) {
	// End-to-end application testing scenarios
	
	t.Run("JSON Configuration Parsing", func(t *testing.T) {
		// Test various configuration formats and edge cases
		testCases := []struct {
			name     string
			config   string
			valid    bool
			expected map[string]interface{}
		}{
			{
				name: "Valid complete configuration",
				config: `
api:
  base_url: "https://api.portal64.com"
  timeout: "45s"
mcp:
  port: 5000
logging:
  level: "info"
  format: "text"
`,
				valid: true,
				expected: map[string]interface{}{
					"api_url": "https://api.portal64.com",
					"port":    5000,
					"level":   "info",
				},
			},
			{
				name: "Minimal configuration",
				config: `
api:
  base_url: "http://localhost:8080"
`,
				valid: true,
				expected: map[string]interface{}{
					"api_url": "http://localhost:8080",
				},
			},
			{
				name: "Invalid timeout format",
				config: `
api:
  timeout: "invalid-duration"
`,
				valid: false,
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				configFile := testutil.CreateTempConfigFile(t, tc.config)
				
				cfg, err := config.Load(configFile)
				
				if tc.valid {
					require.NoError(t, err)
					assert.NotNil(t, cfg)
					
					// Verify expected values if provided
					if apiURL, ok := tc.expected["api_url"]; ok {
						assert.Equal(t, apiURL, cfg.API.BaseURL)
					}
					if port, ok := tc.expected["port"]; ok {
						assert.Equal(t, port, cfg.MCP.Port)
					}
					if level, ok := tc.expected["level"]; ok {
						assert.Equal(t, level, cfg.Logger.Level)
					}
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

func TestApplication_ErrorScenarios(t *testing.T) {
	// Test various error conditions and recovery
	
	t.Run("Invalid Configuration Handling", func(t *testing.T) {
		// Test with completely invalid YAML
		invalidConfig := `
this is not valid yaml
  - broken structure
    missing brackets: [
`
		configFile := testutil.CreateTempConfigFile(t, invalidConfig)
		
		_, err := config.Load(configFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error reading config file")
	})
	
	t.Run("Missing Required Fields", func(t *testing.T) {
		// Test configuration with missing required fields
		cfg := &config.Config{
			API: config.APIConfig{
				BaseURL: "", // Empty required field
			},
			MCP: config.MCPConfig{
				Port: 3000,
			},
		}
		
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api.base_url is required")
	})
	
	t.Run("Port Range Validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			port        int
			shouldError bool
		}{
			{"Valid port 3000", 3000, false},
			{"Valid port 8080", 8080, false},
			{"Invalid port 0", 0, true},
			{"Invalid negative port", -1, true},
			{"Invalid high port", 70000, true},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &config.Config{
					API: config.APIConfig{
						BaseURL: "http://localhost:8080",
						Timeout: 30 * time.Second,
					},
					MCP: config.MCPConfig{
						Port: tc.port,
					},
				}
				
				err := cfg.Validate()
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestApplication_Performance(t *testing.T) {
	// Performance and load testing scenarios
	
	t.Run("Configuration Loading Performance", func(t *testing.T) {
		// Test configuration loading performance
		configContent := `
api:
  base_url: "http://localhost:8080"
  timeout: "30s"
mcp:
  port: 3000
logging:
  level: "info"
  format: "json"
`
		configFile := testutil.CreateTempConfigFile(t, configContent)
		
		// Load configuration multiple times to test performance
		const iterations = 100
		for i := 0; i < iterations; i++ {
			cfg, err := config.Load(configFile)
			require.NoError(t, err)
			assert.NotNil(t, cfg)
		}
	})
	
	t.Run("JSON Marshaling Performance", func(t *testing.T) {
		// Test JSON marshaling performance for large configurations
		largeConfig := map[string]interface{}{
			"api": map[string]interface{}{
				"base_url": "http://localhost:8080",
				"timeout":  "30s",
			},
			"mcp": map[string]interface{}{
				"port": 3000,
			},
			"logging": map[string]interface{}{
				"level":  "info",
				"format": "json",
			},
			"metadata": generateLargeMetadata(1000),
		}
		
		// Test marshaling performance
		const iterations = 100
		for i := 0; i < iterations; i++ {
			_, err := json.Marshal(largeConfig)
			require.NoError(t, err)
		}
	})
}

// Helper function to generate test metadata
func generateLargeMetadata(size int) map[string]interface{} {
	metadata := make(map[string]interface{})
	for i := 0; i < size; i++ {
		key := "key_" + string(rune(i))
		metadata[key] = map[string]interface{}{
			"value":       i,
			"description": "Test metadata entry " + string(rune(i)),
			"timestamp":   "2024-01-01T00:00:00Z",
		}
	}
	return metadata
}

// Benchmark tests
func BenchmarkConfigurationLoading(b *testing.B) {
	configContent := `
api:
  base_url: "http://localhost:8080"
  timeout: "30s"
mcp:
  port: 3000
logging:
  level: "info"
  format: "json"
`
	configFile := testutil.CreateTempConfigFile(&testing.T{}, configContent)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := config.Load(configFile)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkConfigurationValidation(b *testing.B) {
	cfg := &config.Config{
		API: config.APIConfig{
			BaseURL: "http://localhost:8080",
		},
		MCP: config.MCPConfig{
			Port: 3000,
		},
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := cfg.Validate()
			if err != nil {
				b.Error(err)
			}
		}
	})
}
