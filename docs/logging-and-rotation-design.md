# Logging and Log Rotation Design

## Overview

This document describes the enhanced logging and log rotation system for the Portal64 MCP Server. The design builds upon the existing logrus implementation to provide comprehensive file-based logging with automatic rotation, compression, and cleanup.

## Design Goals

- **Dual Output**: Support both console and file logging simultaneously
- **Automatic Rotation**: Rotate logs based on size (100MB) OR time (daily)
- **Efficient Storage**: Compress logs after 1 day, retain 30 rotated files
- **Performance**: Async logging to prevent I/O blocking
- **Observability**: Structured metrics for monitoring and debugging
- **Organization**: Separate logs by type (app, access, error)
- **Zero Downtime**: Hot reload configuration without service restart

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Application   │───▶│  Enhanced Logger │───▶│  Multiple       │
│   Components    │    │                  │    │  Outputs        │
│                 │    │  ┌─────────────┐ │    │                 │
│  ├─ MCP Server  │    │  │   Logrus    │ │    │  ├─ Console     │
│  ├─ API Client  │    │  │   Core      │ │    │  ├─ App Logs    │
│  ├─ HTTP Bridge │    │  └─────────────┘ │    │  ├─ Access Logs │
│  └─ SSL Utils   │    │                  │    │  └─ Error Logs  │
└─────────────────┘    │  ┌─────────────┐ │    └─────────────────┘
                       │  │ Async Buffer│ │             │
                       │  │  & Hooks    │ │    ┌─────────────────┐
                       │  └─────────────┘ │    │ Log Management  │
                       └──────────────────┘    │                 │
                                              │ ├─ Lumberjack    │
                                              │ │   Rotation     │
                                              │ ├─ Compression   │
                                              │ └─ Cleanup      │
                                              └─────────────────┘
```

## Log File Organization

```
logs/
├── app/
│   ├── portal64-mcp.log           # Current application log
│   ├── portal64-mcp.log.1         # Yesterday's log
│   ├── portal64-mcp.log.2.gz      # Compressed older log
│   └── ...                        # Up to 30 rotated files
├── access/
│   ├── access.log                 # Current HTTP access log
│   ├── access.log.1.gz            # Compressed access logs
│   └── ...
├── error/
│   ├── error.log                  # Error-level and above
│   ├── error.log.1.gz             # Compressed error logs
│   └── ...
└── metrics/
    ├── metrics.log                # Performance metrics
    └── ...
```

## Configuration Schema

Enhanced `config.yaml` logging configuration:

```yaml
logging:
  # Basic configuration (existing)
  level: "info"                    # debug, info, warn, error, fatal, panic
  format: "json"                   # json, text
  
  # Console output
  console:
    enabled: true                  # Enable console output
    force_colors: false            # Force colored output
    
  # File output
  file:
    enabled: true                  # Enable file logging
    base_path: "logs"              # Base directory for all logs
    
  # Rotation configuration
  rotation:
    max_size: 100                  # MB - rotate when file reaches this size
    max_age: 1                     # days - rotate daily
    max_backups: 30                # number of rotated files to keep
    compress: true                 # compress rotated files
    compress_after: 1              # days - compress files older than this
    
  # Log separation
  separation:
    enabled: true                  # Enable log separation
    access_log: true               # Separate access logs
    error_log: true                # Separate error logs
    metrics_log: true              # Separate metrics logs
    
  # Performance configuration
  async:
    enabled: true                  # Enable async logging
    buffer_size: 1000              # Buffer size for async writes
    flush_interval: "5s"           # Force flush interval
    shutdown_timeout: "10s"        # Graceful shutdown timeout
    
  # Structured metrics
  metrics:
    enabled: true                  # Enable performance metrics
    include_caller: true           # Include file:line in logs
    include_request_id: true       # Include request IDs
    include_duration: true         # Include request duration
