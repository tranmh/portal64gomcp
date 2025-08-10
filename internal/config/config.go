package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the MCP server
type Config struct {
	API         APIConfig         `mapstructure:"api"`
	MCP         MCPConfig         `mapstructure:"mcp"`
	Logger      LoggerConfig      `mapstructure:"logging"`
	Development DevelopmentConfig `mapstructure:"development"`
}

// APIConfig holds Portal64 API configuration
type APIConfig struct {
	BaseURL string        `mapstructure:"base_url" json:"base_url"`
	Timeout time.Duration `mapstructure:"timeout" json:"timeout"`
	SSL     APISSLConfig  `mapstructure:"ssl" json:"ssl"`
}

// APISSLConfig holds API client SSL configuration
type APISSLConfig struct {
	Verify               bool   `mapstructure:"verify"`
	CAFile               string `mapstructure:"ca_file"`
	ClientCert           string `mapstructure:"client_cert"`
	ClientKey            string `mapstructure:"client_key"`
	InsecureSkipVerify   bool   `mapstructure:"insecure_skip_verify"`
}

// MCPConfig holds MCP server configuration  
type MCPConfig struct {
	Port     int           `mapstructure:"port"`
	Mode     string        `mapstructure:"mode"`     // "stdio", "http", or "both"
	HTTPPort int           `mapstructure:"http_port"`
	SSL      MCPSSLConfig  `mapstructure:"ssl"`
}

// MCPSSLConfig holds MCP server SSL configuration
type MCPSSLConfig struct {
	Enabled              bool     `mapstructure:"enabled"`
	CertFile             string   `mapstructure:"cert_file"`
	KeyFile              string   `mapstructure:"key_file"`
	CAFile               string   `mapstructure:"ca_file"`
	MinVersion           string   `mapstructure:"min_version"`
	MaxVersion           string   `mapstructure:"max_version"`
	CipherSuites         []string `mapstructure:"cipher_suites"`
	RequireClientCert    bool     `mapstructure:"require_client_cert"`
	HSTSMaxAge           int64    `mapstructure:"hsts_max_age"`
	AutoRedirectHTTP     bool     `mapstructure:"auto_redirect_http"`
	AutoGenerateCerts    bool     `mapstructure:"auto_generate_certs"`
	AutoCertHosts        []string `mapstructure:"auto_cert_hosts"`
}

// DevelopmentConfig holds development-specific overrides
type DevelopmentConfig struct {
	MCP MCPConfig `mapstructure:"mcp"`
	API APIConfig `mapstructure:"api"`
}

// LoggerConfig holds enhanced logging configuration
type LoggerConfig struct {
	Level      string           `mapstructure:"level"`
	Format     string           `mapstructure:"format"`
	Console    ConsoleConfig    `mapstructure:"console"`
	File       FileConfig       `mapstructure:"file"`
	Rotation   RotationConfig   `mapstructure:"rotation"`
	Separation SeparationConfig `mapstructure:"separation"`
	Async      AsyncConfig      `mapstructure:"async"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
}

// ConsoleConfig holds console output configuration
type ConsoleConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	ForceColors bool `mapstructure:"force_colors"`
}

// FileConfig holds file output configuration
type FileConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	BasePath string `mapstructure:"base_path"`
}

// RotationConfig holds log rotation configuration
type RotationConfig struct {
	MaxSize       int  `mapstructure:"max_size"`        // MB
	MaxAge        int  `mapstructure:"max_age"`         // days
	MaxBackups    int  `mapstructure:"max_backups"`     // number of files
	Compress      bool `mapstructure:"compress"`        // compress old files
	CompressAfter int  `mapstructure:"compress_after"`  // days before compression
}

// SeparationConfig holds log separation configuration
type SeparationConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	AccessLog  bool `mapstructure:"access_log"`
	ErrorLog   bool `mapstructure:"error_log"`
	MetricsLog bool `mapstructure:"metrics_log"`
}

// AsyncConfig holds async logging configuration
type AsyncConfig struct {
	Enabled         bool          `mapstructure:"enabled" json:"enabled"`
	BufferSize      int           `mapstructure:"buffer_size" json:"buffer_size"`
	FlushInterval   time.Duration `mapstructure:"flush_interval" json:"flush_interval"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	IncludeCaller     bool `mapstructure:"include_caller"`
	IncludeRequestID  bool `mapstructure:"include_request_id"`
	IncludeDuration   bool `mapstructure:"include_duration"`
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

	// Set SSL-enhanced defaults
	setDefaults()

	// Bind environment variables
	bindEnvVars()

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

	// Apply development overrides if in development mode
	if isDevelopment() {
		applyDevelopmentOverrides(&config)
	}

	return &config, nil
}

