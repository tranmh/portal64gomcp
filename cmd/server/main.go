package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/svw-info/portal64gomcp/internal/api"
	"github.com/svw-info/portal64gomcp/internal/config"
	"github.com/svw-info/portal64gomcp/internal/logger"
	"github.com/svw-info/portal64gomcp/internal/mcp"
)

var (
	configPath = flag.String("config", "", "Path to configuration file")
	logLevel   = flag.String("log-level", "", "Log level (debug, info, warn, error)")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override log level if specified via flag
	if *logLevel != "" {
		cfg.Logger.Level = *logLevel
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup enhanced logging
	loggerConfig := &logger.Config{
		Level:   cfg.Logger.Level,
		Format:  cfg.Logger.Format,
		Console: logger.ConsoleConfig{
			Enabled:     cfg.Logger.Console.Enabled,
			ForceColors: cfg.Logger.Console.ForceColors,
		},
		File: logger.FileConfig{
			Enabled:  cfg.Logger.File.Enabled,
			BasePath: cfg.Logger.File.BasePath,
		},
		Rotation: logger.RotationConfig{
			MaxSize:       cfg.Logger.Rotation.MaxSize,
			MaxAge:        cfg.Logger.Rotation.MaxAge,
			MaxBackups:    cfg.Logger.Rotation.MaxBackups,
			Compress:      cfg.Logger.Rotation.Compress,
			CompressAfter: cfg.Logger.Rotation.CompressAfter,
		},
		Separation: logger.SeparationConfig{
			Enabled:    cfg.Logger.Separation.Enabled,
			AccessLog:  cfg.Logger.Separation.AccessLog,
			ErrorLog:   cfg.Logger.Separation.ErrorLog,
			MetricsLog: cfg.Logger.Separation.MetricsLog,
		},
		Async: logger.AsyncConfig{
			Enabled:         cfg.Logger.Async.Enabled,
			BufferSize:      cfg.Logger.Async.BufferSize,
			FlushInterval:   cfg.Logger.Async.FlushInterval,
			ShutdownTimeout: cfg.Logger.Async.ShutdownTimeout,
		},
		Metrics: logger.MetricsConfig{
			Enabled:          cfg.Logger.Metrics.Enabled,
			IncludeCaller:    cfg.Logger.Metrics.IncludeCaller,
			IncludeRequestID: cfg.Logger.Metrics.IncludeRequestID,
			IncludeDuration:  cfg.Logger.Metrics.IncludeDuration,
		},
	}

	// Create logger factory
	factory, err := logger.NewFactory(loggerConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger factory: %v\n", err)
		os.Exit(1)
	}

	// Create main logger instance
	mainLogger, err := factory.Create("main")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	// Ensure graceful shutdown of logging system
	defer mainLogger.Close()

	mainLogger.WithFields(map[string]interface{}{
		"api_url":        cfg.API.BaseURL,
		"timeout":        cfg.API.Timeout,
		"log_level":      cfg.Logger.Level,
		"mode":           cfg.MCP.Mode,
		"http_port":      cfg.MCP.HTTPPort,
		"ssl_enabled":    cfg.MCP.SSL.Enabled,
		"api_ssl_verify": cfg.API.SSL.Verify,
		"file_logging":   cfg.Logger.File.Enabled,
		"async_logging":  cfg.Logger.Async.Enabled,
	}).Info("Starting Portal64 MCP Server with enhanced logging")

	// Create API client
	apiClient := api.NewClient(cfg.API.BaseURL, cfg.API.Timeout, mainLogger, cfg.API)

	// Create MCP server
	server := mcp.NewServer(cfg, mainLogger, apiClient)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		mainLogger.WithField("signal", sig.String()).Info("Received shutdown signal")
		
		// Flush logs before shutdown
		mainLogger.Flush()
		
		server.Stop()
	}()

	// Start server
	mainLogger.Info("MCP server starting...")
	if err := server.Start(); err != nil {
		mainLogger.WithError(err).Fatal("Server failed to start")
	}

	mainLogger.Info("MCP server stopped")
}