```

## Implementation Components

### 1. Enhanced Logger Package (`internal/logger/`)

```
internal/logger/
├── logger.go              # Main logger interface and factory
├── config.go              # Logger configuration
├── hooks.go               # Custom logrus hooks
├── formatters.go          # Custom formatters
├── async.go               # Async logging implementation
├── metrics.go             # Metrics collection
└── rotation.go            # Log rotation management
```

### 2. Core Dependencies

Add to `go.mod`:
```go
require (
    gopkg.in/natefinch/lumberjack.v2 v2.2.1  // Log rotation
    github.com/sirupsen/logrus v1.9.3        // Already present
)
```

### 3. Logger Interface

```go
// Logger interface for dependency injection
type Logger interface {
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})
    
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
    WithError(err error) Logger
    
    // New methods for enhanced functionality
    WithRequestID(id string) Logger
    WithDuration(d time.Duration) Logger
    LogHTTPRequest(req *http.Request, status int, duration time.Duration)
    
    // Lifecycle methods
    Flush() error
    Close() error
}
```

## Key Features

### 1. Dual Output Strategy
- **Console Output**: For development and containerized environments
- **File Output**: For traditional deployments and log aggregation
- **Conditional**: Each output can be independently enabled/disabled

### 2. Smart Rotation
- **Size-based**: Rotate when logs exceed 100MB
- **Time-based**: Rotate daily at midnight
- **Combined**: Whichever condition is met first triggers rotation
- **Atomic**: Rotation is atomic to prevent log loss

### 3. Compression Strategy
- **Immediate**: New rotated files are created uncompressed
- **Delayed**: Files older than 1 day are compressed with gzip
- **Space Efficient**: Reduces storage requirements by 70-80%

### 4. Async Logging
- **Non-blocking**: Application threads don't wait for disk I/O
- **Buffered**: Configurable buffer size (default: 1000 entries)
- **Periodic Flush**: Automatic flush every 5 seconds
- **Graceful Shutdown**: Ensures all logs are written on exit

### 5. Log Separation
- **Application Logs**: General application events and debugging
- **Access Logs**: HTTP request/response logging
- **Error Logs**: Error-level and above for quick incident response
- **Metrics Logs**: Performance metrics and timing data

### 6. Structured Metrics
```json
{
  "timestamp": "2024-08-10T15:30:45.123Z",
  "level": "info",
  "message": "HTTP request completed",
  "request_id": "req_123abc",
  "method": "GET",
  "path": "/api/tournaments",
  "status_code": 200,
  "duration_ms": 45.2,
  "user_agent": "curl/7.68.0",
  "remote_addr": "192.168.1.100",
  "bytes_sent": 1024,
  "component": "http_bridge",
  "caller": "http_bridge.go:112"
}
```

## Integration Points

### 1. HTTP Middleware
```go
// Automatic request logging with timing
func LoggingMiddleware(logger Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            requestID := generateRequestID()
            
            // Add request ID to context
            ctx := context.WithValue(r.Context(), "request_id", requestID)
            r = r.WithContext(ctx)
            
            // Wrap response writer to capture status and size
            wrapped := &responseWriter{ResponseWriter: w}
            
            // Process request
            next.ServeHTTP(wrapped, r)
            
            // Log request completion
            duration := time.Since(start)
            logger.LogHTTPRequest(r, wrapped.status, duration)
        })
    }
}
```

### 2. Error Handling Integration
```go
// Structured error logging with context
func (s *Server) handleError(ctx context.Context, err error, msg string) {
    logger := s.logger.WithError(err)
    
    if requestID := ctx.Value("request_id"); requestID != nil {
        logger = logger.WithField("request_id", requestID.(string))
    }
    
    logger.Error(msg)
}
```

### 3. Performance Monitoring
```go
// Automatic timing for critical operations
func (c *Client) makeRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
    start := time.Now()
    defer func() {
        c.logger.WithDuration(time.Since(start)).
               WithField("operation", "api_request").
               Debug("API request completed")
    }()
    
    return c.httpClient.Do(req)
}
```

## Migration Strategy

### Phase 1: Infrastructure Setup
1. Add lumberjack dependency
2. Create `internal/logger` package
3. Implement basic rotation without breaking existing functionality

### Phase 2: Configuration Enhancement  
1. Extend `LoggerConfig` struct
2. Add new configuration validation
3. Maintain backward compatibility with existing config

### Phase 3: Feature Rollout
1. Enable file logging (keeping console as primary)
2. Add async logging capability
3. Implement log separation

### Phase 4: Optimization
1. Add structured metrics
2. Implement compression
3. Performance tuning and monitoring

## Configuration Examples

### Development Environment
```yaml
logging:
  level: "debug"
  format: "text"
  console:
    enabled: true
    force_colors: true
  file:
    enabled: false         # Disable file logging in development
  async:
    enabled: false         # Synchronous for debugging
```

### Production Environment
```yaml
logging:
  level: "info"
  format: "json"
  console:
    enabled: true          # For container stdout
  file:
    enabled: true          # For persistent storage
  rotation:
    max_size: 100
    max_age: 1
    max_backups: 30
    compress: true
  async:
    enabled: true          # For performance
  separation:
    enabled: true          # For log analysis
  metrics:
    enabled: true          # For monitoring
```

## Monitoring and Alerting

### Log Health Metrics
- Log write failures
- Buffer overflow events
- Rotation failures
- Disk space usage
- Log parsing errors

### Alerting Scenarios
- Log directory full (>90% disk usage)
- Failed log rotations
- High error rates in error.log
- Async buffer overflows
- Missing log entries (gaps in timestamps)

## Performance Characteristics

### Expected Impact
- **CPU Overhead**: <2% additional CPU usage with async logging
- **Memory Usage**: ~10MB for buffers and metadata
- **Disk I/O**: Reduced by 60-80% due to async batching
- **Latency**: <1ms impact on request handling (async mode)

### Benchmarking Results (Projected)
```
Synchronous File Logging:  ~500 logs/sec
Asynchronous File Logging: ~50,000 logs/sec  
Console Only:              ~5,000 logs/sec
Dual Output (Async):       ~25,000 logs/sec
```

## Security Considerations

### Log Content Security
- No automatic PII scrubbing (no compliance requirements)
- Structured format prevents log injection attacks
- Request IDs enable secure log correlation

### File System Security
- Logs written with restricted permissions (0640)
- Log rotation preserves file permissions
- Directory structure prevents unauthorized access

## Operational Procedures

### Log Analysis
```bash
# View recent errors
tail -f logs/error/error.log | jq '.'

# Search for specific request
grep "req_123abc" logs/app/*.log | jq '.'

# Monitor performance metrics
grep "duration_ms" logs/metrics/metrics.log | jq '.duration_ms'
```

### Maintenance Tasks
```bash
# Manual log rotation (if needed)
kill -USR1 $(pidof portal64-mcp)

# Check log disk usage
du -sh logs/

# Verify log compression
ls -la logs/app/*.gz
```

## Future Enhancements

### Potential Additions
- **Log streaming**: Real-time log shipping to external systems
- **Log sampling**: Reduce high-frequency debug logs in production
- **Distributed tracing**: Integration with OpenTelemetry
- **Log aggregation**: Built-in support for ELK/Grafana
- **Hot reload**: Configuration changes without restart

### Integration Opportunities
- **Prometheus**: Export log metrics
- **Grafana**: Log visualization dashboards
- **Alert Manager**: Automated incident response
- **Log aggregators**: Fluentd, Logstash compatibility

---

This design provides a robust, scalable logging solution that enhances observability while maintaining high performance and operational simplicity.
