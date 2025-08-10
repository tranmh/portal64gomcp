package logger

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EnhancedLogger implements the Logger interface with full functionality
type EnhancedLogger struct {
	logrus          *logrus.Logger
	config          *Config
	rotationManager *RotationManager
	asyncWriter     *AsyncWriter
	metricsManager  *MetricsManager
	httpRequestLogger *HTTPRequestLogger
	
	// Runtime state
	isRunning       bool
	shutdownOnce    sync.Once
	
	// Context for request tracking
	baseFields      map[string]interface{}
}

// Factory creates and configures loggers
type Factory struct {
	defaultConfig *Config
	instances     map[string]*EnhancedLogger
	mutex         sync.RWMutex
}

// NewFactory creates a new logger factory
func NewFactory(config *Config) (*Factory, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger configuration: %w", err)
	}
	
	return &Factory{
		defaultConfig: config,
		instances:     make(map[string]*EnhancedLogger),
	}, nil
}

// Create creates a new logger instance
func (f *Factory) Create(name string, config ...*Config) (Logger, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	// Use provided config or default
	var cfg *Config
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = f.defaultConfig
	}
	
	// Check if instance already exists
	if existing, exists := f.instances[name]; exists {
		return existing, nil
	}
	
	// Create new logger instance
	logger, err := f.createLogger(name, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger %s: %w", name, err)
	}
	
	f.instances[name] = logger
	return logger, nil
}

// createLogger creates a new enhanced logger instance
func (f *Factory) createLogger(name string, config *Config) (*EnhancedLogger, error) {
	// Create log directories
	if err := config.CreateLogDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create log directories: %w", err)
	}
	
	// Create base logrus logger
	logrusLogger := logrus.New()
	logrusLogger.SetReportCaller(config.Metrics.IncludeCaller)
	
	// Set log level
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	logrusLogger.SetLevel(level)
	
	// Create rotation manager
	rotationManager := NewRotationManager(config.Rotation)
	
	// Create metrics manager
	metricsManager := NewMetricsManager(config.Metrics.Enabled)
	
	// Create enhanced logger
	enhancedLogger := &EnhancedLogger{
		logrus:          logrusLogger,
		config:          config,
		rotationManager: rotationManager,
		metricsManager:  metricsManager,
		isRunning:       true,
		baseFields: map[string]interface{}{
			"service":  "portal64-mcp",
			"logger":   name,
			"pid":      os.Getpid(),
			"hostname": getHostname(),
		},
	}
	
	// Create HTTP request logger
	enhancedLogger.httpRequestLogger = NewHTTPRequestLogger(enhancedLogger)
	
	// Setup outputs and hooks
	if err := f.setupOutputs(enhancedLogger); err != nil {
		return nil, fmt.Errorf("failed to setup outputs: %w", err)
	}
	
	return enhancedLogger, nil
}

// setupOutputs configures all output destinations and hooks
func (f *Factory) setupOutputs(logger *EnhancedLogger) error {
	config := logger.config
	
	// Create formatter
	formatter := NewEnhancedFormatter(LogFormat(config.Format), config.Metrics.IncludeCaller)
	
	// Setup console output
	if config.Console.Enabled {
		logger.logrus.SetOutput(os.Stdout)
		logger.logrus.SetFormatter(formatter)
	} else {
		// Discard console output if not enabled
		logger.logrus.SetOutput(io.Discard)
	}
	
	// Setup file outputs
	if config.File.Enabled {
		// Create async writer if enabled
		if config.Async.Enabled {
			logger.asyncWriter = NewAsyncWriter(config.Async)
		}
		
		// Add context hook (always first)
		logger.logrus.AddHook(NewContextHook("portal64-mcp", "1.0.0", getEnvironment()))
		
		// Add separated file hooks
		separatedHook := NewSeparatedFileHook(logger.rotationManager, config, formatter)
		
		if config.Async.Enabled {
			// Add separated hook to async writer
			logger.asyncWriter.AddWriter(separatedHook)
			
			// Add async hook to logrus
			logger.logrus.AddHook(NewAsyncHook(logger.asyncWriter, logrus.AllLevels))
			
			// Start async writer
			logger.asyncWriter.Start()
		} else {
			// Add separated hook directly to logrus
			logger.logrus.AddHook(separatedHook)
		}
		
		// Add access log hook if enabled
		if config.Separation.Enabled && config.Separation.AccessLog {
			if accessPath := config.GetLogFilePath("access"); accessPath != "" {
				accessHook := NewAccessLogHook(logger.rotationManager, accessPath)
				
				if config.Async.Enabled {
					logger.asyncWriter.AddWriter(accessHook)
				} else {
					logger.logrus.AddHook(accessHook)
				}
			}
		}
		
		// Add metrics hook if enabled
		if config.Separation.Enabled && config.Separation.MetricsLog {
			if metricsPath := config.GetLogFilePath("metrics"); metricsPath != "" {
				metricsHook := NewMetricsHook(logger.rotationManager, metricsPath)
				
				if config.Async.Enabled {
					logger.asyncWriter.AddWriter(metricsHook)
				} else {
					logger.logrus.AddHook(metricsHook)
				}
			}
		}
	}
	
	return nil
}

