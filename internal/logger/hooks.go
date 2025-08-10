package logger

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// EnhancedFormatter adds enhanced formatting with caller info and request context
type EnhancedFormatter struct {
	baseFormatter   logrus.Formatter
	includeCaller   bool
	includeHostname bool
	hostname        string
}

// NewEnhancedFormatter creates a new enhanced formatter
func NewEnhancedFormatter(format LogFormat, includeCaller bool) *EnhancedFormatter {
	var baseFormatter logrus.Formatter
	
	switch format {
	case JSONFormat:
		baseFormatter = &logrus.JSONFormatter{
			TimestampFormat:   "2006-01-02T15:04:05.000Z07:00",
			DisableHTMLEscape: true,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		}
	default:
		baseFormatter = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		}
	}
	
	hostname, _ := os.Hostname()
	
	return &EnhancedFormatter{
		baseFormatter:   baseFormatter,
		includeCaller:   includeCaller,
		includeHostname: true,
		hostname:        hostname,
	}
}

// Format formats the log entry with enhanced information
func (f *EnhancedFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Clone entry data to avoid modifying original
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	
	// Add hostname
	if f.includeHostname && f.hostname != "" {
		data["hostname"] = f.hostname
	}
	
	// Add caller information
	if f.includeCaller && entry.Caller != nil {
		data["caller"] = fmt.Sprintf("%s:%d", 
			strings.TrimPrefix(entry.Caller.File, runtime.GOROOT()), 
			entry.Caller.Line)
		data["function"] = entry.Caller.Function
	}
	
	// Create new entry with enhanced data
	enhancedEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    data,
		Time:    entry.Time,
		Level:   entry.Level,
		Caller:  entry.Caller,
		Message: entry.Message,
	}
	
	return f.baseFormatter.Format(enhancedEntry)
}

// AccessLogHook handles HTTP access logging
type AccessLogHook struct {
	writer    *RotationManager
	logPath   string
	formatter logrus.Formatter
}

// NewAccessLogHook creates a new access log hook
func NewAccessLogHook(rotationManager *RotationManager, logPath string) *AccessLogHook {
	formatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	}
	
	return &AccessLogHook{
		writer:    rotationManager,
		logPath:   logPath,
		formatter: formatter,
	}
}

// Levels returns info level only for access logs
func (hook *AccessLogHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel}
}

// Fire writes access log entries
func (hook *AccessLogHook) Fire(entry *logrus.Entry) error {
	// Only log entries marked as access logs
	if entry.Data["log_type"] != "access" {
		return nil
	}
	
	writer := hook.writer.GetWriter(hook.logPath)
	msg, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}
	
	_, err = writer.Write(msg)
	return err
}

// MetricsHook handles performance metrics logging
type MetricsHook struct {
	writer     *RotationManager
	logPath    string
	formatter  logrus.Formatter
	metrics    *MetricsCollector
}

// MetricsCollector collects and aggregates performance metrics
type MetricsCollector struct {
	httpRequestCount    int64
	httpRequestDuration int64 // Total duration in nanoseconds
	httpStatusCounts    map[int]int64
	errorCount          int64
	lastReset           time.Time
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook(rotationManager *RotationManager, logPath string) *MetricsHook {
	formatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level", 
			logrus.FieldKeyMsg:   "message",
		},
	}
	
	return &MetricsHook{
		writer:    rotationManager,
		logPath:   logPath,
		formatter: formatter,
		metrics: &MetricsCollector{
			httpStatusCounts: make(map[int]int64),
			lastReset:        time.Now(),
		},
	}
}

