package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svw-info/portal64gomcp/test/testutil"
)

func TestLoad_DefaultConfiguration(t *testing.T) {
	// Clear any existing environment variables
	clearEnvVars(t)
	
	config, err := Load("")
	require.NoError(t, err)
	
	// Verify default values
	assert.Equal(t, "http://localhost:8080", config.API.BaseURL)
	assert.Equal(t, 30*time.Second, config.API.Timeout)
	assert.Equal(t, 3000, config.MCP.Port)
	assert.Equal(t, "info", config.Logger.Level)
	assert.Equal(t, "json", config.Logger.Format)
}

func TestLoad_FromConfigFile(t *testing.T) {
	clearEnvVars(t)
	
	configContent := `
api:
  base_url: "http://test.example.com:9000"
  timeout: "45s"
mcp:
  port: 4000
logging:
  level: "debug"
  format: "text"
`
	
	configFile := testutil.CreateTempConfigFile(t, configContent)
	
	config, err := Load(configFile)
	require.NoError(t, err)
	
	assert.Equal(t, "http://test.example.com:9000", config.API.BaseURL)
	assert.Equal(t, 45*time.Second, config.API.Timeout)
	assert.Equal(t, 4000, config.MCP.Port)
	assert.Equal(t, "debug", config.Logger.Level)
	assert.Equal(t, "text", config.Logger.Format)
}

func TestLoad_FromEnvironmentVariables(t *testing.T) {
	clearEnvVars(t)
	
	// Set environment variables
	setEnvVar(t, "PORTAL64_API_URL", "http://env.example.com:7000")
	setEnvVar(t, "MCP_SERVER_PORT", "5000")
	setEnvVar(t, "LOG_LEVEL", "warn")
	setEnvVar(t, "API_TIMEOUT", "60s")
	
	config, err := Load("")
	require.NoError(t, err)
	
	assert.Equal(t, "http://env.example.com:7000", config.API.BaseURL)
	assert.Equal(t, 60*time.Second, config.API.Timeout)
	assert.Equal(t, 5000, config.MCP.Port)
	assert.Equal(t, "warn", config.Logger.Level)
}

func TestLoad_EnvironmentOverridesConfigFile(t *testing.T) {
	clearEnvVars(t)
	
	configContent := `
api:
  base_url: "http://file.example.com:8000"
  timeout: "30s"
mcp:
  port: 3000
`
	
	configFile := testutil.CreateTempConfigFile(t, configContent)
	
	// Environment variable should override config file
	setEnvVar(t, "PORTAL64_API_URL", "http://env.example.com:9000")
	setEnvVar(t, "MCP_SERVER_PORT", "4000")
	
	config, err := Load(configFile)
	require.NoError(t, err)
	
	assert.Equal(t, "http://env.example.com:9000", config.API.BaseURL) // From env
	assert.Equal(t, 4000, config.MCP.Port)                            // From env
	assert.Equal(t, 30*time.Second, config.API.Timeout)               // From file
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	clearEnvVars(t)
	
	invalidConfig := `
invalid yaml content
  - missing structure
`
	
	configFile := testutil.CreateTempConfigFile(t, invalidConfig)
	
	_, err := Load(configFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading config file")
}

func TestLoad_NonExistentConfigFile(t *testing.T) {
	clearEnvVars(t)
	
	// Use a non-existent file path that's more realistic for the OS
	nonExistentPath := filepath.Join(os.TempDir(), "non_existent_config_file.yaml")
	
	// When a specific config file is requested but doesn't exist, it should return an error
	_, err := Load(nonExistentPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading config file")
}

func TestValidate_ValidConfiguration(t *testing.T) {
	config := &Config{
		API: APIConfig{
			BaseURL: "http://localhost:8080",
			Timeout: 30 * time.Second,
		},
		MCP: MCPConfig{
			Port: 3000,
		},
		Logger: LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}
	
	err := config.Validate()
	assert.NoError(t, err)
}

func TestValidate_EmptyBaseURL(t *testing.T) {
	config := &Config{
		API: APIConfig{
			BaseURL: "",
			Timeout: 30 * time.Second,
		},
		MCP: MCPConfig{
			Port: 3000,
		},
	}
	
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api.base_url is required")
}

func TestValidate_InvalidPort(t *testing.T) {
	testCases := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"port too high", 65536},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				API: APIConfig{
					BaseURL: "http://localhost:8080",
					Timeout: 30 * time.Second,
				},
				MCP: MCPConfig{
					Port: tc.port,
				},
			}
			
			err := config.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mcp.port must be between 1 and 65535")
		})
	}
}

func TestValidate_InvalidTimeout(t *testing.T) {
	config := &Config{
		API: APIConfig{
			BaseURL: "http://localhost:8080",
			Timeout: -1 * time.Second,
		},
		MCP: MCPConfig{
			Port: 3000,
		},
	}
	
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api.timeout must be positive")
}

func TestLoad_InvalidTimeout(t *testing.T) {
	clearEnvVars(t)
	
	configContent := `
api:
  timeout: "invalid-duration"
`
	
	configFile := testutil.CreateTempConfigFile(t, configContent)
	
	_, err := Load(configFile)
	assert.Error(t, err)
}

// Helper functions

func clearEnvVars(t *testing.T) {
	envVars := []string{
		"PORTAL64_API_URL",
		"MCP_SERVER_PORT", 
		"LOG_LEVEL",
		"API_TIMEOUT",
	}
	
	for _, env := range envVars {
		original := os.Getenv(env)
		os.Unsetenv(env)
		
		// Restore original value after test
		if original != "" {
			t.Cleanup(func() {
				os.Setenv(env, original)
			})
		}
	}
}

func setEnvVar(t *testing.T, key, value string) {
	original := os.Getenv(key)
	os.Setenv(key, value)
	
	// Restore original value after test
	t.Cleanup(func() {
		if original != "" {
			os.Setenv(key, original)
		} else {
			os.Unsetenv(key)
		}
	})
}
