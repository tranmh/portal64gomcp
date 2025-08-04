package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the MCP server
type Config struct {
	API    APIConfig    `mapstructure:"api"`
	MCP    MCPConfig    `mapstructure:"mcp"`
	Logger LoggerConfig `mapstructure:"logging"`
}

// APIConfig holds Portal64 API configuration
type APIConfig struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Port int `mapstructure:"port"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from environment variables and config files
func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/portal64gomcp/")
	viper.AddConfigPath("$HOME/.portal64gomcp")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	}

	// Set defaults
	viper.SetDefault("api.base_url", "http://localhost:8080")
	viper.SetDefault("api.timeout", "30s")
	viper.SetDefault("mcp.port", 3000)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Bind environment variables
	viper.SetEnvPrefix("PORTAL64")
	viper.AutomaticEnv()
	viper.BindEnv("api.base_url", "PORTAL64_API_URL")
	viper.BindEnv("mcp.port", "MCP_SERVER_PORT")
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("api.timeout", "API_TIMEOUT")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.API.BaseURL == "" {
		return fmt.Errorf("api.base_url is required")
	}

	if c.MCP.Port <= 0 || c.MCP.Port > 65535 {
		return fmt.Errorf("mcp.port must be between 1 and 65535")
	}

	if c.API.Timeout <= 0 {
		return fmt.Errorf("api.timeout must be positive")
	}

	return nil
}
