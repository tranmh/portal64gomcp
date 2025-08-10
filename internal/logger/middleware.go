package logger

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strconv"
	"time"
)

// HTTPMiddleware provides logging middleware for HTTP requests
type HTTPMiddleware struct {
	logger Logger
	config MiddlewareConfig
}

// MiddlewareConfig holds configuration for HTTP middleware
type MiddlewareConfig struct {
	// Request logging
	LogRequests   bool     `json:"log_requests"`
	LogResponses  bool     `json:"log_responses"`
	LogHeaders    bool     `json:"log_headers"`
	LogBody       bool     `json:"log_body"`
	MaxBodySize   int      `json:"max_body_size"`   // Maximum body size to log in bytes
	
	// Header filtering
	SensitiveHeaders []string `json:"sensitive_headers"` // Headers to exclude from logs
	
	// Path filtering
	SkipPaths    []string `json:"skip_paths"`    // Paths to skip logging
	IncludePaths []string `json:"include_paths"` // Only log these paths (if specified)
	
	// Performance
	EnableMetrics bool `json:"enable_metrics"`
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		LogRequests:  true,
		LogResponses: true,
		LogHeaders:   false,
		LogBody:      false,
		MaxBodySize:  1024, // 1KB
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"Set-Cookie", 
			"X-API-Key",
			"X-Auth-Token",
			"Proxy-Authorization",
		},
		SkipPaths: []string{
			"/health",
			"/metrics", 
			"/ping",
			"/favicon.ico",
		},
		EnableMetrics: true,
	}
}

// NewHTTPMiddleware creates a new HTTP middleware
func NewHTTPMiddleware(logger Logger, config ...MiddlewareConfig) *HTTPMiddleware {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	
	return &HTTPMiddleware{
		logger: logger,
		config: cfg,
	}
}

// Handler returns the HTTP middleware handler function
func (m *HTTPMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Generate request ID
		requestID := generateRequestID()
		
		// Add request ID to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)
		
		// Check if we should skip this path
		if m.shouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		
		// Log request if enabled
		if m.config.LogRequests {
			m.logRequest(r, requestID)
		}
		
		// Wrap response writer to capture response data
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default status code
		}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Calculate duration
		duration := time.Since(start)
		
		// Log response if enabled
		if m.config.LogResponses {
			m.logResponse(r, wrapped, duration, requestID)
		}
		
		// Record metrics if enabled
		if m.config.EnableMetrics {
			// This would typically be handled by the logger's metrics manager
			m.logger.LogHTTPRequest(r, wrapped.statusCode, duration, wrapped.bytesWritten)
		}
	})
}

// shouldSkip determines if a path should be skipped from logging
func (m *HTTPMiddleware) shouldSkip(path string) bool {
	// If include paths are specified, only log those
	if len(m.config.IncludePaths) > 0 {
		for _, includePath := range m.config.IncludePaths {
			if path == includePath {
				return false
			}
		}
		return true
	}
	
	// Check skip paths
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	
	return false
}

// logRequest logs the incoming HTTP request
func (m *HTTPMiddleware) logRequest(r *http.Request, requestID string) {
	fields := map[string]interface{}{
		"log_type":    "request",
		"request_id":  requestID,
		"method":      r.Method,
		"path":        r.URL.Path,
		"query":       r.URL.RawQuery,
		"protocol":    r.Proto,
		"host":        r.Host,
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"component":   "http_middleware",
	}
	
	// Add content length if available
	if r.ContentLength > 0 {
		fields["content_length"] = r.ContentLength
	}
	
	// Add referer if present
	if referer := r.Header.Get("Referer"); referer != "" {
		fields["referer"] = referer
	}
	
	// Add headers if enabled
	if m.config.LogHeaders {
		headers := m.filterHeaders(r.Header)
		if len(headers) > 0 {
			fields["headers"] = headers
		}
	}
	
	// Add body if enabled (be careful with large bodies)
	if m.config.LogBody && r.ContentLength > 0 && r.ContentLength <= int64(m.config.MaxBodySize) {
		// Note: This is a simplified example. In practice, you'd need to
		// carefully handle body reading to avoid consuming the request body
		// that the actual handler needs.
		fields["body_logged"] = true
		fields["body_size"] = r.ContentLength
	}
	
	m.logger.WithFields(fields).Info("HTTP request received")
}