// Standard logging methods implementation
func (l *EnhancedLogger) Debug(args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Debug(args...)
	l.recordLogMetrics(logrus.DebugLevel, time.Since(start))
}

func (l *EnhancedLogger) Info(args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Info(args...)
	l.recordLogMetrics(logrus.InfoLevel, time.Since(start))
}

func (l *EnhancedLogger) Warn(args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Warn(args...)
	l.recordLogMetrics(logrus.WarnLevel, time.Since(start))
}

func (l *EnhancedLogger) Error(args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Error(args...)
	l.recordLogMetrics(logrus.ErrorLevel, time.Since(start))
}

func (l *EnhancedLogger) Fatal(args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Fatal(args...)
	l.recordLogMetrics(logrus.FatalLevel, time.Since(start))
}

func (l *EnhancedLogger) Panic(args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Panic(args...)
	l.recordLogMetrics(logrus.PanicLevel, time.Since(start))
}

// Formatted logging methods
func (l *EnhancedLogger) Debugf(format string, args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Debugf(format, args...)
	l.recordLogMetrics(logrus.DebugLevel, time.Since(start))
}

func (l *EnhancedLogger) Infof(format string, args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Infof(format, args...)
	l.recordLogMetrics(logrus.InfoLevel, time.Since(start))
}

func (l *EnhancedLogger) Warnf(format string, args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Warnf(format, args...)
	l.recordLogMetrics(logrus.WarnLevel, time.Since(start))
}

func (l *EnhancedLogger) Errorf(format string, args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Errorf(format, args...)
	l.recordLogMetrics(logrus.ErrorLevel, time.Since(start))
}

func (l *EnhancedLogger) Fatalf(format string, args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Fatalf(format, args...)
	l.recordLogMetrics(logrus.FatalLevel, time.Since(start))
}

func (l *EnhancedLogger) Panicf(format string, args ...interface{}) {
	start := time.Now()
	l.logrus.WithFields(logrus.Fields(l.baseFields)).Panicf(format, args...)
	l.recordLogMetrics(logrus.PanicLevel, time.Since(start))
}

// Field-based logging methods
func (l *EnhancedLogger) WithField(key string, value interface{}) Logger {
	return l.withFields(map[string]interface{}{key: value})
}

func (l *EnhancedLogger) WithFields(fields map[string]interface{}) Logger {
	return l.withFields(fields)
}

func (l *EnhancedLogger) WithError(err error) Logger {
	return l.withFields(map[string]interface{}{"error": err.Error()})
}

func (l *EnhancedLogger) WithContext(ctx context.Context) Logger {
	fields := make(map[string]interface{})
	
	// Extract request ID from context
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	
	// Extract trace ID from context (if using distributed tracing)
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}
	
	return l.withFields(fields)
}

// Enhanced logging methods
func (l *EnhancedLogger) WithRequestID(id string) Logger {
	return l.withFields(map[string]interface{}{"request_id": id})
}

func (l *EnhancedLogger) WithDuration(d time.Duration) Logger {
	return l.withFields(map[string]interface{}{
		"duration":    d,
		"duration_ms": float64(d) / float64(time.Millisecond),
	})
}

func (l *EnhancedLogger) WithComponent(component string) Logger {
	return l.withFields(map[string]interface{}{"component": component})
}

func (l *EnhancedLogger) LogHTTPRequest(req *http.Request, status int, duration time.Duration, bytesWritten int64) {
	requestID := generateRequestID()
	if contextID := req.Context().Value("request_id"); contextID != nil {
		requestID = contextID.(string)
	}
	
	// Record metrics
	l.metricsManager.RecordHTTPRequest(req.Method, req.URL.Path, status, duration, bytesWritten)
	
	// Log request
	l.httpRequestLogger.LogRequest(req, status, duration, bytesWritten, requestID)
}

// Lifecycle methods
func (l *EnhancedLogger) Flush() error {
	if l.asyncWriter != nil {
		l.asyncWriter.Flush()
	}
	
	// Allow some time for async flush to complete
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

func (l *EnhancedLogger) Close() error {
	l.shutdownOnce.Do(func() {
		l.isRunning = false
		
		// Stop async writer
		if l.asyncWriter != nil {
			l.asyncWriter.Stop()
		}
		
		// Close rotation manager
		if l.rotationManager != nil {
			l.rotationManager.Close()
		}
	})
	
	return nil
}

// Helper methods
func (l *EnhancedLogger) withFields(fields map[string]interface{}) Logger {
	// Merge with base fields
	mergedFields := make(map[string]interface{})
	for k, v := range l.baseFields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}
	
	return &fieldLogger{
		enhanced: l,
		entry:    l.logrus.WithFields(logrus.Fields(mergedFields)),
	}
}