// setDefaults sets all configuration defaults
func setDefaults() {
	// API defaults
	viper.SetDefault("api.base_url", "https://localhost:8443")
	viper.SetDefault("api.timeout", "30s")
	viper.SetDefault("api.ssl.verify", true)
	viper.SetDefault("api.ssl.ca_file", "")
	viper.SetDefault("api.ssl.client_cert", "")
	viper.SetDefault("api.ssl.client_key", "")
	viper.SetDefault("api.ssl.insecure_skip_verify", false)

	// MCP defaults
	viper.SetDefault("mcp.port", 3000)
	viper.SetDefault("mcp.mode", "stdio")
	viper.SetDefault("mcp.http_port", 8888)
	viper.SetDefault("mcp.ssl.enabled", true)
	viper.SetDefault("mcp.ssl.cert_file", "certs/server.crt")
	viper.SetDefault("mcp.ssl.key_file", "certs/server.key")
	viper.SetDefault("mcp.ssl.ca_file", "")
	viper.SetDefault("mcp.ssl.min_version", "1.2")
	viper.SetDefault("mcp.ssl.max_version", "1.3")
	viper.SetDefault("mcp.ssl.cipher_suites", []string{})
	viper.SetDefault("mcp.ssl.require_client_cert", false)
	viper.SetDefault("mcp.ssl.hsts_max_age", 31536000)
	viper.SetDefault("mcp.ssl.auto_redirect_http", false)
	viper.SetDefault("mcp.ssl.auto_generate_certs", true)
	viper.SetDefault("mcp.ssl.auto_cert_hosts", []string{"localhost", "127.0.0.1"})

	// Logging defaults
	// Enhanced logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	
	// Console output defaults
	viper.SetDefault("logging.console.enabled", true)
	viper.SetDefault("logging.console.force_colors", false)
	
	// File output defaults
	viper.SetDefault("logging.file.enabled", true)
	viper.SetDefault("logging.file.base_path", "logs")
	
	// Rotation defaults
	viper.SetDefault("logging.rotation.max_size", 100)      // 100MB
	viper.SetDefault("logging.rotation.max_age", 1)         // 1 day
	viper.SetDefault("logging.rotation.max_backups", 30)    // 30 files
	viper.SetDefault("logging.rotation.compress", true)     // compress old files
	viper.SetDefault("logging.rotation.compress_after", 1)  // compress after 1 day
	
	// Separation defaults
	viper.SetDefault("logging.separation.enabled", true)
	viper.SetDefault("logging.separation.access_log", true)
	viper.SetDefault("logging.separation.error_log", true)
	viper.SetDefault("logging.separation.metrics_log", true)
	
	// Async defaults
	viper.SetDefault("logging.async.enabled", true)
	viper.SetDefault("logging.async.buffer_size", 1000)
	viper.SetDefault("logging.async.flush_interval", "5s")
	viper.SetDefault("logging.async.shutdown_timeout", "10s")
	
	// Metrics defaults
	viper.SetDefault("logging.metrics.enabled", true)
	viper.SetDefault("logging.metrics.include_caller", true)
	viper.SetDefault("logging.metrics.include_request_id", true)
	viper.SetDefault("logging.metrics.include_duration", true)

	// Development overrides
	viper.SetDefault("development.mcp.ssl.enabled", false)
	viper.SetDefault("development.mcp.ssl.auto_generate_certs", true)
	viper.SetDefault("development.api.ssl.insecure_skip_verify", true)
}

// bindEnvVars binds environment variables
func bindEnvVars() {
	viper.SetEnvPrefix("PORTAL64")
	viper.AutomaticEnv()
	
	// API env bindings
	viper.BindEnv("api.base_url", "PORTAL64_API_URL")
	viper.BindEnv("api.timeout", "API_TIMEOUT")
	viper.BindEnv("api.ssl.verify", "API_SSL_VERIFY")
	viper.BindEnv("api.ssl.ca_file", "API_SSL_CA_FILE")
	viper.BindEnv("api.ssl.client_cert", "API_SSL_CLIENT_CERT")
	viper.BindEnv("api.ssl.client_key", "API_SSL_CLIENT_KEY")
	
	// MCP env bindings
	viper.BindEnv("mcp.port", "MCP_SERVER_PORT")
	viper.BindEnv("mcp.mode", "MCP_SERVER_MODE")
	viper.BindEnv("mcp.http_port", "MCP_HTTP_PORT")
	viper.BindEnv("mcp.ssl.enabled", "MCP_SSL_ENABLED")
	viper.BindEnv("mcp.ssl.cert_file", "MCP_SSL_CERT_FILE")
	viper.BindEnv("mcp.ssl.key_file", "MCP_SSL_KEY_FILE")
	
	// Logging env bindings
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")
	viper.BindEnv("logging.file.enabled", "LOG_FILE_ENABLED")
	viper.BindEnv("logging.file.base_path", "LOG_FILE_PATH")
	viper.BindEnv("logging.async.enabled", "LOG_ASYNC_ENABLED")
	viper.BindEnv("logging.metrics.enabled", "LOG_METRICS_ENABLED")
}