// logResponse logs the HTTP response
func (m *HTTPMiddleware) logResponse(r *http.Request, w *responseWriter, duration time.Duration, requestID string) {
	fields := map[string]interface{}{
		"log_type":      "response",
		"request_id":    requestID,
		"method":        r.Method,
		"path":          r.URL.Path,
		"status_code":   w.statusCode,
		"duration":      duration,
		"duration_ms":   float64(duration) / float64(time.Millisecond),
		"bytes_written": w.bytesWritten,
		"component":     "http_middleware",
	}
	
	// Add response headers if enabled
	if m.config.LogHeaders {
		headers := m.filterHeaders(w.Header())
		if len(headers) > 0 {
			fields["response_headers"] = headers
		}
	}
	
	// Choose log level based on status code
	message := "HTTP request completed"
	switch {
	case w.statusCode >= 500:
		m.logger.WithFields(fields).Error(message)
	case w.statusCode >= 400:
		m.logger.WithFields(fields).Warn(message)
	default:
		m.logger.WithFields(fields).Info(message)
	}
}

// filterHeaders removes sensitive headers from logging
func (m *HTTPMiddleware) filterHeaders(headers http.Header) map[string]string {
	filtered := make(map[string]string)
	
	for name, values := range headers {
		// Check if this header should be filtered out
		shouldFilter := false
		for _, sensitive := range m.config.SensitiveHeaders {
			if name == sensitive {
				shouldFilter = true
				break
			}
		}
		
		if !shouldFilter && len(values) > 0 {
			// Join multiple values with comma
			if len(values) == 1 {
				filtered[name] = values[0]
			} else {
				filtered[name] = values[0] + " (+" + strconv.Itoa(len(values)-1) + " more)"
			}
		}
	}
	
	return filtered
}

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	headerWritten bool
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.headerWritten {
		rw.statusCode = statusCode
		rw.headerWritten = true
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body size
func (rw *responseWriter) Write(data []byte) (int, error) {
	// Ensure WriteHeader is called with default status if not already called
	if !rw.headerWritten {
		rw.WriteHeader(200)
	}
	
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += int64(n)
	return n, err
}

// Hijack implements http.Hijacker interface
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

// Flush implements http.Flusher interface
func (rw *responseWriter) Flush() {
	f, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		f.Flush()
	}
}

// Push implements http.Pusher interface
func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	p, ok := rw.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}

// ErrorHandler provides structured error logging for HTTP handlers
type ErrorHandler struct {
	logger Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError logs HTTP errors with context
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	requestID := r.Context().Value("request_id")
	var reqID string
	if requestID != nil {
		reqID = requestID.(string)
	} else {
		reqID = generateRequestID()
	}
	
	fields := map[string]interface{}{
		"log_type":    "error",
		"request_id":  reqID,
		"method":      r.Method,
		"path":        r.URL.Path,
		"status_code": statusCode,
		"error":       err.Error(),
		"component":   "error_handler",
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	}
	
	// Add stack trace for server errors
	if statusCode >= 500 {
		// In a real implementation, you might want to capture
		// and include stack trace information here
		fields["severity"] = "high"
	}
	
	eh.logger.WithFields(fields).WithError(err).Error("HTTP request error")
	
	// Write error response
	http.Error(w, http.StatusText(statusCode), statusCode)
}

// PanicRecoveryHandler provides panic recovery with logging
func (eh *ErrorHandler) PanicRecoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := r.Context().Value("request_id")
				var reqID string
				if requestID != nil {
					reqID = requestID.(string)
				} else {
					reqID = generateRequestID()
				}
				
				fields := map[string]interface{}{
					"log_type":    "panic",
					"request_id":  reqID,
					"method":      r.Method,
					"path":        r.URL.Path,
					"panic":       err,
					"component":   "panic_recovery",
					"severity":    "critical",
				}
				
				eh.logger.WithFields(fields).Error("HTTP handler panic recovered")
				
				// Return 500 Internal Server Error
				if !w.(interface{ HeadersWritten() bool }).HeadersWritten() {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware adds request ID to context if not present
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		w.Header().Set("X-Request-ID", requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HealthCheckMiddleware provides health check endpoint
func HealthCheckMiddleware(logger Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Simple health check response
			response := `{
				"status": "healthy",
				"timestamp": "` + time.Now().Format(time.RFC3339) + `",
				"service": "portal64-mcp"
			}`
			
			w.Write([]byte(response))
			return
		}
		
		http.NotFound(w, r)
	}
}

// MetricsEndpoint provides metrics endpoint
func MetricsEndpoint(metricsManager *MetricsManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !metricsManager.IsEnabled() {
			http.Error(w, "Metrics collection is disabled", http.StatusServiceUnavailable)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// In a real implementation, you'd use json.Marshal here
		// This is a simplified example
		w.Write([]byte(`{"status": "metrics endpoint - implement JSON serialization"}`))
	}
}