func (l *EnhancedLogger) recordLogMetrics(level logrus.Level, writeTime time.Duration) {
	component := ""
	if comp, exists := l.baseFields["component"]; exists {
		component = comp.(string)
	}
	
	l.metricsManager.RecordLogEntry(level, component, writeTime)
}

// fieldLogger wraps a logrus entry with enhanced functionality
type fieldLogger struct {
	enhanced *EnhancedLogger
	entry    *logrus.Entry
}

// Implement Logger interface for fieldLogger
func (fl *fieldLogger) Debug(args ...interface{}) {
	start := time.Now()
	fl.entry.Debug(args...)
	fl.enhanced.recordLogMetrics(logrus.DebugLevel, time.Since(start))
}

func (fl *fieldLogger) Info(args ...interface{}) {
	start := time.Now()
	fl.entry.Info(args...)
	fl.enhanced.recordLogMetrics(logrus.InfoLevel, time.Since(start))
}

func (fl *fieldLogger) Warn(args ...interface{}) {
	start := time.Now()
	fl.entry.Warn(args...)
	fl.enhanced.recordLogMetrics(logrus.WarnLevel, time.Since(start))
}

func (fl *fieldLogger) Error(args ...interface{}) {
	start := time.Now()
	fl.entry.Error(args...)
	fl.enhanced.recordLogMetrics(logrus.ErrorLevel, time.Since(start))
}

func (fl *fieldLogger) Fatal(args ...interface{}) {
	start := time.Now()
	fl.entry.Fatal(args...)
	fl.enhanced.recordLogMetrics(logrus.FatalLevel, time.Since(start))
}

func (fl *fieldLogger) Panic(args ...interface{}) {
	start := time.Now()
	fl.entry.Panic(args...)
	fl.enhanced.recordLogMetrics(logrus.PanicLevel, time.Since(start))
}

func (fl *fieldLogger) Debugf(format string, args ...interface{}) {
	start := time.Now()
	fl.entry.Debugf(format, args...)
	fl.enhanced.recordLogMetrics(logrus.DebugLevel, time.Since(start))
}

func (fl *fieldLogger) Infof(format string, args ...interface{}) {
	start := time.Now()
	fl.entry.Infof(format, args...)
	fl.enhanced.recordLogMetrics(logrus.InfoLevel, time.Since(start))
}

func (fl *fieldLogger) Warnf(format string, args ...interface{}) {
	start := time.Now()
	fl.entry.Warnf(format, args...)
	fl.enhanced.recordLogMetrics(logrus.WarnLevel, time.Since(start))
}

func (fl *fieldLogger) Errorf(format string, args ...interface{}) {
	start := time.Now()
	fl.entry.Errorf(format, args...)
	fl.enhanced.recordLogMetrics(logrus.ErrorLevel, time.Since(start))
}

func (fl *fieldLogger) Fatalf(format string, args ...interface{}) {
	start := time.Now()
	fl.entry.Fatalf(format, args...)
	fl.enhanced.recordLogMetrics(logrus.FatalLevel, time.Since(start))
}

func (fl *fieldLogger) Panicf(format string, args ...interface{}) {
	start := time.Now()
	fl.entry.Panicf(format, args...)
	fl.enhanced.recordLogMetrics(logrus.PanicLevel, time.Since(start))
}

func (fl *fieldLogger) WithField(key string, value interface{}) Logger {
	return &fieldLogger{
		enhanced: fl.enhanced,
		entry:    fl.entry.WithField(key, value),
	}
}

func (fl *fieldLogger) WithFields(fields map[string]interface{}) Logger {
	return &fieldLogger{
		enhanced: fl.enhanced,
		entry:    fl.entry.WithFields(logrus.Fields(fields)),
	}
}

func (fl *fieldLogger) WithError(err error) Logger {
	return fl.WithField("error", err.Error())
}

func (fl *fieldLogger) WithContext(ctx context.Context) Logger {
	fields := make(map[string]interface{})
	
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}
	
	return fl.WithFields(fields)
}

func (fl *fieldLogger) WithRequestID(id string) Logger {
	return fl.WithField("request_id", id)
}

func (fl *fieldLogger) WithDuration(d time.Duration) Logger {
	return fl.WithFields(map[string]interface{}{
		"duration":    d,
		"duration_ms": float64(d) / float64(time.Millisecond),
	})
}

func (fl *fieldLogger) WithComponent(component string) Logger {
	return fl.WithField("component", component)
}

func (fl *fieldLogger) LogHTTPRequest(req *http.Request, status int, duration time.Duration, bytesWritten int64) {
	fl.enhanced.LogHTTPRequest(req, status, duration, bytesWritten)
}

func (fl *fieldLogger) Flush() error {
	return fl.enhanced.Flush()
}

func (fl *fieldLogger) Close() error {
	return fl.enhanced.Close()
}

// Utility functions
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), runtime.NumGoroutine())
}
