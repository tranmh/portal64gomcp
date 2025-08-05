package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/svw-info/portal64gomcp/internal/api"
	"github.com/svw-info/portal64gomcp/internal/config"
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

	// Setup logging
	logger := setupLogger(cfg.Logger)

	logger.WithFields(logrus.Fields{
		"api_url":   cfg.API.BaseURL,
		"timeout":   cfg.API.Timeout,
		"log_level": cfg.Logger.Level,
		"mode":      cfg.MCP.Mode,
		"http_port": cfg.MCP.HTTPPort,
	}).Info("Starting Portal64 MCP Server")

	// Create API client
	apiClient := api.NewClient(cfg.API.BaseURL, cfg.API.Timeout, logger)

	// Create MCP server
	server := mcp.NewServer(cfg, logger, apiClient)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.WithField("signal", sig).Info("Received shutdown signal")
		server.Stop()
	}()

	// Start server
	logger.Info("MCP server starting...")
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Server failed to start")
	}

	logger.Info("MCP server stopped")
}

// setupLogger configures the logger based on configuration
func setupLogger(cfg config.LoggerConfig) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	switch cfg.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return logger
}