// Levels returns all levels for metrics collection
func (hook *MetricsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes log entries for metrics collection
func (hook *MetricsHook) Fire(entry *logrus.Entry) error {
	// Collect metrics from log entries
	hook.collectMetrics(entry)
	
	// Only write explicit metrics entries to metrics log
	if entry.Data["log_type"] == "metrics" {
		writer := hook.writer.GetWriter(hook.logPath)
		msg, err := hook.formatter.Format(entry)
		if err != nil {
			return err
		}
		
		_, err = writer.Write(msg)
		return err
	}
	
	return nil
}

// collectMetrics extracts and aggregates metrics from log entries
func (hook *MetricsHook) collectMetrics(entry *logrus.Entry) {
	// Count HTTP requests
	if statusCode, ok := entry.Data["status_code"]; ok {
		atomic.AddInt64(&hook.metrics.httpRequestCount, 1)
		
		if code, ok := statusCode.(int); ok {
			hook.metrics.httpStatusCounts[code]++
		}
		
		// Aggregate duration
		if duration, ok := entry.Data["duration"]; ok {
			if d, ok := duration.(time.Duration); ok {
				atomic.AddInt64(&hook.metrics.httpRequestDuration, int64(d))
			}
		}
	}
	
	// Count errors
	if entry.Level >= logrus.ErrorLevel {
		atomic.AddInt64(&hook.metrics.errorCount, 1)
	}
}

// GetMetrics returns current collected metrics
func (hook *MetricsHook) GetMetrics() map[string]interface{} {
	requestCount := atomic.LoadInt64(&hook.metrics.httpRequestCount)
	totalDuration := atomic.LoadInt64(&hook.metrics.httpRequestDuration)
	errorCount := atomic.LoadInt64(&hook.metrics.errorCount)
	
	var avgDuration float64
	if requestCount > 0 {
		avgDuration = float64(totalDuration) / float64(requestCount) / float64(time.Millisecond)
	}
	
	return map[string]interface{}{
		"http_requests_total":     requestCount,
		"http_request_duration_avg_ms": avgDuration,
		"http_status_counts":      hook.metrics.httpStatusCounts,
		"errors_total":            errorCount,
		"collection_period":       time.Since(hook.metrics.lastReset).String(),
	}
}

// ResetMetrics resets collected metrics
func (hook *MetricsHook) ResetMetrics() {
	atomic.StoreInt64(&hook.metrics.httpRequestCount, 0)
	atomic.StoreInt64(&hook.metrics.httpRequestDuration, 0)
	atomic.StoreInt64(&hook.metrics.errorCount, 0)
	hook.metrics.httpStatusCounts = make(map[int]int64)
	hook.metrics.lastReset = time.Now()
}

// ContextHook adds context information to log entries
type ContextHook struct {
	serviceName string
	version     string
	environment string
}

// NewContextHook creates a context hook with service information
func NewContextHook(serviceName, version, environment string) *ContextHook {
	return &ContextHook{
		serviceName: serviceName,
		version:     version,
		environment: environment,
	}
}

// Levels returns all levels
func (hook *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire adds context information to log entries
func (hook *ContextHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = hook.serviceName
	entry.Data["version"] = hook.version
	entry.Data["environment"] = hook.environment
	return nil
}

// FilterHook filters log entries based on conditions
type FilterHook struct {
	filters []LogFilter
}

// LogFilter represents a condition for filtering logs
type LogFilter struct {
	Field     string
	Value     interface{}
	Condition FilterCondition
	Action    FilterAction
}

// FilterCondition represents filter condition types
type FilterCondition int

const (
	FilterEquals FilterCondition = iota
	FilterNotEquals
	FilterContains
	FilterNotContains
	FilterGreaterThan
	FilterLessThan
)

// FilterAction represents actions to take when filter matches
type FilterAction int

const (
	FilterSkip FilterAction = iota
	FilterModify
	FilterRedirect
)

// NewFilterHook creates a new filter hook
func NewFilterHook(filters []LogFilter) *FilterHook {
	return &FilterHook{
		filters: filters,
	}
}

// Levels returns all levels
func (hook *FilterHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire applies filters to log entries
func (hook *FilterHook) Fire(entry *logrus.Entry) error {
	for _, filter := range hook.filters {
		if hook.matchesFilter(entry, filter) {
			switch filter.Action {
			case FilterSkip:
				// Mark entry for skipping (handled by caller)
				entry.Data["_skip"] = true
				return nil
			case FilterModify:
				// Modify entry data
				entry.Data["_filtered"] = true
			case FilterRedirect:
				// Mark for different output (handled by caller)
				entry.Data["_redirect"] = filter.Value
			}
		}
	}
	
	return nil
}

// matchesFilter checks if an entry matches a filter condition
func (hook *FilterHook) matchesFilter(entry *logrus.Entry, filter LogFilter) bool {
	var fieldValue interface{}
	
	// Get field value
	switch filter.Field {
	case "level":
		fieldValue = entry.Level.String()
	case "message":
		fieldValue = entry.Message
	default:
		fieldValue = entry.Data[filter.Field]
	}
	
	// Apply condition
	switch filter.Condition {
	case FilterEquals:
		return fieldValue == filter.Value
	case FilterNotEquals:
		return fieldValue != filter.Value
	case FilterContains:
		if str, ok := fieldValue.(string); ok {
			if filterStr, ok := filter.Value.(string); ok {
				return strings.Contains(str, filterStr)
			}
		}
		return false
	case FilterNotContains:
		if str, ok := fieldValue.(string); ok {
			if filterStr, ok := filter.Value.(string); ok {
				return !strings.Contains(str, filterStr)
			}
		}
		return true
	default:
		return false
	}
}

// HTTPRequestLogger provides structured HTTP request logging
type HTTPRequestLogger struct {
	logger Logger
}

// NewHTTPRequestLogger creates a new HTTP request logger
func NewHTTPRequestLogger(logger Logger) *HTTPRequestLogger {
	return &HTTPRequestLogger{
		logger: logger,
	}
}

// LogRequest logs an HTTP request with structured data
func (hrl *HTTPRequestLogger) LogRequest(req *http.Request, status int, duration time.Duration, bytesWritten int64, requestID string) {
	fields := map[string]interface{}{
		"log_type":      "access",
		"request_id":    requestID,
		"method":        req.Method,
		"path":          req.URL.Path,
		"query":         req.URL.RawQuery,
		"status_code":   status,
		"duration":      duration,
		"duration_ms":   float64(duration) / float64(time.Millisecond),
		"user_agent":    req.UserAgent(),
		"remote_addr":   req.RemoteAddr,
		"bytes_written": bytesWritten,
		"component":     "http_server",
		"protocol":      req.Proto,
		"host":          req.Host,
	}
	
	// Add headers if needed (be careful with sensitive data)
	if referer := req.Header.Get("Referer"); referer != "" {
		fields["referer"] = referer
	}
	
	// Add request body size if available
	if req.ContentLength > 0 {
		fields["content_length"] = req.ContentLength
	}
	
	// Choose log level based on status code
	var level string
	switch {
	case status >= 500:
		level = "error"
	case status >= 400:
		level = "warn"
	default:
		level = "info"
	}
	
	message := fmt.Sprintf("%s %s %d", req.Method, req.URL.Path, status)
	
	switch level {
	case "error":
		hrl.logger.WithFields(fields).Error(message)
	case "warn":
		hrl.logger.WithFields(fields).Warn(message)
	default:
		hrl.logger.WithFields(fields).Info(message)
	}
}

// LogError logs HTTP errors with additional context
func (hrl *HTTPRequestLogger) LogError(req *http.Request, err error, requestID string) {
	fields := map[string]interface{}{
		"log_type":   "error",
		"request_id": requestID,
		"method":     req.Method,
		"path":       req.URL.Path,
		"component":  "http_server",
		"error":      err.Error(),
	}
	
	hrl.logger.WithFields(fields).WithError(err).Error("HTTP request error")
}
