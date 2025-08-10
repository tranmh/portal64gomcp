package logger

import (
	"os"
	"testing"
	"time"
)

func TestNewFactory(t *testing.T) {
	// Test with default config
	factory, err := NewFactory(nil)
	if err != nil {
		t.Fatalf("Failed to create factory with default config: %v", err)
	}
	if factory == nil {
		t.Fatal("Factory is nil")
	}
	
	// Test with custom config
	config := &Config{
		Level:  "debug",
		Format: "json",
		Console: ConsoleConfig{
			Enabled: true,
		},
		File: FileConfig{
			Enabled:  false, // Disable file logging for tests
			BasePath: "test-logs",
		},
		Async: AsyncConfig{
			Enabled: false, // Disable async for simpler testing
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}
	
	factory2, err := NewFactory(config)
	if err != nil {
		t.Fatalf("Failed to create factory with custom config: %v", err)
	}
	if factory2 == nil {
		t.Fatal("Factory2 is nil")
	}
}

func TestCreateLogger(t *testing.T) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Console: ConsoleConfig{
			Enabled: true,
		},
		File: FileConfig{
			Enabled:  false, // Disable file logging for tests
			BasePath: "test-logs",
		},
		Async: AsyncConfig{
			Enabled: false, // Disable async for simpler testing
		},
		Metrics: MetricsConfig{
			Enabled: false, // Disable metrics for simpler testing
		},
	}
	
	factory, err := NewFactory(config)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}
	
	logger, err := factory.Create("test-logger")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger is nil")
	}
	
	// Test basic logging methods
	logger.Info("Test info message")
	logger.Debug("Test debug message")
	logger.Warn("Test warn message")
	logger.Error("Test error message")
	
	// Test field logging
	logger.WithField("key", "value").Info("Test field message")
	logger.WithFields(map[string]interface{}{
		"field1": "value1",
		"field2": 42,
	}).Info("Test fields message")
	
	// Test request ID
	logger.WithRequestID("req-123").Info("Test request ID")
	
	// Test duration
	logger.WithDuration(100 * time.Millisecond).Info("Test duration")
	
	// Test component
	logger.WithComponent("test-component").Info("Test component")
	
	// Close logger
	err = logger.Close()
	if err != nil {
		t.Errorf("Failed to close logger: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	// Test valid config
	validConfig := DefaultConfig()
	err := validConfig.Validate()
	if err != nil {
		t.Errorf("Valid config should not have errors: %v", err)
	}
	
	// Test invalid log level
	invalidConfig := DefaultConfig()
	invalidConfig.Level = "invalid"
	err = invalidConfig.Validate()
	if err == nil {
		t.Error("Invalid log level should cause validation error")
	}
	
	// Test invalid format
	invalidConfig2 := DefaultConfig()
	invalidConfig2.Format = "invalid"
	err = invalidConfig2.Validate()
	if err == nil {
		t.Error("Invalid format should cause validation error")
	}
	
	// Test no outputs enabled
	invalidConfig3 := DefaultConfig()
	invalidConfig3.Console.Enabled = false
	invalidConfig3.File.Enabled = false
	err = invalidConfig3.Validate()
	if err == nil {
		t.Error("No outputs enabled should cause validation error")
	}
}

func TestFileLogWithRotation(t *testing.T) {
	// Create temp directory for test logs
	tempDir := "test-logs-temp"
	defer os.RemoveAll(tempDir)
	
	config := &Config{
		Level:  "info",
		Format: "json",
		Console: ConsoleConfig{
			Enabled: false, // Only file logging
		},
		File: FileConfig{
			Enabled:  true,
			BasePath: tempDir,
		},
		Rotation: RotationConfig{
			MaxSize:    1, // 1MB for testing
			MaxAge:     1,
			MaxBackups: 5,
			Compress:   false, // Don't compress for easier testing
		},
		Async: AsyncConfig{
			Enabled: false, // Sync for predictable testing
		},
		Metrics: MetricsConfig{
			Enabled: false,
		},
	}
	
	factory, err := NewFactory(config)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}
	
	logger, err := factory.Create("test-file-logger")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	// Write some log entries
	for i := 0; i < 100; i++ {
		logger.WithFields(map[string]interface{}{
			"iteration": i,
			"test":      "file-rotation",
		}).Info("Test log entry for rotation")
	}
	
	// Flush to ensure files are written
	err = logger.Flush()
	if err != nil {
		t.Errorf("Failed to flush logger: %v", err)
	}
	
	// Check that log files exist
	appLogPath := config.GetLogFilePath("app")
	if _, err := os.Stat(appLogPath); os.IsNotExist(err) {
		t.Errorf("App log file should exist: %s", appLogPath)
	}
}

func TestAsyncLogging(t *testing.T) {
	tempDir := "test-logs-async"
	defer os.RemoveAll(tempDir)
	
	config := &Config{
		Level:  "info",
		Format: "json",
		Console: ConsoleConfig{
			Enabled: false,
		},
		File: FileConfig{
			Enabled:  true,
			BasePath: tempDir,
		},
		Async: AsyncConfig{
			Enabled:         true,
			BufferSize:      100,
			FlushInterval:   100 * time.Millisecond,
			ShutdownTimeout: 1 * time.Second,
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}
	
	factory, err := NewFactory(config)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}
	
	logger, err := factory.Create("test-async-logger")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	// Write logs rapidly to test async behavior
	for i := 0; i < 1000; i++ {
		logger.WithFields(map[string]interface{}{
			"iteration": i,
			"test":      "async-logging",
		}).Info("Async test log entry")
	}
	
	// Let async writer process entries
	time.Sleep(200 * time.Millisecond)
	
	// Flush and close
	err = logger.Flush()
	if err != nil {
		t.Errorf("Failed to flush async logger: %v", err)
	}
}

func TestMetricsCollection(t *testing.T) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Console: ConsoleConfig{
			Enabled: true,
		},
		File: FileConfig{
			Enabled: false,
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}
	
	factory, err := NewFactory(config)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}
	
	logger, err := factory.Create("test-metrics-logger")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	// Write various log levels to test metrics
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")
	logger.Debug("Debug message")
	
	// Note: In a real test, you would access the metrics manager
	// and verify the counts, but since it's embedded in the logger
	// implementation, this test mainly verifies no crashes occur
}

// Benchmark tests
func BenchmarkSyncLogging(b *testing.B) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Console: ConsoleConfig{
			Enabled: false,
		},
		File: FileConfig{
			Enabled:  true,
			BasePath: "bench-logs",
		},
		Async: AsyncConfig{
			Enabled: false,
		},
	}
	
	factory, _ := NewFactory(config)
	logger, _ := factory.Create("bench-sync")
	defer logger.Close()
	defer os.RemoveAll("bench-logs")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark log message")
	}
}

func BenchmarkAsyncLogging(b *testing.B) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Console: ConsoleConfig{
			Enabled: false,
		},
		File: FileConfig{
			Enabled:  true,
			BasePath: "bench-logs-async",
		},
		Async: AsyncConfig{
			Enabled:    true,
			BufferSize: 1000,
		},
	}
	
	factory, _ := NewFactory(config)
	logger, _ := factory.Create("bench-async")
	defer logger.Close()
	defer os.RemoveAll("bench-logs-async")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark log message")
	}
}
