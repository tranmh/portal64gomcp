package logger

import (
	"fmt"
	"os"
	"time"
)

// Config holds enhanced logging configuration
type Config struct {
	Level   string        `mapstructure:"level" json:"level"`
	Format  string        `mapstructure:"format" json:"format"`
	Console ConsoleConfig `mapstructure:"console" json:"console"`
	File    FileConfig    `mapstructure:"file" json:"file"`
	Rotation RotationConfig `mapstructure:"rotation" json:"rotation"`
	Separation SeparationConfig `mapstructure:"separation" json:"separation"`
	Async   AsyncConfig   `mapstructure:"async" json:"async"`
	Metrics MetricsConfig `mapstructure:"metrics" json:"metrics"`
}

// ConsoleConfig holds console output configuration
type ConsoleConfig struct {
	Enabled     bool `mapstructure:"enabled" json:"enabled"`
	ForceColors bool `mapstructure:"force_colors" json:"force_colors"`
}

// FileConfig holds file output configuration
type FileConfig struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled"`
	BasePath string `mapstructure:"base_path" json:"base_path"`
}

// RotationConfig holds log rotation configuration
type RotationConfig struct {
	MaxSize       int  `mapstructure:"max_size" json:"max_size"`        // MB
	MaxAge        int  `mapstructure:"max_age" json:"max_age"`         // days
	MaxBackups    int  `mapstructure:"max_backups" json:"max_backups"`     // number of files
	Compress      bool `mapstructure:"compress" json:"compress"`        // compress old files
	CompressAfter int  `mapstructure:"compress_after" json:"compress_after"`  // days before compression
}

// SeparationConfig holds log separation configuration
type SeparationConfig struct {
	Enabled    bool `mapstructure:"enabled" json:"enabled"`
	AccessLog  bool `mapstructure:"access_log" json:"access_log"`
	ErrorLog   bool `mapstructure:"error_log" json:"error_log"`
	MetricsLog bool `mapstructure:"metrics_log" json:"metrics_log"`
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
	Enabled           bool `mapstructure:"enabled" json:"enabled"`
	IncludeCaller     bool `mapstructure:"include_caller" json:"include_caller"`
	IncludeRequestID  bool `mapstructure:"include_request_id" json:"include_request_id"`
	IncludeDuration   bool `mapstructure:"include_duration" json:"include_duration"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  "info",
		Format: "json",
		Console: ConsoleConfig{
			Enabled:     true,
			ForceColors: false,
		},
		File: FileConfig{
			Enabled:  true,
			BasePath: "logs",
		},
		Rotation: RotationConfig{
			MaxSize:       100,  // 100MB
			MaxAge:        1,    // 1 day
			MaxBackups:    30,   // 30 files
			Compress:      true,
			CompressAfter: 1,    // compress after 1 day
		},
		Separation: SeparationConfig{
			Enabled:    true,
			AccessLog:  true,
			ErrorLog:   true,
			MetricsLog: true,
		},
		Async: AsyncConfig{
			Enabled:         true,
			BufferSize:      1000,
			FlushInterval:   5 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Metrics: MetricsConfig{
			Enabled:          true,
			IncludeCaller:    true,
			IncludeRequestID: true,
			IncludeDuration:  true,
		},
	}
}

// Validate validates the logging configuration
func (c *Config) Validate() error {
	// Validate log level
	if _, err := parseLogLevel(c.Level); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Validate log format
	if c.Format != "json" && c.Format != "text" {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'text')", c.Format)
	}

	// Validate file configuration
	if c.File.Enabled {
		if c.File.BasePath == "" {
			return fmt.Errorf("file.base_path is required when file logging is enabled")
		}
		
		// Create base directory if it doesn't exist
		if err := os.MkdirAll(c.File.BasePath, 0755); err != nil {
			return fmt.Errorf("failed to create log directory %s: %w", c.File.BasePath, err)
		}
	}

	// Validate rotation configuration
	if c.Rotation.MaxSize <= 0 {
		return fmt.Errorf("rotation.max_size must be positive")
	}
	if c.Rotation.MaxAge <= 0 {
		return fmt.Errorf("rotation.max_age must be positive")
	}
	if c.Rotation.MaxBackups < 0 {
		return fmt.Errorf("rotation.max_backups cannot be negative")
	}
	if c.Rotation.CompressAfter < 0 {
		return fmt.Errorf("rotation.compress_after cannot be negative")
	}

	// Validate async configuration
	if c.Async.Enabled {
		if c.Async.BufferSize <= 0 {
			return fmt.Errorf("async.buffer_size must be positive")
		}
		if c.Async.FlushInterval <= 0 {
			return fmt.Errorf("async.flush_interval must be positive")
		}
		if c.Async.ShutdownTimeout <= 0 {
			return fmt.Errorf("async.shutdown_timeout must be positive")
		}
	}

	// Validate that at least one output is enabled
	if !c.Console.Enabled && !c.File.Enabled {
		return fmt.Errorf("at least one output (console or file) must be enabled")
	}

	return nil
}

// GetLogFilePath returns the path for a specific log type
func (c *Config) GetLogFilePath(logType string) string {
	if !c.File.Enabled {
		return ""
	}

	switch logType {
	case "app":
		return fmt.Sprintf("%s/app/portal64-mcp.log", c.File.BasePath)
	case "access":
		if c.Separation.Enabled && c.Separation.AccessLog {
			return fmt.Sprintf("%s/access/access.log", c.File.BasePath)
		}
		return ""
	case "error":
		if c.Separation.Enabled && c.Separation.ErrorLog {
			return fmt.Sprintf("%s/error/error.log", c.File.BasePath)
		}
		return ""
	case "metrics":
		if c.Separation.Enabled && c.Separation.MetricsLog {
			return fmt.Sprintf("%s/metrics/metrics.log", c.File.BasePath)
		}
		return ""
	default:
		return fmt.Sprintf("%s/app/portal64-mcp.log", c.File.BasePath)
	}
}

// CreateLogDirectories creates all necessary log directories
func (c *Config) CreateLogDirectories() error {
	if !c.File.Enabled {
		return nil
	}

	directories := []string{
		fmt.Sprintf("%s/app", c.File.BasePath),
	}

	if c.Separation.Enabled {
		if c.Separation.AccessLog {
			directories = append(directories, fmt.Sprintf("%s/access", c.File.BasePath))
		}
		if c.Separation.ErrorLog {
			directories = append(directories, fmt.Sprintf("%s/error", c.File.BasePath))
		}
		if c.Separation.MetricsLog {
			directories = append(directories, fmt.Sprintf("%s/metrics", c.File.BasePath))
		}
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