// isDevelopment checks if running in development mode
func isDevelopment() bool {
	env := strings.ToLower(os.Getenv("ENV"))
	return env == "" || env == "development" || env == "dev"
}

// applyDevelopmentOverrides applies development-specific configuration
func applyDevelopmentOverrides(config *Config) {
	if config.Development.MCP.SSL.Enabled != config.MCP.SSL.Enabled {
		config.MCP.SSL.Enabled = config.Development.MCP.SSL.Enabled
	}
	if config.Development.API.SSL.InsecureSkipVerify {
		config.API.SSL.InsecureSkipVerify = true
	}
}

// GetTLSConfig returns a TLS configuration for the MCP server
func (m *MCPSSLConfig) GetTLSConfig() (*tls.Config, error) {
	if !m.Enabled {
		return nil, fmt.Errorf("SSL is not enabled")
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Default minimum
		MaxVersion: tls.VersionTLS13, // Default maximum
	}

	// Set TLS version range
	if minVer, err := parseTLSVersion(m.MinVersion); err == nil {
		tlsConfig.MinVersion = minVer
	}
	if maxVer, err := parseTLSVersion(m.MaxVersion); err == nil {
		tlsConfig.MaxVersion = maxVer
	}

	// Set cipher suites if specified
	if len(m.CipherSuites) > 0 {
		suites, err := parseCipherSuites(m.CipherSuites)
		if err != nil {
			return nil, fmt.Errorf("invalid cipher suites: %w", err)
		}
		tlsConfig.CipherSuites = suites
	}

	// Configure client certificate requirements
	if m.RequireClientCert {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	} else if m.CAFile != "" {
		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
	}

	return tlsConfig, nil
}

// parseTLSVersion converts string version to TLS constant
func parseTLSVersion(version string) (uint16, error) {
	switch version {
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %s", version)
	}
}

// parseCipherSuites converts string cipher suite names to constants
func parseCipherSuites(suites []string) ([]uint16, error) {
	// For now, return empty to use Go defaults (secure)
	// This could be expanded to include a mapping of cipher suite names to constants
	return []uint16{}, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.API.Validate(); err != nil {
		return fmt.Errorf("api config: %w", err)
	}
	
	if err := c.MCP.Validate(); err != nil {
		return fmt.Errorf("mcp config: %w", err)
	}

	if c.API.Timeout <= 0 {
		return fmt.Errorf("api.timeout must be positive")
	}

	return nil
}

// Validate validates API configuration
func (a *APIConfig) Validate() error {
	if a.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	
	// Validate SSL client cert pair
	if (a.SSL.ClientCert != "" && a.SSL.ClientKey == "") ||
	   (a.SSL.ClientCert == "" && a.SSL.ClientKey != "") {
		return fmt.Errorf("both client_cert and client_key must be specified together")
	}

	return nil
}

// Validate validates MCP configuration
func (m *MCPConfig) Validate() error {
	if m.Port <= 0 || m.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if m.HTTPPort <= 0 || m.HTTPPort > 65535 {
		return fmt.Errorf("http_port must be between 1 and 65535")
	}

	validModes := map[string]bool{"stdio": true, "http": true, "both": true}
	if !validModes[m.Mode] {
		return fmt.Errorf("mode must be one of: stdio, http, both")
	}

	return m.SSL.Validate()
}

// Validate validates SSL configuration
func (s *MCPSSLConfig) Validate() error {
	if !s.Enabled {
		return nil // No validation needed if SSL is disabled
	}

	// Validate certificate files exist or auto-generation is enabled
	if !s.AutoGenerateCerts {
		if s.CertFile == "" || s.KeyFile == "" {
			return fmt.Errorf("cert_file and key_file are required when SSL is enabled and auto_generate_certs is false")
		}
		
		if _, err := os.Stat(s.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("cert_file does not exist: %s", s.CertFile)
		}
		
		if _, err := os.Stat(s.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("key_file does not exist: %s", s.KeyFile)
		}
	}

	// Validate TLS versions
	if _, err := parseTLSVersion(s.MinVersion); err != nil {
		return fmt.Errorf("invalid min_version: %w", err)
	}
	
	if _, err := parseTLSVersion(s.MaxVersion); err != nil {
		return fmt.Errorf("invalid max_version: %w", err)
	}

	return nil
}
