package logger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger interface for dependency injection and enhanced functionality
type Logger interface {
	// Standard logging methods
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	// Formatted logging methods
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	// Field-based logging
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger

	// Enhanced functionality
	WithRequestID(id string) Logger
	WithDuration(d time.Duration) Logger
	WithComponent(component string) Logger
	LogHTTPRequest(req *http.Request, status int, duration time.Duration, bytesWritten int64)

	// Lifecycle methods
	Flush() error
	Close() error
}

// LogLevel represents logging levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// LogFormat represents logging formats
type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

// LoggerMetrics holds metrics about logger performance
type LoggerMetrics struct {
	TotalLogs       int64 `json:"total_logs"`
	LogsByLevel     map[string]int64 `json:"logs_by_level"`
	ErrorsWritten   int64 `json:"errors_written"`
	BufferOverflows int64 `json:"buffer_overflows"`
	FlushCount      int64 `json:"flush_count"`
	LastFlushTime   time.Time `json:"last_flush_time"`
	BytesWritten    int64 `json:"bytes_written"`
}

// RequestInfo holds HTTP request information for structured logging
type RequestInfo struct {
	RequestID    string        `json:"request_id"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Query        string        `json:"query,omitempty"`
	StatusCode   int           `json:"status_code"`
	Duration     time.Duration `json:"duration"`
	UserAgent    string        `json:"user_agent,omitempty"`
	RemoteAddr   string        `json:"remote_addr"`
	BytesWritten int64         `json:"bytes_written"`
	Component    string        `json:"component"`
}

// parseLogLevel converts string to logrus level
func parseLogLevel(level string) (logrus.Level, error) {
	switch LogLevel(level) {
	case DebugLevel:
		return logrus.DebugLevel, nil
	case InfoLevel:
		return logrus.InfoLevel, nil
	case WarnLevel:
		return logrus.WarnLevel, nil
	case ErrorLevel:
		return logrus.ErrorLevel, nil
	case FatalLevel:
		return logrus.FatalLevel, nil
	case PanicLevel:
		return logrus.PanicLevel, nil
	default:
		return logrus.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}
